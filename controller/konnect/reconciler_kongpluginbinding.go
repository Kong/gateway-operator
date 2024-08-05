package konnect

import (
	"context"
	"errors"
	"fmt"
	"time"

	sdkkonnectgo "github.com/Kong/sdk-konnect-go"
	sdkkonnectgocomp "github.com/Kong/sdk-konnect-go/models/components"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kong/gateway-operator/controller/pkg/log"
	"github.com/kong/gateway-operator/controller/pkg/op"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

type KongPluginBindingReconciler struct {
	DevelopmentMode bool
	Client          client.Client
}

func NewKongPluginBindingReconciler(
	developmentMode bool,
	client client.Client,
) *KongPluginBindingReconciler {
	return &KongPluginBindingReconciler{
		DevelopmentMode: developmentMode,
		Client:          client,
	}
}

func (r *KongPluginBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	b := ctrl.NewControllerManagedBy(mgr).
		For(&configurationv1alpha1.KongPluginBinding{}).
		Named("KongPluginBinding")

	return b.Complete(r)
}

func (r *KongPluginBindingReconciler) Reconcile(
	ctx context.Context, req ctrl.Request,
) (ctrl.Result, error) {
	var (
		entityTypeName = "KongPluginBinding"
		logger         = log.GetLogger(ctx, entityTypeName, r.DevelopmentMode)
	)

	var kongPluginBinding configurationv1alpha1.KongPluginBinding
	if err := r.Client.Get(ctx, req.NamespacedName, &kongPluginBinding); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	log.Debug(logger, "reconciling", kongPluginBinding)

	var (
		apiAuth        konnectv1alpha1.KonnectAPIAuthConfiguration
		apiAuthRef     *types.NamespacedName
		kongService    configurationv1alpha1.KongService
		controlPlaneID string
	)
	if kongPluginBinding.Spec.Kong.ServiceReference != nil {
		if err := r.Client.Get(ctx, client.ObjectKey{
			Namespace: kongPluginBinding.Namespace,
			Name:      kongPluginBinding.Spec.Kong.ServiceReference.Name,
		}, &kongService); err != nil {
			return ctrl.Result{}, err
		}

		if kongService.Status.Konnect.ControlPlaneID == "" {
			return ctrl.Result{}, nil
		}

		controlPlaneID = kongService.Status.Konnect.ControlPlaneID
		apiAuthRef = &types.NamespacedName{
			Name:      kongService.GetKonnectAPIAuthConfigurationRef().Name,
			Namespace: kongService.GetNamespace(),
			// TODO(pmalek): enable if cross namespace refs are allowed
			// Namespace: ent.GetKonnectAPIAuthConfigurationRef().Namespace,
		}
	}

	if apiAuthRef == nil {
		return ctrl.Result{}, errors.New("no entity with an API Auth Configuration reference found")
	}

	if err := r.Client.Get(ctx, *apiAuthRef, &apiAuth); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get KonnectAPIAuthConfiguration: %w", err)
	}

	if cond, present := k8sutils.GetCondition(KonnectEntityAPIAuthConfigurationValidConditionType, &apiAuth); present && cond.Status != metav1.ConditionTrue {
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				KonnectEntityAPIAuthConfigurationResolvedRefConditionType,
				metav1.ConditionFalse,
				KonnectEntityAPIAuthConfigurationReasonInvalid,
				"",
				kongPluginBinding.GetGeneration(),
			),
			&kongPluginBinding,
		)
	}

	k8sutils.SetCondition(
		k8sutils.NewConditionWithGeneration(
			KonnectEntityAPIAuthConfigurationResolvedRefConditionType,
			metav1.ConditionTrue,
			KonnectEntityAPIAuthConfigurationReasonValid,
			fmt.Sprintf("referenced KonnectAPIAuthConfiguration %s is valid", apiAuthRef),
			kongPluginBinding.GetGeneration(),
		),
		&kongPluginBinding,
	)
	if err := r.Client.Status().Update(ctx, &kongPluginBinding); err != nil {
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

	if !kongPluginBinding.GetDeletionTimestamp().IsZero() {
		logger.Info("resource is being deleted")
		// wait for termination grace period before cleaning up
		if kongPluginBinding.GetDeletionTimestamp().After(time.Now()) {
			logger.Info("resource still under grace period, requeueing")
			return ctrl.Result{
				// Requeue when grace period expires.
				// If deletion timestamp is changed,
				// the update will trigger another round of reconciliation.
				// so we do not consider updates of deletion timestamp here.
				RequeueAfter: time.Until(kongPluginBinding.GetDeletionTimestamp().Time),
			}, nil
		}

		if controllerutil.RemoveFinalizer(&kongPluginBinding, KonnectCleanupFinalizer) {
			var (
				res op.Result
				err error
			)
			if res, err = deletePlugin(ctx, sdk, logger, r.Client, kongPluginBinding.GetKonnectStatus().ID, controlPlaneID, &kongPluginBinding); err != nil {
				return ctrl.Result{}, err
			}
			if err := r.Client.Update(ctx, &kongPluginBinding); err != nil {
				if k8serrors.IsConflict(err) {
					return ctrl.Result{Requeue: true}, nil
				}
				return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
			}
			if res == op.Deleted {
				log.Info(logger, "cleanup completed", kongPluginBinding)
			}
		}

		return ctrl.Result{}, nil
	}

	var (
		kongPlugin    configurationv1.KongPlugin
		konnectPlugin *sdkkonnectgocomp.PluginInput
		err           error
	)
	if err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: kongPluginBinding.Namespace,
		Name:      kongPluginBinding.Spec.PluginReference.Name,
	}, &kongPlugin); err != nil {
		return ctrl.Result{}, err
	}

	if konnectPlugin, err = kongPluginToCreatePlugin(&kongPlugin, &kongService); err != nil {
		return ctrl.Result{}, err
	}

	if _, err = upsertPlugin(ctx, sdk, logger, r.Client, konnectPlugin, controlPlaneID, &kongPluginBinding); err != nil {
		// TODO(pmalek): this is actually not 100% error prone because when status
		// update fails we don't store the Konnect ID and hence the reconciler
		// will try to create the resource again on next reconciliation.
		if err := r.Client.Status().Update(ctx, &kongPluginBinding); err != nil {
			if k8serrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, fmt.Errorf("failed to update status after creating object: %w", err)
		}

		return ctrl.Result{}, err
	}

	kongPluginBinding.GetKonnectStatus().ServerURL = apiAuth.Spec.ServerURL
	kongPluginBinding.GetKonnectStatus().OrgID = apiAuth.Status.OrganizationID
	if err := r.Client.Status().Update(ctx, &kongPluginBinding); err != nil {
		if k8serrors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
	}

	if controllerutil.AddFinalizer(&kongPluginBinding, KonnectCleanupFinalizer) {
		if err := r.Client.Update(ctx, &kongPluginBinding); err != nil {
			if k8serrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, fmt.Errorf("failed to update finalizer: %w", err)
		}
	}

	// NOTE: we don't need to requeue here because the object update will
	// trigger another reconciliation.
	return ctrl.Result{
		RequeueAfter: configurableSyncPeriod,
	}, nil
}
