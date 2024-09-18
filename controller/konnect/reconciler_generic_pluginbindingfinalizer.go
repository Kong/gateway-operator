package konnect

import (
	"context"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kong/gateway-operator/controller/konnect/constraints"
	"github.com/kong/gateway-operator/controller/pkg/log"
	"github.com/kong/gateway-operator/pkg/consts"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

// KonnectEntityPluginBindingFinalizerReconciler reconciles a Konnect entities.
// It uses the generic type constraints to constrain the supported types.
type KonnectEntityPluginBindingFinalizerReconciler[
	T constraints.SupportedKonnectEntityPluginReferenceableType,
	TEnt constraints.EntityType[T],
] struct {
	DevelopmentMode bool
	Client          client.Client
}

// NewKonnectEntityPluginReconciler returns a new KonnectEntityPluginReconciler
// for the given Konnect entity type.
func NewKonnectEntityPluginReconciler[
	T constraints.SupportedKonnectEntityPluginReferenceableType,
	TEnt constraints.EntityType[T],
](
	developmentMode bool,
	client client.Client,
) *KonnectEntityPluginBindingFinalizerReconciler[T, TEnt] {
	r := &KonnectEntityPluginBindingFinalizerReconciler[T, TEnt]{
		DevelopmentMode: developmentMode,
		Client:          client,
	}
	return r
}

// SetupWithManager sets up the controller with the given manager.
func (r *KonnectEntityPluginBindingFinalizerReconciler[T, TEnt]) SetupWithManager(
	ctx context.Context, mgr ctrl.Manager,
) error {
	var (
		entityTypeName = constraints.EntityTypeName[T]()
		b              = ctrl.NewControllerManagedBy(mgr).Named(entityTypeName + "PluginBindingCleanupFinalizer")
	)

	r.setControllerBuilderOptionsForKongPluginBinding(b)

	return b.Complete(r)
}

func enqueueKongServiceForKongPluginBinding() func(
	ctx context.Context, obj client.Object) []reconcile.Request {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		kpb, ok := obj.(*configurationv1alpha1.KongPluginBinding)
		if !ok {
			return nil
		}

		if kpb.Spec.Targets.ServiceReference == nil ||
			kpb.Spec.Targets.ServiceReference.Kind != "KongService" ||
			kpb.Spec.Targets.ServiceReference.Group != configurationv1alpha1.GroupVersion.Group {
			return nil
		}

		return []ctrl.Request{
			{
				NamespacedName: types.NamespacedName{
					Namespace: kpb.Namespace,
					Name:      kpb.Spec.Targets.ServiceReference.Name,
				},
			},
		}
	}
}

