package konnect

import (
	"context"
	"reflect"
	"strings"
	"time"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	"github.com/samber/lo"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kong/gateway-operator/controller/konnect/ops"
	sdkops "github.com/kong/gateway-operator/controller/konnect/ops/sdk"
	"github.com/kong/gateway-operator/controller/pkg/log"
	"github.com/kong/gateway-operator/controller/pkg/patch"
	operatorerrors "github.com/kong/gateway-operator/internal/errors"
	"github.com/kong/gateway-operator/internal/utils/index"
	"github.com/kong/gateway-operator/pkg/consts"
	konnectresource "github.com/kong/gateway-operator/pkg/utils/konnect/resources"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	operatorv1alpha1 "github.com/kong/kubernetes-configuration/api/gateway-operator/v1alpha1"
	operatorv1beta1 "github.com/kong/kubernetes-configuration/api/gateway-operator/v1beta1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

// KonnectExtensionReconciler reconciles a KonnectExtension object.
type KonnectExtensionReconciler struct {
	client.Client
	developmentMode bool
	sdkFactory      sdkops.SDKFactory
	SyncPeriod      time.Duration
}

// NewKonnectAPIAuthConfigurationReconciler creates a new KonnectAPIAuthConfigurationReconciler.
func NewKonnectExtensionReconciler(
	sdkFactory sdkops.SDKFactory,
	developmentMode bool,
	client client.Client,
	SyncPeriod time.Duration,
) *KonnectExtensionReconciler {
	return &KonnectExtensionReconciler{
		Client:          client,
		sdkFactory:      sdkFactory,
		developmentMode: developmentMode,
		SyncPeriod:      SyncPeriod,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KonnectExtensionReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	ls := metav1.LabelSelector{
		// A secret must have `konghq.com/konnect-dp-cert` label to be watched by the controller.
		// This constraint is added to prevent from watching all secrets which may cause high resource consumption.
		// TODO: https://github.com/Kong/gateway-operator/issues/1255 set label constraints of `Secret`s on manager level if possible.
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      SecretKonnectDataPlaneCertificateLabel,
				Operator: metav1.LabelSelectorOpExists,
			},
		},
	}
	labelSelectorPredicate, err := predicate.LabelSelectorPredicate(ls)
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&konnectv1alpha1.KonnectExtension{}).
		Watches(
			&operatorv1beta1.DataPlane{},
			handler.EnqueueRequestsFromMapFunc(r.listDataPlaneExtensionsReferenced),
		).
		Watches(
			&konnectv1alpha1.KonnectAPIAuthConfiguration{},
			handler.EnqueueRequestsFromMapFunc(
				enqueueObjectsForKonnectAPIAuthConfiguration[konnectv1alpha1.KonnectExtensionList](
					mgr.GetClient(),
					IndexFieldKonnectExtensionOnAPIAuthConfiguration,
				),
			),
		).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(
				enqueueKonnectExtensionsForSecret(mgr.GetClient()),
			),
			builder.WithPredicates(
				labelSelectorPredicate,
			),
		).
		Complete(r)
}

// listDataPlaneExtensionsReferenced returns a list of all the KonnectExtensions referenced by the DataPlane object.
// Maximum one reference is expected.
func (r *KonnectExtensionReconciler) listDataPlaneExtensionsReferenced(ctx context.Context, obj client.Object) []reconcile.Request {
	logger := ctrllog.FromContext(ctx)
	dataPlane, ok := obj.(*operatorv1beta1.DataPlane)
	if !ok {
		logger.Error(
			operatorerrors.ErrUnexpectedObject,
			"failed to run map funcs",
			"expected", "DataPlane", "found", reflect.TypeOf(obj),
		)
		return nil
	}

	if len(dataPlane.Spec.Extensions) == 0 {
		return nil
	}

	recs := []reconcile.Request{}

	for _, ext := range dataPlane.Spec.Extensions {
		if ext.Group != operatorv1alpha1.SchemeGroupVersion.Group ||
			ext.Kind != konnectv1alpha1.KonnectExtensionKind {
			continue
		}
		namespace := dataPlane.Namespace
		if ext.Namespace != nil && *ext.Namespace != namespace {
			continue
		}
		recs = append(recs, reconcile.Request{
			NamespacedName: client.ObjectKey{
				Namespace: namespace,
				Name:      ext.Name,
			},
		})
	}
	return recs
}

