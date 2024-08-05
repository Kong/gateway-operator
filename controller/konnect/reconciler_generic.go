package konnect

import (
	"context"
	"fmt"
	"time"

	sdkkonnectgo "github.com/Kong/sdk-konnect-go"
	sdkkonnectgocomp "github.com/Kong/sdk-konnect-go/models/components"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kong/gateway-operator/controller/pkg/log"
	"github.com/kong/gateway-operator/pkg/consts"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

const (
	// TODO(pmalek) make configurable
	configurableSyncPeriod = 1 * time.Minute
)

const (
	// KonnectCleanupFinalizer is the finalizer that is added to the Konnect
	// entities when they are created in Konnect, and which is removed when
	// the CR and Konnect entity are deleted.
	KonnectCleanupFinalizer = "gateway.konghq.com/konnect-cleanup"
)

type KonnectEntityReconciler[T SupportedKonnectEntityType, TEnt EntityType[T]] struct {
	DevelopmentMode bool
	Client          client.Client
}

func NewKonnectEntityReconciler[
	T SupportedKonnectEntityType,
	TEnt EntityType[T],
](
	t T,
	developmentMode bool,
	client client.Client,
) *KonnectEntityReconciler[T, TEnt] {
	return &KonnectEntityReconciler[T, TEnt]{
		DevelopmentMode: developmentMode,
		Client:          client,
	}
}

func (r *KonnectEntityReconciler[T, TEnt]) SetupWithManager(mgr ctrl.Manager) error {
	var (
		e   T
		ent = TEnt(&e)
		b   = ctrl.NewControllerManagedBy(mgr).
			For(ent).
			Named(entityTypeName[T]()).
			WithOptions(controller.Options{
				// TODO: investigate
				MaxConcurrentReconciles: 8,
				// TODO: investigate NewQueue
			})
	)

	for _, dep := range ReconciliationWatchOptionsForEntity(r.Client, ent) {
		b = dep(b)
	}
	return b.Complete(r)
}

func (r *KonnectEntityReconciler[T, TEnt]) Reconcile(
	ctx context.Context, req ctrl.Request,
) (ctrl.Result, error) {
	var (
		entityTypeName = entityTypeName[T]()
		logger         = log.GetLogger(ctx, entityTypeName, r.DevelopmentMode)
	)

	var (
		e   T
		ent = TEnt(&e)
	)
	if err := r.Client.Get(ctx, req.NamespacedName, ent); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	log.Debug(logger, "reconciling", ent)
	var (
		apiAuth    konnectv1alpha1.KonnectAPIAuthConfiguration
		apiAuthRef = types.NamespacedName{
			Name:      ent.GetKonnectAPIAuthConfigurationRef().Name,
			Namespace: ent.GetNamespace(),
			// TODO(pmalek): enable if cross namespace refs are allowed
			// Namespace: ent.GetKonnectAPIAuthConfigurationRef().Namespace,
		}
	)
	// if apiAuthRef.Namespace == "" {
	// 	apiAuthRef.Namespace = ent.GetNamespace()
	// }
	if err := r.Client.Get(ctx, apiAuthRef, &apiAuth); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get KonnectAPIAuthConfiguration: %w", err)
	}

	if cond, present := k8sutils.GetCondition(KonnectEntityAPIAuthConfigurationValidConditionType, &apiAuth); present && cond.Status != metav1.ConditionTrue {
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				KonnectEntityAPIAuthConfigurationResolvedRefConditionType,
				metav1.ConditionFalse,
				KonnectEntityAPIAuthConfigurationReasonInvalid,
				"",
				ent.GetGeneration(),
			),
			ent,
		)

		return ctrl.Result{}, nil
	}

	k8sutils.SetCondition(
		k8sutils.NewConditionWithGeneration(
			KonnectEntityAPIAuthConfigurationResolvedRefConditionType,
			metav1.ConditionTrue,
			KonnectEntityAPIAuthConfigurationReasonValid,
			fmt.Sprintf("referenced KonnectAPIAuthConfiguration %s is valid", apiAuthRef),
			ent.GetGeneration(),
		),
		ent,
	)
	if err := r.Client.Status().Update(ctx, ent); err != nil {
		if k8serrors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to update status with APIAuthRefValid condition: %w", err)
	}

	// NOTE: /organizations/me is not public in OpenAPI spec so we can use it
	// but not using the SDK
	// https://kongstrong.slack.com/archives/C04RXLGNB6K/p1719830395775599?thread_ts=1719406468.883729&cid=C04RXLGNB6K
	sdk := sdkkonnectgo.New(
		sdkkonnectgo.WithSecurity(
			sdkkonnectgocomp.Security{
				PersonalAccessToken: sdkkonnectgo.String(apiAuth.Spec.Token),
			},
		),
		sdkkonnectgo.WithServerURL("https://"+apiAuth.Spec.ServerURL),
	)

	if !ent.GetDeletionTimestamp().IsZero() {
		logger.Info("resource is being deleted")
		// wait for termination grace period before cleaning up
		if ent.GetDeletionTimestamp().After(time.Now()) {
			logger.Info("resource still under grace period, requeueing")
			return ctrl.Result{
				// Requeue when grace period expires.
				// If deletion timestamp is changed,
				// the update will trigger another round of reconciliation.
				// so we do not consider updates of deletion timestamp here.
				RequeueAfter: time.Until(ent.GetDeletionTimestamp().Time),
			}, nil
		}

		if controllerutil.RemoveFinalizer(ent, KonnectCleanupFinalizer) {
			if err := Delete[T, TEnt](ctx, sdk, logger, r.Client, ent); err != nil {
				return ctrl.Result{}, err
			}
			if err := r.Client.Update(ctx, ent); err != nil {
				if k8serrors.IsConflict(err) {
					return ctrl.Result{Requeue: true}, nil
				}
				return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
			}
		}

		return ctrl.Result{}, nil
	}

	if typeHasControlPlaneRef(ent) {
		cpRef := getControlPlaneRef(ent)
		switch cpRef.Type {
		case configurationv1alpha1.ControlPlaneRefKonnectNamespacedRef:
			cp := konnectv1alpha1.KonnectControlPlane{}
			nn := types.NamespacedName{
				Name:      cpRef.KonnectNamespacedRef.Name,
				Namespace: cpRef.KonnectNamespacedRef.Namespace,
			}
			if nn.Namespace == "" {
				nn.Namespace = ent.GetNamespace()
			}
			if err := r.Client.Get(ctx, nn, &cp); err != nil {
				k8sutils.SetCondition(
					k8sutils.NewConditionWithGeneration(
						ControlPlaneRefValidConditionType,
						metav1.ConditionFalse,
						ControlPlaneRefReasonInvalid,
						err.Error(),
						ent.GetGeneration(),
					),
					ent,
				)
				if err := r.Client.Status().Update(ctx, ent); err != nil {
					if k8serrors.IsConflict(err) {
						return ctrl.Result{Requeue: true}, nil
					}
					return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
				}

				return ctrl.Result{}, fmt.Errorf("ControlPlane %s doesn't exist", nn)
			}

			cond, ok := k8sutils.GetCondition(KonnectEntityProgrammedConditionType, &cp)
			if !ok || cond.Status != metav1.ConditionTrue /*|| cond.ObservedGeneration != cp.GetGeneration() */ {
				ent.GetKonnectStatus().SetKonnectID("")
				k8sutils.SetCondition(
					k8sutils.NewConditionWithGeneration(
						ControlPlaneRefValidConditionType,
						metav1.ConditionFalse,
						ControlPlaneRefReasonInvalid,
						fmt.Sprintf("Referenced ControlPlane %s is not programmed yet", nn),
						ent.GetGeneration(),
					),
					ent,
				)
				if err := r.Client.Status().Update(ctx, ent); err != nil {
					if k8serrors.IsConflict(err) {
						return ctrl.Result{Requeue: true}, nil
					}
					return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
				}
				return ctrl.Result{Requeue: true}, nil
			}

			old := ent.DeepCopyObject().(TEnt)
			if err := controllerutil.SetOwnerReference(&cp, ent, r.Client.Scheme(), controllerutil.WithBlockOwnerDeletion(true)); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to set owner reference: %w", err)
			}
			if err := r.Client.Patch(ctx, ent, client.MergeFrom(old)); err != nil {
				if k8serrors.IsConflict(err) {
					return ctrl.Result{Requeue: true}, nil
				}
				return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
			}

			// TODO(pmalek): make this generic.
			// CP ID is not stored in KonnectEntityStatus because not all entities
			// have a ControlPlaneRef, hence the type constraints in the reconciler can't be used.
			switch resource := any(ent).(type) {
			case *configurationv1alpha1.KongService:
				resource.Status.Konnect.ControlPlaneID = cp.Status.ID
			case *configurationv1alpha1.KongRoute:
				resource.Status.Konnect.ControlPlaneID = cp.Status.ID
			case *configurationv1.KongConsumer:
				resource.Status.Konnect.ControlPlaneID = cp.Status.ID
			}

			k8sutils.SetCondition(
				k8sutils.NewConditionWithGeneration(
					ControlPlaneRefValidConditionType,
					metav1.ConditionTrue,
					ControlPlaneRefReasonValid,
					fmt.Sprintf("Referenced ControlPlane %s programmed", nn),
					ent.GetGeneration(),
				),
				ent,
			)
			if err := r.Client.Status().Patch(ctx, ent, client.MergeFrom(old)); err != nil {
				if k8serrors.IsConflict(err) {
					return ctrl.Result{Requeue: true}, nil
				}
				return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
			}

		default:
			return ctrl.Result{}, fmt.Errorf("unimplemented control plane ref type %q", cpRef.Type)
		}

		// TODO: handle control plane ref
	}

	// TODO this is a bit of a mess, we should refactor this
	if typeHasServiceRef(ent) {
		ref := getServiceRef(ent)
		switch ref.Type {
		case configurationv1alpha1.ServiceRefNamespacedRef:
			svc := configurationv1alpha1.KongService{}
			nn := types.NamespacedName{
				Name:      ref.NamespacedRef.Name,
				Namespace: ref.NamespacedRef.Namespace,
			}
			if nn.Namespace == "" {
				nn.Namespace = ent.GetNamespace()
			}
			if err := r.Client.Get(ctx, nn, &svc); err != nil {
				k8sutils.SetCondition(
					k8sutils.NewConditionWithGeneration(
						ServiceRefValidConditionType,
						metav1.ConditionFalse,
						ServiceRefReasonInvalid,
						err.Error(),
						ent.GetGeneration(),
					),
					ent,
				)
				if err := r.Client.Status().Update(ctx, ent); err != nil {
					if k8serrors.IsConflict(err) {
						return ctrl.Result{Requeue: true}, nil
					}
					return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
				}

				return ctrl.Result{}, fmt.Errorf("Can't get the referenced KongService %s: %w", nn, err)
			}

			cond, ok := k8sutils.GetCondition(KonnectEntityProgrammedConditionType, &svc)
			if !ok || cond.Status != metav1.ConditionTrue /*|| cond.ObservedGeneration != cp.GetGeneration() */ {
				ent.GetKonnectStatus().SetKonnectID("")
				k8sutils.SetCondition(
					k8sutils.NewConditionWithGeneration(
						ServiceRefValidConditionType,
						metav1.ConditionFalse,
						ServiceRefReasonInvalid,
						fmt.Sprintf("Referenced KongService %s is not programmed yet", nn),
						ent.GetGeneration(),
					),
					ent,
				)
				if err := r.Client.Status().Update(ctx, ent); err != nil {
					if k8serrors.IsConflict(err) {
						return ctrl.Result{Requeue: true}, nil
					}
					return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
				}
				return ctrl.Result{Requeue: true}, nil
			}

			old := ent.DeepCopyObject().(TEnt)
			if err := controllerutil.SetOwnerReference(&svc, ent, r.Client.Scheme(), controllerutil.WithBlockOwnerDeletion(true)); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to set owner reference: %w", err)
			}
			if err := r.Client.Patch(ctx, ent, client.MergeFrom(old)); err != nil {
				if k8serrors.IsConflict(err) {
					return ctrl.Result{Requeue: true}, nil
				}
				return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
			}

			// TODO(pmalek): make this generic.
			// Service ID is not stored in KonnectEntityStatus because not all entities
			// have a ServiceRef, hence the type constraints in the reconciler can't be used.
			if route, ok := any(ent).(*configurationv1alpha1.KongRoute); ok {
				route.Status.Konnect.ServiceID = svc.Status.Konnect.GetKonnectID()
			}

			k8sutils.SetCondition(
				k8sutils.NewConditionWithGeneration(
					ServiceRefValidConditionType,
					metav1.ConditionTrue,
					ServiceRefReasonValid,
					fmt.Sprintf("Referenced KongService %s programmed", nn),
					ent.GetGeneration(),
				),
				ent,
			)
			if err := r.Client.Status().Patch(ctx, ent, client.MergeFrom(old)); err != nil {
				if k8serrors.IsConflict(err) {
					return ctrl.Result{Requeue: true}, nil
				}
				return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
			}

		default:
			return ctrl.Result{}, fmt.Errorf("unimplemented KongService ref type %q", ref.Type)
		}

		// TODO: handle control plane ref
	}

	// TODO: relying on status ID is OK but we need to rethink this because
	// we're using a cached client so after creating the resource on Konnect it might
	// happen that we've just created the resource but the status ID is not there yet.
	//
	// We should look at the "expectations" for this:
	// https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/controller_utils.go
	if id := ent.GetKonnectStatus().GetKonnectID(); id == "" {
		_, err := Create[T, TEnt](ctx, sdk, logger, r.Client, ent)
		if err != nil {
			// TODO(pmalek): this is actually not 100% error prone because when status
			// update fails we don't store the Konnect ID and hence the reconciler
			// will try to create the resource again on next reconciliation.
			if err := r.Client.Status().Update(ctx, ent); err != nil {
				if k8serrors.IsConflict(err) {
					return ctrl.Result{Requeue: true}, nil
				}
				return ctrl.Result{}, fmt.Errorf("failed to update status after creating object: %w", err)
			}

			return ctrl.Result{}, FailedKonnectOpError[T]{
				Op:  CreateOp,
				Err: err,
			}
		}

		ent.GetKonnectStatus().ServerURL = apiAuth.Spec.ServerURL
		ent.GetKonnectStatus().OrgID = apiAuth.Status.OrganizationID
		if err := r.Client.Status().Update(ctx, ent); err != nil {
			if k8serrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
		}

		if controllerutil.AddFinalizer(ent, KonnectCleanupFinalizer) {
			if err := r.Client.Update(ctx, ent); err != nil {
				if k8serrors.IsConflict(err) {
					return ctrl.Result{Requeue: true}, nil
				}
				return ctrl.Result{}, fmt.Errorf("failed to update finalizer: %w", err)
			}
		}

		// NOTE: we don't need to requeue here because the object update will
		// trigger another reconciliation.
		return ctrl.Result{}, nil
	}

	if res, err := Update[T, TEnt](ctx, sdk, logger, r.Client, ent); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update object: %w", err)
	} else if res.Requeue || res.RequeueAfter > 0 {
		return res, nil
	}

	ent.GetKonnectStatus().ServerURL = apiAuth.Spec.ServerURL
	ent.GetKonnectStatus().OrgID = apiAuth.Status.OrganizationID
	if err := r.Client.Status().Update(ctx, ent); err != nil {
		if k8serrors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to update in cluster resource after Konnect update: %w", err)
	}

	return ctrl.Result{
		RequeueAfter: configurableSyncPeriod,
	}, nil
}

