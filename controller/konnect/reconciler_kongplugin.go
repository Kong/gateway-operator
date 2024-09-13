package konnect

import (
	"context"
	"strings"

	"github.com/samber/lo"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/kong/gateway-operator/controller/pkg/log"
	"github.com/kong/gateway-operator/pkg/consts"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

// KongPluginReconciler reconciles a KongPlugin object.
type KongPluginReconciler struct {
	developmentMode bool
	client          client.Client
}

// NewKongPluginReconciler creates a new KongPluginReconciler.
func NewKongPluginReconciler(
	developmentMode bool,
	client client.Client,
) *KongPluginReconciler {
	return &KongPluginReconciler{
		developmentMode: developmentMode,
		client:          client,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KongPluginReconciler) SetupWithManager(_ context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("KongPlugin").
		For(&configurationv1.KongPlugin{}).
		Watches(
			&configurationv1alpha1.KongPluginBinding{},
			handler.EnqueueRequestsFromMapFunc(r.mapKongPluginBindings),
		).
		Watches(
			&configurationv1alpha1.KongService{},
			handler.EnqueueRequestsFromMapFunc(r.mapKongServices),
		).
		Complete(r)
}

// Reconcile reconciles a KongPlugin object.
// The purpose of this reconciler is to handle annotations on Kong entities objects that reference KongPlugin objects.
// As a result of such annotations, KongPluginBinding objects are created and managed by the controller.
func (r *KongPluginReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var (
		entityTypeName = "KongPlugin"
		logger         = log.GetLogger(ctx, entityTypeName, r.developmentMode)
	)

	// Fetch the KongPlugin instance
	var kongPlugin configurationv1.KongPlugin
	if err := r.client.Get(ctx, req.NamespacedName, &kongPlugin); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Debug(logger, "reconciling", kongPlugin)

	// Get the pluginBindings that use this KongPlugin
	kongPluginBindingList := configurationv1alpha1.KongPluginBindingList{}
	err := r.client.List(
		ctx,
		&kongPluginBindingList,
		client.MatchingFields{
			IndexFieldKongPluginBindingKongPluginReference: kongPlugin.Namespace + "/" + kongPlugin.Name,
		},
		client.InNamespace(kongPlugin.Namespace),
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	// TODO(mlavacca): So far we are supporting only KongService targets here. We need to implement
	// the same logic for KongRoute, KongConsumer, and KongConsumerGroup as well.
	// https://github.com/Kong/gateway-operator/issues/583

	// Group the PluginBindings by KongService name
	pluginBindingsByServiceName := map[string][]configurationv1alpha1.KongPluginBinding{}
	for _, pluginBinding := range kongPluginBindingList.Items {
		if pluginBinding.Spec.Targets.ServiceReference == nil ||
			pluginBinding.Spec.Targets.ServiceReference.Group != configurationv1alpha1.GroupVersion.Group ||
			pluginBinding.Spec.Targets.ServiceReference.Kind != "KongService" {
			continue
		}
		pluginBindingsByServiceName[pluginBinding.Spec.Targets.ServiceReference.Name] = append(pluginBindingsByServiceName[pluginBinding.Spec.Targets.ServiceReference.Name], pluginBinding)
	}

	// Get all the KongServices referenced by the KongPluginBindings
	// TODO(mlavacca): use indexers instead of listing all KongServices
	// https://github.com/Kong/gateway-operator/issues/596
	kongServiceList := configurationv1alpha1.KongServiceList{}
	err = r.client.List(
		ctx,
		&kongServiceList,
		client.InNamespace(kongPlugin.Namespace),
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	pluginBindingsToDelete := []configurationv1alpha1.KongPluginBinding{}
	// pluginReferenceFound represents whether the plugin is referenced by any KongService
	var pluginReferenceFound bool
	for _, kongService := range kongServiceList.Items {
		var pluginSlice []string

		// in case no kongpluginbindings are referencing the kongservice, but it has the finalizer,
		// we need to remove the finalizer.
		if len(pluginBindingsByServiceName[kongService.Name]) == 0 {
			if controllerutil.RemoveFinalizer(&kongService, consts.CleanupPluginBindingFinalizer) {
				if err = r.client.Update(ctx, &kongService); err != nil {
					if k8serrors.IsConflict(err) {
						return ctrl.Result{Requeue: true}, nil
					}
					return ctrl.Result{}, err
				}
				log.Debug(logger, "KongService finalizer removed", kongService)
				return ctrl.Result{}, nil
			}
		}

		// if the KongService is marked for deletion, we need to delete all the PluginBindings that reference it.
		if !kongService.DeletionTimestamp.IsZero() {
			for _, pb := range pluginBindingsByServiceName[kongService.Name] {
				if err := r.client.Delete(ctx, &pb); err != nil {
					return ctrl.Result{}, err
				}
				log.Debug(logger, "KongPluginBinding deleted", pb)
			}
			// Once all the KongPluginBindings that use the KongService have been deleted, remove the finalizer.
			controllerutil.RemoveFinalizer(&kongService, consts.CleanupPluginBindingFinalizer)
			if err = r.client.Update(ctx, &kongService); err != nil {
				return ctrl.Result{}, err
			}
			log.Debug(logger, "KongService finalizer removed", kongService)
			return ctrl.Result{}, nil
		}

		// get the referenced plugins from the KongService annotations
		plugins, ok := kongService.Annotations[consts.PluginsAnnotationKey]
		if !ok {
			// if the konghq.com/plugins annotation is not present, we need to delete all the managed
			// KongPluginBindings that reference the KongService
			for _, pb := range pluginBindingsByServiceName[kongService.Name] {
				if lo.ContainsBy(pb.OwnerReferences, func(ownerRef metav1.OwnerReference) bool {
					if ownerRef.Kind == "KongPlugin" && ownerRef.Name == kongPlugin.Name && ownerRef.UID == kongPlugin.UID {
						return true
					}
					return false
				}) {
					// The PluginBinding is dangling, so it needs to be deleted
					pluginBindingsToDelete = append(pluginBindingsToDelete, pb)
				} else {
					pluginReferenceFound = true
				}
			}
		} else {
			pluginSlice = strings.Split(plugins, ",")

			for _, pb := range kongPluginBindingList.Items {
				// if the kongPluginBinding targets the KongService,
				if pb.Spec.Targets.ServiceReference != nil &&
					pb.Spec.Targets.ServiceReference.Name == kongService.Name &&
					// but the service does not contain the plugin referenced by the binding in the annotation, and
					!lo.Contains(pluginSlice, pb.Spec.PluginReference.Name) &&
					// the plugin is managed (created out of an annotation)
					lo.ContainsBy(pb.OwnerReferences, func(ownerRef metav1.OwnerReference) bool {
						if ownerRef.Kind == "KongPlugin" && ownerRef.Name == kongPlugin.Name && ownerRef.UID == kongPlugin.UID {
							return true
						}
						return false
					}) {
					// then mark it for deletion, as the plugin is not referenced by the KongService anymore
					pluginBindingsToDelete = append(pluginBindingsToDelete, pb)
				}
			}

			// iterate over all the KongService annotations
			for _, pluginName := range pluginSlice {
				if pluginName != kongPlugin.Name {
					continue
				}

				pluginReferenceFound = true
				// if the KongPluginBinding does not exist yet, create it
				if len(pluginBindingsByServiceName[kongService.Name]) == 0 {
					kongPluginBinding := configurationv1alpha1.KongPluginBinding{
						ObjectMeta: metav1.ObjectMeta{
							GenerateName: kongPlugin.Name + "-",
							Namespace:    kongPlugin.Namespace,
						},
						Spec: configurationv1alpha1.KongPluginBindingSpec{
							Targets: configurationv1alpha1.KongPluginBindingTargets{
								ServiceReference: &configurationv1alpha1.TargetRefWithGroupKind{
									Group: configurationv1alpha1.GroupVersion.Group,
									Kind:  "KongService",
									Name:  kongService.Name,
								},
							},
							// TODO: Cross check this with other types of ControlPlaneRefs
							// used by Route, Consumer and/or ConsumerGroups that also bind this plugin
							// in this KongPluginBinding spec.
							ControlPlaneRef: kongService.Spec.ControlPlaneRef,
							PluginReference: configurationv1alpha1.PluginRef{
								Name: kongPlugin.Name,
							},
						},
					}
					k8sutils.SetOwnerForObject(&kongPluginBinding, &kongPlugin)
					if err = r.client.Create(ctx, &kongPluginBinding); err != nil {
						return ctrl.Result{}, err
					}
					log.Debug(logger, "Managed KongPluginBinding created", kongPluginBinding)
				}
				// add a finalizer to the KongService that means the associated kongPluginBindings need to be cleaned up
				if controllerutil.AddFinalizer(&kongService, consts.CleanupPluginBindingFinalizer) {
					if err = r.client.Update(ctx, &kongService); err != nil {
						return ctrl.Result{}, err
					}
					log.Debug(logger, "KongService finalizer added", kongService)
					return ctrl.Result{}, nil
				}
			}
		}

		// iterate over all the KongPluginBindings to be deleted and delete them.
		for _, pb := range pluginBindingsToDelete {
			if err = r.client.Delete(ctx, &pb); err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return ctrl.Result{}, err
			}
			log.Info(logger, "KongPluginBinding deleted", pb)
			return ctrl.Result{}, nil
		}
	}

	// If some KongService is using the plugin, add a finalizer on the plugin.
	// The KongPlugin cannot be deleted until all objects that reference it are
	// deleted or do not reference it anymore.
	if pluginReferenceFound {
		if controllerutil.AddFinalizer(&kongPlugin, consts.PluginInUseFinalizer) {
			if err = r.client.Update(ctx, &kongPlugin); err != nil {
				return ctrl.Result{}, err
			}
			log.Debug(logger, "KongPlugin finalizer added", kongPlugin)
			return ctrl.Result{}, nil
		}
	} else {
		if controllerutil.RemoveFinalizer(&kongPlugin, consts.PluginInUseFinalizer) {
			if err = r.client.Update(ctx, &kongPlugin); err != nil {
				return ctrl.Result{}, err
			}
			log.Debug(logger, "KongPlugin finalizer removed", kongPlugin)
			return ctrl.Result{}, nil
		}
	}

	log.Debug(logger, "reconciliation completed", kongPlugin)
	return ctrl.Result{}, nil
}