// Reconcile reconciles the Konnect entity that can be set as KongPluginBinding target.
// Its purpose is to:
//   - check if the entity is marked for deletion and mark KongPluginBindings
//     that reference it.
//   - add a finalizer to the entity if there are KongPluginBindings referencing it.
//     This finalizer designates that this entity needs to have its KongPluginBindings
//     removed upon deletion
//   - remove the finalizer if all KongPluginBindings referencing it are removed.
func (r *KonnectEntityPluginBindingFinalizerReconciler[T, TEnt]) Reconcile(
	ctx context.Context, req ctrl.Request,
) (ctrl.Result, error) {
	var (
		entityTypeName = constraints.EntityTypeName[T]()
		logger         = log.GetLogger(ctx, entityTypeName, r.DevelopmentMode)
	)

	var (
		e   T
		ent = TEnt(&e)
	)
	if err := r.Client.Get(ctx, req.NamespacedName, ent); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	ctx = ctrllog.IntoContext(ctx, logger)
	log.Debug(logger, "reconciling", ent)

	cl := client.NewNamespacedClient(r.Client, ent.GetNamespace())
	kongPluginBindingList := configurationv1alpha1.KongPluginBindingList{}
	err := cl.List(
		ctx,
		&kongPluginBindingList,
		client.MatchingFields{
			getIndexField(ent): ent.GetName(),
		},
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	var finalizersChangedAction string
	// if the entity is marked for deletion, we need to delete all the PluginBindings that reference it.
	if !ent.GetDeletionTimestamp().IsZero() {
		for _, kpb := range kongPluginBindingList.Items {
			if err := cl.Delete(ctx, &kpb); err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return ctrl.Result{}, err
			}
			log.Debug(logger, "KongPluginBinding deleted", kpb)
		}
		// in case no KongPluginBindings are referencing the entity, but it has the finalizer,
		// we need to remove the finalizer.
		if controllerutil.RemoveFinalizer(ent, consts.CleanupPluginBindingFinalizer) {
			finalizersChangedAction = "removed"
		}
	} else {
		// if the entity is not marked for deletion, update the finalizers accordingly.
		switch len(kongPluginBindingList.Items) {
		case 0:
			// in case no KongPluginBindings are referencing the entity, but it has the finalizer,
			// we need to remove the finalizer.
			if controllerutil.RemoveFinalizer(ent, consts.CleanupPluginBindingFinalizer) {
				finalizersChangedAction = "removed"
			}

		default:
			// add a finalizer to the entity that means the associated
			// kongPluginBindings need to be cleaned up upon deletion.
			if controllerutil.AddFinalizer(ent, consts.CleanupPluginBindingFinalizer) {
				finalizersChangedAction = "added"
			}
		}
	}

	if finalizersChangedAction != "" {
		if err = r.Client.Update(ctx, ent); err != nil {
			if k8serrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}
		log.Debug(logger, "finalizers changed", ent,
			"action", finalizersChangedAction,
			"finalizer", consts.CleanupPluginBindingFinalizer,
		)
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func getIndexField[
	T constraints.SupportedKonnectEntityPluginReferenceableType,
	TEnt constraints.EntityType[T],
](ent TEnt) string {
	switch any(ent).(type) {
	case *configurationv1alpha1.KongService:
		return IndexFieldKongPluginBindingKongServiceReference
	default:
		panic(fmt.Sprintf("unsupported entity type %s", constraints.EntityTypeName[T]()))
	}
}

func (r *KonnectEntityPluginBindingFinalizerReconciler[T, TEnt]) setControllerBuilderOptionsForKongPluginBinding(
	b *builder.TypedBuilder[ctrl.Request],
) {
	kongPluginsAnnotationChangedPredicate := predicate.Funcs{
		CreateFunc: func(e event.TypedCreateEvent[client.Object]) bool {
			_, ok := e.Object.GetAnnotations()[consts.PluginsAnnotationKey]
			return ok
		},
		UpdateFunc: func(e event.TypedUpdateEvent[client.Object]) bool {
			if e.ObjectOld == nil || e.ObjectNew == nil {
				return false
			}
			return e.ObjectNew.GetAnnotations()[consts.PluginsAnnotationKey] !=
				e.ObjectOld.GetAnnotations()[consts.PluginsAnnotationKey]
		},
		DeleteFunc: func(e event.TypedDeleteEvent[client.Object]) bool {
			_, ok := e.Object.GetAnnotations()[consts.PluginsAnnotationKey]
			return ok
		},
	}

	var (
		e   T
		ent = TEnt(&e)
	)

	switch any(ent).(type) {
	case *configurationv1alpha1.KongService:
		b.
			For(&configurationv1alpha1.KongService{},
				builder.WithPredicates(
					predicate.NewPredicateFuncs(kongServiceRefersToKonnectGatewayControlPlane),
					kongPluginsAnnotationChangedPredicate,
				),
			).
			Watches(
				&configurationv1alpha1.KongPluginBinding{},
				handler.EnqueueRequestsFromMapFunc(
					enqueueKongServiceForKongPluginBinding(),
				),
			)
	default:
		panic(fmt.Sprintf("unsupported entity type %s", constraints.EntityTypeName[T]()))
	}
}