// Reconcile reconciles a KonnectExtension object.
func (r *KonnectExtensionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var ext konnectv1alpha1.KonnectExtension
	if err := r.Client.Get(ctx, req.NamespacedName, &ext); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger := log.GetLogger(ctx, konnectv1alpha1.KonnectExtensionKind, r.developmentMode)
	ctx = ctrllog.IntoContext(ctx, logger)
	log.Debug(logger, "reconciling")

	var dataPlaneList operatorv1beta1.DataPlaneList
	if err := r.List(ctx, &dataPlaneList, client.MatchingFields{
		index.KonnectExtensionIndex: client.ObjectKeyFromObject(&ext).String(),
	}); err != nil {
		return ctrl.Result{}, err
	}

	var updated bool
	switch len(dataPlaneList.Items) {
	case 0:
		updated = controllerutil.RemoveFinalizer(&ext, consts.DataPlaneKonnectExtensionFinalizer)
	default:
		updated = controllerutil.AddFinalizer(&ext, consts.DataPlaneKonnectExtensionFinalizer)
	}
	if updated {
		if err := r.Client.Update(ctx, &ext); err != nil {
			if k8serrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}

		log.Info(logger, "KonnectExtension finalizer updated")
	}

	if !ext.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(&ext, consts.DataPlaneKonnectExtensionFinalizer) {
		if ext.DeletionTimestamp.After(time.Now()) {
			log.Debug(logger, "deletion still under grace period")
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: time.Until(ext.DeletionTimestamp.Time),
			}, nil
		}

		certificateSecret, err := getCertificateSecret(ctx, r.Client, ext)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		updated = controllerutil.RemoveFinalizer(certificateSecret, consts.SecretKonnectExtensionFinalizer)
		if updated {
			if err := r.Client.Update(ctx, certificateSecret); err != nil {
				if k8serrors.IsConflict(err) {
					return ctrl.Result{Requeue: true}, nil
				}
				return ctrl.Result{}, err
			}
			log.Info(logger, "Secret finalizer removed")
		}
		return ctrl.Result{}, nil
	}

	if cond, present := k8sutils.GetCondition(konnectv1alpha1.KonnectExtensionReadyConditionType, &ext); !present ||
		(cond.Status == metav1.ConditionFalse && cond.Reason == konnectv1alpha1.KonnectExtensionReadyReasonPending) ||
		cond.ObservedGeneration != ext.GetGeneration() {
		if res, err := patch.StatusWithCondition(
			ctx, r.Client, &ext,
			consts.ConditionType(konnectv1alpha1.KonnectExtensionReadyConditionType),
			metav1.ConditionFalse,
			consts.ConditionReason(konnectv1alpha1.KonnectExtensionReadyReasonProvisioning),
			"provisioning in progress",
		); err != nil || !res.IsZero() {
			return res, err
		}
	}

	apiAuthRef, err := getKonnectAPIAuthRefNN(ctx, r.Client, &ext)
	// returning an error here instead of setting status conditions, as no error is returned at all
	// once https://github.com/Kong/gateway-operator/issues/889#issue-2695605217 is implemented.
	if err != nil {
		return ctrl.Result{}, err
	}

	var apiAuth konnectv1alpha1.KonnectAPIAuthConfiguration
	err = r.Client.Get(ctx, apiAuthRef, &apiAuth)
	if requeue, res, retErr := handleAPIAuthStatusCondition(ctx, r.Client, &ext, apiAuth, err); requeue {
		return res, retErr
	}

	token, err := getTokenFromKonnectAPIAuthConfiguration(ctx, r.Client, &apiAuth)
	if err != nil {
		if res, errStatus := patch.StatusWithCondition(
			ctx, r.Client, &ext,
			konnectv1alpha1.KonnectEntityAPIAuthConfigurationValidConditionType,
			metav1.ConditionFalse,
			konnectv1alpha1.KonnectEntityAPIAuthConfigurationReasonInvalid,
			err.Error(),
		); errStatus != nil || !res.IsZero() {
			return res, errStatus
		}
		return ctrl.Result{}, err
	}
	log.Debug(logger, "token retrieved")

	// NOTE: We need to create a new SDK instance for each reconciliation
	// because the token is retrieved in runtime through KonnectAPIAuthConfiguration.
	serverURL := ops.NewServerURL[*konnectv1alpha1.KonnectExtension](apiAuth.Spec.ServerURL)
	sdk := r.sdkFactory.NewKonnectSDK(
		serverURL.String(),
		sdkops.SDKToken(token),
	)

	cp, err := ops.GetControlPlaneByID(ctx, sdk.GetControlPlaneSDK(), *ext.Spec.KonnectControlPlane.ControlPlaneRef.KonnectID)
	if err != nil {
		_, err := patch.StatusWithCondition(
			ctx, r.Client, &ext,
			consts.ConditionType(konnectv1alpha1.ControlPlaneRefValidConditionType),
			metav1.ConditionFalse,
			consts.ConditionReason(konnectv1alpha1.ControlPlaneRefReasonInvalid),
			err.Error(),
		)
		return ctrl.Result{}, err
	}
	if res, err := patch.StatusWithCondition(
		ctx, r.Client, &ext,
		consts.ConditionType(konnectv1alpha1.ControlPlaneRefValidConditionType),
		metav1.ConditionTrue,
		consts.ConditionReason(konnectv1alpha1.ControlPlaneRefReasonValid),
		"ControlPlaneRef is valid",
	); err != nil || !res.IsZero() {
		return res, err
	}

	log.Debug(logger, "controlPlane reference validity checked")

	certificateSecret, err := getCertificateSecret(ctx, r.Client, ext)
	if err != nil {
		_, err := patch.StatusWithCondition(
			ctx, r.Client, &ext,
			consts.ConditionType(konnectv1alpha1.DataPlaneCertificateProvisionedConditionType),
			metav1.ConditionFalse,
			consts.ConditionReason(konnectv1alpha1.DataPlaneCertificateProvisionedReasonRefNotFound),
			err.Error(),
		)
		return ctrl.Result{}, err
	}

	certData, ok := certificateSecret.Data[consts.TLSCRT]
	if !ok {
		_, err := patch.StatusWithCondition(
			ctx, r.Client, &ext,
			consts.ConditionType(konnectv1alpha1.DataPlaneCertificateProvisionedConditionType),
			metav1.ConditionFalse,
			consts.ConditionReason(konnectv1alpha1.DataPlaneCertificateProvisionedReasonInvalidSecret),
			"the secret does not contain a valid tls secret",
		)
		return ctrl.Result{}, err
	}

	log.Debug(logger, "DataPlane client certificate validity checked")

	updated = controllerutil.AddFinalizer(certificateSecret, consts.SecretKonnectExtensionFinalizer)
	if updated {
		if err := r.Client.Update(ctx, certificateSecret); err != nil {
			if k8serrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}

		log.Info(logger, "Secret finalizer updated")
		return ctrl.Result{}, nil
	}

	dpCertificates, err := ops.ListKongDataPlaneClientCertificates(ctx, sdk.GetDataPlaneCertificatesSDK(), cp.ID)
	if err != nil {
		_, err := patch.StatusWithCondition(
			ctx, r.Client, &ext,
			consts.ConditionType(konnectv1alpha1.DataPlaneCertificateProvisionedConditionType),
			metav1.ConditionFalse,
			consts.ConditionReason(konnectv1alpha1.DataPlaneCertificateProvisionedReasonKonnectAPIOpFailed),
			err.Error(),
		)
		return ctrl.Result{}, err
	}

	certFound := lo.ContainsBy(dpCertificates, func(cert sdkkonnectcomp.DataPlaneClientCertificateItem) bool {
		if cert.Cert != nil {
			if strings.ReplaceAll(*cert.Cert, "\r", "") == string(certData) {
				return true
			}
		}
		return false
	})
	if !certFound {
		dpCert := konnectresource.GenerateKongDataPlaneClientCertificate(
			certificateSecret.Name,
			certificateSecret.Namespace,
			&ext.Spec.KonnectControlPlane.ControlPlaneRef,
			string(certificateSecret.Data[consts.TLSCRT]),
			func(dpCert *configurationv1alpha1.KongDataPlaneClientCertificate) {
				// setting the status as a workaround for the GetControlPlaneID method, that expects the ID to be set in the status.
				dpCert.Status.Konnect = &konnectv1alpha1.KonnectEntityStatusWithControlPlaneRef{
					ControlPlaneID: cp.ID,
				}
			},
		)
		if err := ops.CreateKongDataPlaneClientCertificate(ctx, sdk.GetDataPlaneCertificatesSDK(), &dpCert); err != nil {
			_, err := patch.StatusWithCondition(
				ctx, r.Client, &ext,
				consts.ConditionType(konnectv1alpha1.DataPlaneCertificateProvisionedConditionType),
				metav1.ConditionFalse,
				consts.ConditionReason(konnectv1alpha1.DataPlaneCertificateProvisionedReasonKonnectAPIOpFailed),
				err.Error(),
			)
			// In case of an error in the Konnect ops, the resync period will take care of a new creation attempt.
			return ctrl.Result{}, err
		}
		log.Debug(logger, "DataPlane client certificate enforced in Konnect")
	}

	if res, err := patch.StatusWithCondition(
		ctx, r.Client, &ext,
		consts.ConditionType(konnectv1alpha1.DataPlaneCertificateProvisionedConditionType),
		metav1.ConditionTrue,
		consts.ConditionReason(konnectv1alpha1.DataPlaneCertificateProvisionedReasonProvisioned),
		"DataPlane client certificate is provisioned",
	); err != nil || !res.IsZero() {
		return res, err
	}

	if ext.Status.Konnect == nil {
		ext.Status.Konnect = &konnectv1alpha1.KonnectExtensionControlPlaneStatus{
			ControlPlaneID: cp.ID,
			ClusterType:    konnectClusterTypeToCRDClusterType(cp.Config.ClusterType),
			Endpoints: konnectv1alpha1.KonnectEndpoints{
				ControlPlaneEndpoint: cp.Config.ControlPlaneEndpoint,
				TelemetryEndpoint:    cp.Config.TelemetryEndpoint,
			},
		}
		ext.Status.DataPlaneClientAuth = &konnectv1alpha1.DataPlaneClientAuthStatus{
			CertificateSecretRef: &konnectv1alpha1.SecretRef{
				Name: certificateSecret.Name,
			},
		}
		if err := r.Client.Status().Update(ctx, &ext); err != nil {
			return ctrl.Result{}, err
		}
		log.Debug(logger, "Status data updated")
	}

	if res, err := patch.StatusWithCondition(
		ctx, r.Client, &ext,
		consts.ConditionType(konnectv1alpha1.KonnectExtensionReadyConditionType),
		metav1.ConditionTrue,
		consts.ConditionReason(konnectv1alpha1.KonnectExtensionReadyReasonReady),
		"KonnectExtension is ready",
	); err != nil || !res.IsZero() {
		return res, err
	}

	log.Debug(logger, "reconciled")

	// NOTE: We requeue here to keep enforcing the state of the resource in Konnect.
	// Konnect does not allow subscribing to changes so we need to keep pushing the
	// desired state periodically.
	return ctrl.Result{
		RequeueAfter: r.SyncPeriod,
	}, nil
}