func typeHasControlPlaneRef[T SupportedKonnectEntityType, TEnt EntityType[T]](
	e TEnt,
) bool {
	switch e := any(e).(type) {
	case *konnectv1alpha1.KonnectControlPlane:
		return false
	case *configurationv1alpha1.KongService:
		return true
	case *configurationv1alpha1.KongRoute:
		return true
	case *configurationv1.KongConsumer:
		return true
	default:
		panic(fmt.Sprintf("unsupported entity type %T", e))
	}
}

func getControlPlaneRef[T SupportedKonnectEntityType, TEnt EntityType[T]](
	e TEnt,
) configurationv1alpha1.ControlPlaneRef {
	switch e := any(e).(type) {
	case *konnectv1alpha1.KonnectControlPlane:
		// TODO: handle better
		// Should never happen
		panic(fmt.Sprintf("unsupported entity type %T", e))
	case *configurationv1alpha1.KongService:
		return e.Spec.ControlPlaneRef
	case *configurationv1alpha1.KongRoute:
		return e.Spec.ControlPlaneRef
	case *configurationv1.KongConsumer:
		return e.Spec.ControlPlaneRef
	default:
		panic(fmt.Sprintf("unsupported entity type %T", e))
	}
}

func typeHasServiceRef[T SupportedKonnectEntityType, TEnt EntityType[T]](
	e TEnt,
) bool {
	switch any(e).(type) {
	case *konnectv1alpha1.KonnectControlPlane:
		return false
	case *configurationv1alpha1.KongService:
		return false
	case *configurationv1alpha1.KongRoute:
		return true
	case *configurationv1.KongConsumer:
		return false
	default:
		return false
	}
}

