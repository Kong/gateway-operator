package konnect

import (
	"context"
	"errors"
	"strings"

	"github.com/samber/lo"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kong/gateway-operator/controller/pkg/log"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

type KongPluginReconciler struct {
	DevelopmentMode bool
	Client          client.Client
}

func NewKongPluginReconciler(
	developmentMode bool,
	client client.Client,
) *KongPluginReconciler {
	return &KongPluginReconciler{
		DevelopmentMode: developmentMode,
		Client:          client,
	}
}

func (r *KongPluginReconciler) SetupWithManager(mgr ctrl.Manager) error {
	b := ctrl.NewControllerManagedBy(mgr).
		For(&configurationv1.KongPlugin{}).
		Watches(&configurationv1alpha1.KongPluginBinding{}, handler.EnqueueRequestsFromMapFunc(r.mapKongPlugins)).
		Watches(&configurationv1alpha1.KongService{}, handler.EnqueueRequestsFromMapFunc(r.mapKongServices)).
		Named("KongPlugin")

	return b.Complete(r)
}

func (r *KongPluginReconciler) mapKongPlugins(ctx context.Context, obj client.Object) []reconcile.Request {
	logger := log.GetLogger(ctx, "KongPlugin", r.DevelopmentMode)
	kongPlugin, ok := obj.(*configurationv1alpha1.KongPluginBinding)
	if !ok {
		log.Error(logger, errors.New("cannot cast object to KongPluginBinding"), "KongPluginBinding mapping handler", obj)
		return []ctrl.Request{}
	}

	return []ctrl.Request{
		{
			NamespacedName: types.NamespacedName{
				Namespace: kongPlugin.Namespace,
				Name:      kongPlugin.Spec.PluginReference.Name,
			},
		},
	}
}

func (r *KongPluginReconciler) mapKongServices(ctx context.Context, obj client.Object) []reconcile.Request {
	logger := log.GetLogger(ctx, "KongPlugin", r.DevelopmentMode)
	kongService, ok := obj.(*configurationv1alpha1.KongService)
	if !ok {
		log.Error(logger, errors.New("cannot cast object to KongService"), "KongService mapping handler", obj)
		return []ctrl.Request{}
	}

	requests := []ctrl.Request{}
	if plugins, ok := kongService.Annotations["konghq.com/plugins"]; ok {
		for _, p := range strings.Split(plugins, ",") {
			requests = append(requests, ctrl.Request{
				NamespacedName: client.ObjectKey{
					Namespace: kongService.Namespace,
					Name:      p,
				},
			})
		}
	}
	return requests
}

// TODO(mlavacca): watch for KongService and KongPluginBinding
func (r *KongPluginReconciler) Reconcile(
	ctx context.Context, req ctrl.Request,
) (ctrl.Result, error) {
	var (
		entityTypeName = "KongPlugin"
		logger         = log.GetLogger(ctx, entityTypeName, r.DevelopmentMode)
	)

	var plugin configurationv1.KongPlugin
	if err := r.Client.Get(ctx, req.NamespacedName, &plugin); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	log.Debug(logger, "reconciling", plugin)

	pluginBindings := []configurationv1alpha1.KongPluginBinding{}
	referencingBindingList := configurationv1alpha1.KongPluginBindingList{}
	err := r.Client.List(ctx, &referencingBindingList)
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, pluginBinding := range referencingBindingList.Items {
		if pluginBinding.Spec.PluginReference.Name == plugin.Name {
			pluginBindings = append(pluginBindings, pluginBinding)
		}
	}

	pluginBindingsByServiceName := map[string][]configurationv1alpha1.KongPluginBinding{}
	pluginBindingsToDelete := []configurationv1alpha1.KongPluginBinding{}

	for _, pluginBinding := range pluginBindings {
		pluginBindingsByServiceName[pluginBinding.Spec.Kong.ServiceReference.Name] = append(pluginBindingsByServiceName[pluginBinding.Spec.Kong.ServiceReference.Name], pluginBinding)
	}

	kongServiceList := configurationv1alpha1.KongServiceList{}
	err = r.Client.List(ctx, &kongServiceList)
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, kongService := range kongServiceList.Items {
		// TODO(mlavacca): filter services via a better informer way
		var (
			kongServiceFound bool
			pluginSlice      []string
		)
		plugins, ok := kongService.Annotations["konghq.com/plugins"]
		if !ok {
			for _, pb := range pluginBindingsByServiceName[kongService.Name] {
				if lo.ContainsBy(pb.OwnerReferences, func(ownerRef metav1.OwnerReference) bool {
					if ownerRef.Kind == "KongPlugin" && ownerRef.Name == plugin.Name && ownerRef.UID == plugin.UID {
						return true
					}
					return false
				}) {
					pluginBindingsToDelete = append(pluginBindingsToDelete, pb)
				}
			}
		} else {
			pluginSlice = strings.Split(plugins, ",")

			for _, pb := range pluginBindings {
				if pb.Spec.Kong.ServiceReference.Name == kongService.Name &&
					!lo.Contains(pluginSlice, pb.Spec.PluginReference.Name) &&
					lo.ContainsBy(pb.OwnerReferences, func(ownerRef metav1.OwnerReference) bool {
						if ownerRef.Kind == "KongPlugin" && ownerRef.Name == plugin.Name && ownerRef.UID == plugin.UID {
							return true
						}
						return false
					}) {
					pluginBindingsToDelete = append(pluginBindingsToDelete, pb)
				}
			}

			for _, p := range pluginSlice {
				if p == plugin.Name {
					kongServiceFound = true
				}
			}
		}
		for _, pb := range pluginBindingsToDelete {
			if err = r.Client.Delete(ctx, &pb); err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return ctrl.Result{}, err
			}
			// in case a pb has been deleted, let's return and let the deletion trigger a new reconciliation loop.
			log.Info(logger, "deleted KongPluginBinding", pb)
			return ctrl.Result{}, nil
		}

		if !kongServiceFound {
			continue
		}

		if len(pluginBindingsByServiceName[kongService.Name]) == 0 {
			kongPluginBinding := configurationv1alpha1.KongPluginBinding{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: plugin.Name + "-",
					Namespace:    plugin.Namespace,
				},
				Spec: configurationv1alpha1.KongPluginBindingSpec{
					Kong: &configurationv1alpha1.KongReferences{
						ServiceReference: &configurationv1alpha1.EntityRef{
							Name: kongService.Name,
						},
					},
					PluginReference: configurationv1alpha1.PluginRef{
						Name: plugin.Name,
					},
				},
			}
			k8sutils.SetOwnerForObject(&kongPluginBinding, &plugin)
			if err = r.Client.Create(ctx, &kongPluginBinding); err != nil {
				return ctrl.Result{}, err
			}
			// in case the KongPluginBinding was created, we can return.
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}
