package konnect

import (
	"context"
	"fmt"
	"time"

	sdkkonnectgo "github.com/Kong/sdk-konnect-go"
	sdkkonnectgocomp "github.com/Kong/sdk-konnect-go/models/components"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
	"github.com/kong/gateway-operator/controller/pkg/log"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
)

type KonnectAPIAuthConfigurationReconciler struct {
	DevelopmentMode bool
	Client          client.Client
}

func NewKonnectAPIAuthConfigurationReconciler(
	developmentMode bool,
	client client.Client,
) *KonnectAPIAuthConfigurationReconciler {
	return &KonnectAPIAuthConfigurationReconciler{
		DevelopmentMode: developmentMode,
		Client:          client,
	}
}

func (r *KonnectAPIAuthConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	b := ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.KonnectAPIAuthConfiguration{}).
		Named("KonnectAPIAuthConfiguration")

	return b.Complete(r)
}

func (r *KonnectAPIAuthConfigurationReconciler) Reconcile(
	ctx context.Context, req ctrl.Request,
) (ctrl.Result, error) {
	var (
		entityTypeName = "KonnectAPIAuthConfiguration"
		logger         = log.GetLogger(ctx, entityTypeName, r.DevelopmentMode)
	)

	var apiAuth operatorv1alpha1.KonnectAPIAuthConfiguration
	logger.Info("reconciling")
	if err := r.Client.Get(ctx, req.NamespacedName, &apiAuth); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if !apiAuth.GetDeletionTimestamp().IsZero() {
		logger.Info("resource is being deleted")
		// wait for termination grace period before cleaning up
		if apiAuth.GetDeletionTimestamp().After(time.Now()) {
			logger.Info("resource still under grace period, requeueing")
			return ctrl.Result{
				// Requeue when grace period expires.
				// If deletion timestamp is changed,
				// the update will trigger another round of reconciliation.
				// so we do not consider updates of deletion timestamp here.
				RequeueAfter: time.Until(apiAuth.GetDeletionTimestamp().Time),
			}, nil
		}

		return ctrl.Result{}, nil
	}

	// TODO(pmalek): check if api auth config has a valid status condition
	// If not then return an error.
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

	respOrg, err := sdk.Me.GetOrganizationsMe(ctx)
	if err != nil {
		apiAuth.Status.OrganizationID = ""
		apiAuth.Status.ServerURL = apiAuth.Spec.ServerURL
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				KonnectAPIAuthConfigurationValidConditionType,
				metav1.ConditionFalse,
				KonnectAPIAuthConfigurationReasonInvalid,
				err.Error(),
				apiAuth.GetGeneration(),
			),
			&apiAuth.Status,
		)
		if err := r.Client.Status().Update(ctx, &apiAuth); err != nil {
			if k8serrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, fmt.Errorf("failed to update status of KonnectAPIAuthConfiguration: %w", err)
		}
		return ctrl.Result{}, fmt.Errorf("failed to get organization info from Konnect: %w", err)
	}

	if cond, ok := k8sutils.GetCondition(KonnectAPIAuthConfigurationValidConditionType, &apiAuth.Status); !ok ||
		cond.Status != metav1.ConditionTrue ||
		cond.Message != "" ||
		cond.Reason != KonnectAPIAuthConfigurationReasonValid ||
		cond.ObservedGeneration != apiAuth.GetGeneration() ||
		apiAuth.Status.OrganizationID != *respOrg.MeOrganization.ID ||
		apiAuth.Status.ServerURL != apiAuth.Spec.ServerURL {
		apiAuth.Status.OrganizationID = *respOrg.MeOrganization.ID
		apiAuth.Status.ServerURL = apiAuth.Spec.ServerURL
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				KonnectAPIAuthConfigurationValidConditionType,
				metav1.ConditionTrue,
				KonnectAPIAuthConfigurationReasonValid,
				"",
				apiAuth.GetGeneration(),
			),
			&apiAuth.Status,
		)
		if err := r.Client.Status().Update(ctx, &apiAuth); err != nil {
			if k8serrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, fmt.Errorf("failed to update status of KonnectAPIAuthConfiguration: %w", err)
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{
		RequeueAfter: configurableSyncPeriod,
	}, nil
}