func getServiceRef[T SupportedKonnectEntityType, TEnt EntityType[T]](
	e TEnt,
) configurationv1alpha1.ServiceRef {
	switch e := any(e).(type) {
	case *konnectv1alpha1.KonnectControlPlane, *configurationv1alpha1.KongService:
		// TODO: handle better
		// Should never happen
		panic(fmt.Sprintf("unsupported entity type %T", e))
	case *configurationv1alpha1.KongRoute:
		return e.Spec.ServiceRef
	default:
		panic(fmt.Sprintf("unsupported entity type %T", e))
	}
}

func updateStatusWithCondition[T interface {
	client.Object
	k8sutils.ConditionsAware
}](
	ctx context.Context,
	client client.Client,
	ent T,
	conditionType consts.ConditionType,
	conditionStatus metav1.ConditionStatus,
	conditionReason consts.ConditionReason,
	conditionMessage string,
) (ctrl.Result, error) {
	k8sutils.SetCondition(
		k8sutils.NewConditionWithGeneration(
			conditionType,
			conditionStatus,
			conditionReason,
			conditionMessage,
			ent.GetGeneration(),
		),
		ent,
	)

	if err := client.Status().Update(ctx, ent); err != nil {
		if k8serrors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, fmt.Errorf(
			"failed to update status with %s condition: %w",
			KonnectEntityAPIAuthConfigurationResolvedRefConditionType, err,
		)
	}

	return ctrl.Result{}, nil
}

//nolint:unused
func conditionMessageReferenceKonnectAPIAuthConfigurationInvalid(apiAuthRef types.NamespacedName) string {
	return fmt.Sprintf("referenced KonnectAPIAuthConfiguration %s is invalid", apiAuthRef)
}

//nolint:unused
func conditionMessageReferenceKonnectAPIAuthConfigurationValid(apiAuthRef types.NamespacedName) string {
	return fmt.Sprintf("referenced KonnectAPIAuthConfiguration %s is valid", apiAuthRef)
}
