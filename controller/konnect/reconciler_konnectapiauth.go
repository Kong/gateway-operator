package konnect

import (
	"context"
	"fmt"
	"time"

	sdkkonnectops "github.com/Kong/sdk-konnect-go/models/operations"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/kong/gateway-operator/controller/konnect/ops"
	"github.com/kong/gateway-operator/controller/pkg/log"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"

	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

// KonnectAPIAuthConfigurationReconciler reconciles a KonnectAPIAuthConfiguration object.
type KonnectAPIAuthConfigurationReconciler struct {
	sdkFactory      ops.SDKFactory
	developmentMode bool
	client          client.Client
}

const (
	// SecretTokenKey is the key used to store the token in the Secret.
	SecretTokenKey = "token"
	// SecretCredentialLabel is the label used to identify Secrets holding
	// KonnectAPIAuthConfiguration tokens.
	SecretCredentialLabel = "konghq.com/credential" //nolint:gosec
	// SecretCredentialLabelValueKonnect is the value of the label used to
	// identify Secrets holding KonnectAPIAuthConfiguration tokens.
	SecretCredentialLabelValueKonnect = "konnect"
)

// NewKonnectAPIAuthConfigurationReconciler creates a new KonnectAPIAuthConfigurationReconciler.
func NewKonnectAPIAuthConfigurationReconciler(
	sdkFactory ops.SDKFactory,
	developmentMode bool,
	client client.Client,
) *KonnectAPIAuthConfigurationReconciler {
	return &KonnectAPIAuthConfigurationReconciler{
		sdkFactory:      sdkFactory,
		developmentMode: developmentMode,
		client:          client,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KonnectAPIAuthConfigurationReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	secretLabelPredicate, err := predicate.LabelSelectorPredicate(
		metav1.LabelSelector{
			MatchLabels: map[string]string{
				SecretCredentialLabel: SecretCredentialLabelValueKonnect,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create Secret label selector predicate: %w", err)
	}

	b := ctrl.NewControllerManagedBy(mgr).
		For(&konnectv1alpha1.KonnectAPIAuthConfiguration{}).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(
				listKonnectAPIAuthConfigurationsReferencingSecret(mgr.GetClient()),
			),
			builder.WithPredicates(secretLabelPredicate),
		).
		Named("KonnectAPIAuthConfiguration")

	return b.Complete(r)
}

// Reconcile reconciles a KonnectAPIAuthConfiguration object.
func (r *KonnectAPIAuthConfigurationReconciler) Reconcile(
	ctx context.Context, req ctrl.Request,
) (ctrl.Result, error) {
	var apiAuth konnectv1alpha1.KonnectAPIAuthConfiguration
	if err := r.client.Get(ctx, req.NamespacedName, &apiAuth); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var (
		entityTypeName = "KonnectAPIAuthConfiguration"
		logger         = log.GetLogger(ctx, entityTypeName, r.developmentMode)
	)

	log.Debug(logger, "reconciling", apiAuth)
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

	token, err := getTokenFromKonnectAPIAuthConfiguration(ctx, r.client, &apiAuth)
	if err != nil {
		if res, errStatus := updateStatusWithCondition(
			ctx, r.client, &apiAuth,
			konnectv1alpha1.KonnectEntityAPIAuthConfigurationValidConditionType,
			metav1.ConditionFalse,
			konnectv1alpha1.KonnectEntityAPIAuthConfigurationReasonInvalid,
			err.Error(),
		); errStatus != nil || !res.IsZero() {
			return res, errStatus
		}
		return ctrl.Result{}, err
	}

	serverURL := ops.NewServerURL(apiAuth.Spec.ServerURL)
	sdk := r.sdkFactory.NewKonnectSDK(
		serverURL.String(),
		ops.SDKToken(token),
	)

	// TODO(pmalek): check if api auth config has a valid status condition
	// If not then return an error.

	// NOTE: /organizations/me is not public in OpenAPI spec so we can use it
	// but not using the SDK
	// https://kongstrong.slack.com/archives/C04RXLGNB6K/p1719830395775599?thread_ts=1719406468.883729&cid=C04RXLGNB6K

	// NOTE: This is needed because currently the SDK only lists the prod global API as supported:
	// https://github.com/Kong/sdk-konnect-go/blob/999d9a987e1aa7d2e09ac11b1450f4563adf21ea/models/operations/getorganizationsme.go#L10-L12
	respOrg, err := sdk.GetMeSDK().GetOrganizationsMe(ctx, sdkkonnectops.WithServerURL(serverURL.String()))
	if err != nil {
		logger.Error(err, "failed to get organization info from Konnect")
		if cond, ok := k8sutils.GetCondition(konnectv1alpha1.KonnectEntityAPIAuthConfigurationValidConditionType, &apiAuth); !ok ||
			cond.Status != metav1.ConditionFalse ||
			cond.Reason != konnectv1alpha1.KonnectEntityAPIAuthConfigurationReasonInvalid ||
			cond.ObservedGeneration != apiAuth.GetGeneration() ||
			apiAuth.Status.OrganizationID != "" ||
			apiAuth.Status.ServerURL != serverURL.String() {

			apiAuth.Status.OrganizationID = ""
			apiAuth.Status.ServerURL = serverURL.String()

			res, errUpdate := updateStatusWithCondition(
				ctx, r.client, &apiAuth,
				konnectv1alpha1.KonnectEntityAPIAuthConfigurationValidConditionType,
				metav1.ConditionFalse,
				konnectv1alpha1.KonnectEntityAPIAuthConfigurationReasonInvalid,
				err.Error(),
			)
			if errUpdate != nil || !res.IsZero() {
				return res, errUpdate
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
	}

	// Update the status only if it would change to prevent unnecessary updates.
	condMessage := "Token is valid"
	if apiAuth.Spec.Type == konnectv1alpha1.KonnectAPIAuthTypeSecretRef {
		nn := types.NamespacedName{
			Namespace: apiAuth.Spec.SecretRef.Namespace,
			Name:      apiAuth.Spec.SecretRef.Name,
		}
		if nn.Namespace == "" {
			nn.Namespace = apiAuth.Namespace
		}
		condMessage = fmt.Sprintf("Token from Secret %s is valid", nn)
	}
	if cond, ok := k8sutils.GetCondition(konnectv1alpha1.KonnectEntityAPIAuthConfigurationValidConditionType, &apiAuth); !ok ||
		cond.Status != metav1.ConditionTrue ||
		cond.Message != condMessage ||
		cond.Reason != konnectv1alpha1.KonnectEntityAPIAuthConfigurationReasonValid ||
		cond.ObservedGeneration != apiAuth.GetGeneration() ||
		apiAuth.Status.OrganizationID != *respOrg.MeOrganization.ID ||
		apiAuth.Status.ServerURL != serverURL.String() {

		apiAuth.Status.OrganizationID = *respOrg.MeOrganization.ID
		apiAuth.Status.ServerURL = serverURL.String()

		res, err := updateStatusWithCondition(
			ctx, r.client, &apiAuth,
			konnectv1alpha1.KonnectEntityAPIAuthConfigurationValidConditionType,
			metav1.ConditionTrue,
			konnectv1alpha1.KonnectEntityAPIAuthConfigurationReasonValid,
			condMessage,
		)
		if err != nil || !res.IsZero() {
			return res, err
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// getTokenFromKonnectAPIAuthConfiguration returns the token from the secret reference or the token field.
func getTokenFromKonnectAPIAuthConfiguration(
	ctx context.Context, cl client.Client, apiAuth *konnectv1alpha1.KonnectAPIAuthConfiguration,
) (string, error) {
	switch apiAuth.Spec.Type {
	case konnectv1alpha1.KonnectAPIAuthTypeToken:
		return apiAuth.Spec.Token, nil
	case konnectv1alpha1.KonnectAPIAuthTypeSecretRef:
		nn := types.NamespacedName{
			Namespace: apiAuth.Spec.SecretRef.Namespace,
			Name:      apiAuth.Spec.SecretRef.Name,
		}
		if nn.Namespace == "" {
			nn.Namespace = apiAuth.Namespace
		}

		var secret corev1.Secret
		if err := cl.Get(ctx, nn, &secret); err != nil {
			return "", fmt.Errorf("failed to get Secret %s: %w", nn, err)
		}
		if secret.Labels == nil || secret.Labels[SecretCredentialLabel] != SecretCredentialLabelValueKonnect {
			return "", fmt.Errorf("secret %s does not have label %s: %s", nn, SecretCredentialLabel, SecretCredentialLabelValueKonnect)
		}
		if secret.Data == nil {
			return "", fmt.Errorf("secret %s has no data", nn)
		}
		if _, ok := secret.Data[SecretTokenKey]; !ok {
			return "", fmt.Errorf("secret %s does not have key %s", nn, SecretTokenKey)
		}
		return string(secret.Data[SecretTokenKey]), nil
	}

	return "", fmt.Errorf("unknown KonnectAPIAuthType: %s", apiAuth.Spec.Type)
}
