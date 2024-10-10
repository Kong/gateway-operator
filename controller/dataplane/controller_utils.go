package dataplane

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1beta1 "github.com/kong/gateway-operator/api/v1beta1"
	"github.com/kong/gateway-operator/controller/pkg/log"
	"github.com/kong/gateway-operator/internal/versions"
	"github.com/kong/gateway-operator/pkg/consts"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
)

// -----------------------------------------------------------------------------
// DataPlane - Private Functions - Generators
// -----------------------------------------------------------------------------

func generateDataPlaneImage(dataplane *operatorv1beta1.DataPlane, defaultImage string, validators ...versions.VersionValidationOption) (string, error) {
	if dataplane.Spec.DataPlaneOptions.Deployment.PodTemplateSpec == nil {
		return defaultImage, nil // TODO: https://github.com/Kong/gateway-operator-archive/issues/20
	}

	container := k8sutils.GetPodContainerByName(&dataplane.Spec.DataPlaneOptions.Deployment.PodTemplateSpec.Spec, consts.DataPlaneProxyContainerName)
	if container != nil && container.Image != "" {
		for _, v := range validators {
			supported, err := v(container.Image)
			if err != nil {
				return "", err
			}
			if !supported {
				return "", fmt.Errorf("unsupported DataPlane image %s", container.Image)
			}
		}
		return container.Image, nil
	}

	if relatedKongImage := os.Getenv("RELATED_IMAGE_KONG"); relatedKongImage != "" {
		// RELATED_IMAGE_KONG is set by the operator-sdk when building the operator bundle.
		// https://github.com/Kong/gateway-operator-archive/issues/261
		return relatedKongImage, nil
	}

	return defaultImage, nil // TODO: https://github.com/Kong/gateway-operator-archive/issues/20
}

// -----------------------------------------------------------------------------
// DataPlane - Private Functions - Kubernetes Object Labels and Annotations
// -----------------------------------------------------------------------------

func addAnnotationsForDataPlaneIngressService(obj client.Object, dataplane operatorv1beta1.DataPlane) {
	specAnnotations := extractDataPlaneIngressServiceAnnotations(&dataplane)
	if specAnnotations == nil {
		return
	}
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	for k, v := range specAnnotations {
		annotations[k] = v
	}
	encodedSpecAnnotations, err := json.Marshal(specAnnotations)
	if err == nil {
		annotations[consts.AnnotationLastAppliedAnnotations] = string(encodedSpecAnnotations)
	}
	obj.SetAnnotations(annotations)
}

func extractDataPlaneIngressServiceAnnotations(dataplane *operatorv1beta1.DataPlane) map[string]string {
	if dataplane.Spec.DataPlaneOptions.Network.Services == nil ||
		dataplane.Spec.DataPlaneOptions.Network.Services.Ingress == nil ||
		dataplane.Spec.DataPlaneOptions.Network.Services.Ingress.Annotations == nil {
		return nil
	}

	anns := dataplane.Spec.DataPlaneOptions.Network.Services.Ingress.Annotations
	return anns
}

// extractOutdatedDataPlaneIngressServiceAnnotations returns the last applied annotations
// of ingress service from `DataPlane` spec but disappeared in current `DataPlane` spec.
func extractOutdatedDataPlaneIngressServiceAnnotations(
	dataplane *operatorv1beta1.DataPlane, existingAnnotations map[string]string,
) (map[string]string, error) {
	if existingAnnotations == nil {
		return nil, nil
	}
	lastAppliedAnnotationsEncoded, ok := existingAnnotations[consts.AnnotationLastAppliedAnnotations]
	if !ok {
		return nil, nil
	}
	outdatedAnnotations := map[string]string{}
	err := json.Unmarshal([]byte(lastAppliedAnnotationsEncoded), &outdatedAnnotations)
	if err != nil {
		return nil, fmt.Errorf("failed to decode last applied annotations: %w", err)
	}
	// If an annotation is present in last applied annotations but not in current spec of annotations,
	// the annotation is outdated and should be removed.
	// So we remove the annotations present in current spec in last applied annotations,
	// the remaining annotations are outdated and should be removed.
	currentSpecifiedAnnotations := extractDataPlaneIngressServiceAnnotations(dataplane)
	for k := range currentSpecifiedAnnotations {
		delete(outdatedAnnotations, k)
	}
	return outdatedAnnotations, nil
}

// ensureDataPlaneReadyStatus ensures that the provided DataPlane gets an up to
// date Ready status condition.
// It sets the condition based on the readiness of DataPlane's Deployment and
// its ingress Service receiving an address.
func ensureDataPlaneReadyStatus(
	ctx context.Context,
	cl client.Client,
	logger logr.Logger,
	dataplane *operatorv1beta1.DataPlane,
	generation int64,
) (ctrl.Result, error) {
	// retrieve a fresh copy of the dataplane to reduce the number of times we have to error on update
	// due to new changes when the `DataPlane` resource is very active.
	if err := cl.Get(ctx, client.ObjectKeyFromObject(dataplane), dataplane); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed getting DataPlane %s/%s: %w", dataplane.Namespace, dataplane.Name, err)
	}

	deployments, err := listDataPlaneLiveDeployments(ctx, cl, dataplane)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed listing deployments for DataPlane %s/%s: %w", dataplane.Namespace, dataplane.Name, err)
	}

	switch len(deployments) {
	case 0:
		log.Debug(logger, "Deployment for DataPlane not present yet", dataplane)

		// Set Ready to false for dataplane as the underlying deployment is not ready.
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				consts.ReadyType,
				metav1.ConditionFalse,
				consts.WaitingToBecomeReadyReason,
				consts.WaitingToBecomeReadyMessage,
				generation,
			),
			dataplane,
		)
		ensureDataPlaneReadinessStatus(dataplane, appsv1.DeploymentStatus{
			Replicas:      0,
			ReadyReplicas: 0,
		})
		res, err := patchDataPlaneStatus(ctx, cl, logger, dataplane)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed patching status (Deployment not present) for DataPlane %s/%s: %w", dataplane.Namespace, dataplane.Name, err)
		}
		if res {
			return ctrl.Result{}, nil
		}

	case 1: // Expect just 1.

	default: // More than 1.
		log.Info(logger, "expected only 1 Deployment for DataPlane", dataplane)
		return ctrl.Result{Requeue: true}, nil
	}

	deployment := deployments[0]
	if _, ready := isDeploymentReady(deployment.Status); !ready {
		log.Debug(logger, "Deployment for DataPlane not ready yet", dataplane)

		// Set Ready to false for dataplane as the underlying deployment is not ready.
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				consts.ReadyType,
				metav1.ConditionFalse,
				consts.WaitingToBecomeReadyReason,
				fmt.Sprintf("%s: Deployment %s is not ready yet", consts.WaitingToBecomeReadyMessage, deployment.Name),
				generation,
			),
			dataplane,
		)
		ensureDataPlaneReadinessStatus(dataplane, deployment.Status)
		if _, err := patchDataPlaneStatus(ctx, cl, logger, dataplane); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed patching status (Deployment not ready) for DataPlane %s/%s: %w", dataplane.Namespace, dataplane.Name, err)
		}
		return ctrl.Result{}, nil
	}

	services, err := listDataPlaneLiveServices(ctx, cl, dataplane)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed listing ingress services for DataPlane %s/%s: %w", dataplane.Namespace, dataplane.Name, err)
	}

	switch len(services) {
	case 0:
		log.Debug(logger, "Ingress Service for DataPlane not present", dataplane)

		// Set Ready to false for dataplane as the Service is not ready yet.
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				consts.ReadyType,
				metav1.ConditionFalse,
				consts.WaitingToBecomeReadyReason,
				consts.WaitingToBecomeReadyMessage,
				generation,
			),
			dataplane,
		)
		ensureDataPlaneReadinessStatus(dataplane, deployment.Status)
		_, err := patchDataPlaneStatus(ctx, cl, logger, dataplane)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed patching status (ingress Service not present) for DataPlane %s/%s: %w", dataplane.Namespace, dataplane.Name, err)
		}
		return ctrl.Result{}, nil

	case 1: // Expect just 1.

	default: // More than 1.
		log.Info(logger, "expected only 1 ingress Service for DataPlane", dataplane)
		return ctrl.Result{Requeue: true}, nil
	}

	ingressService := services[0]
	if !dataPlaneIngressServiceIsReady(&ingressService) {
		log.Debug(logger, "Ingress Service for DataPlane not ready yet", dataplane)

		// Set Ready to false for dataplane as the Service is not ready yet.
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				consts.ReadyType,
				metav1.ConditionFalse,
				consts.WaitingToBecomeReadyReason,
				fmt.Sprintf("%s: ingress Service %s is not ready yet", consts.WaitingToBecomeReadyMessage, ingressService.Name),
				generation,
			),
			dataplane,
		)
		ensureDataPlaneReadinessStatus(dataplane, deployment.Status)
		_, err := patchDataPlaneStatus(ctx, cl, logger, dataplane)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed patching status (ingress Service not ready) for DataPlane %s/%s: %w", dataplane.Namespace, dataplane.Name, err)
		}
		return ctrl.Result{}, nil
	}

	k8sutils.SetReadyWithGeneration(dataplane, generation)
	ensureDataPlaneReadinessStatus(dataplane, deployment.Status)

	if _, err := patchDataPlaneStatus(ctx, cl, logger, dataplane); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed patching status for DataPlane %s/%s: %w", dataplane.Namespace, dataplane.Name, err)
	}

	return ctrl.Result{}, nil
}

func listDataPlaneLiveDeployments(
	ctx context.Context,
	cl client.Client,
	dataplane *operatorv1beta1.DataPlane,
) ([]appsv1.Deployment, error) {
	return k8sutils.ListDeploymentsForOwner(ctx,
		cl,
		dataplane.Namespace,
		dataplane.UID,
		client.MatchingLabels{
			"app":                                dataplane.Name,
			consts.DataPlaneDeploymentStateLabel: consts.DataPlaneStateLabelValueLive,
		},
	)
}

func listDataPlaneLiveServices(
	ctx context.Context,
	cl client.Client,
	dataplane *operatorv1beta1.DataPlane,
) ([]corev1.Service, error) {
	return k8sutils.ListServicesForOwner(ctx,
		cl,
		dataplane.Namespace,
		dataplane.UID,
		client.MatchingLabels{
			"app":                             dataplane.Name,
			consts.DataPlaneServiceStateLabel: consts.DataPlaneStateLabelValueLive,
			consts.DataPlaneServiceTypeLabel:  string(consts.DataPlaneIngressServiceLabelValue),
		},
	)
}

func isDeploymentReady(deploymentStatus appsv1.DeploymentStatus) (metav1.ConditionStatus, bool) {
	// We check if the Deployment is not Ready.
	// This is the case when status has replicas set to 0 or status.availableReplicas
	// in status is less than status.replicas.
	// The second condition takes into account the time when new version (ReplicaSet)
	// is being rolled out by Deployment controller and there might be more available
	// replicas than specified in spec.replicas but we don't consider it fully ready
	// until it stabilized to be equal to status.replicas.
	// If any of those conditions is specified we mark the DataPlane as not ready yet.
	if deploymentStatus.Replicas > 0 &&
		deploymentStatus.AvailableReplicas == deploymentStatus.Replicas {
		return metav1.ConditionTrue, true
	} else {
		return metav1.ConditionFalse, false
	}
}

// -----------------------------------------------------------------------------
// DataPlane - Private Functions - extensions
// -----------------------------------------------------------------------------

// applyExtensions patches the dataplane spec by taking into account customizations from the referenced extensions.
// In case any extension is referenced, it adds a resolvedRefs condition to the dataplane, indicating the status of the
// extension reference. it returns 3 values:
//   - patched: a boolean indicating if the dataplane was patched. If the dataplane was patched, a reconciliation loop will be automatically re-triggered.
//   - requeue: a boolean indicating if the dataplane should be requeued. If the error was unexpected (e.g., because of API server error), the dataplane should be requeued.
//     In case the error is related to a misconfiguration, the dataplane does not need to be requeued, and feedback is provided into the dataplane status.
//   - err: an error in case of failure.
func applyExtensions(ctx context.Context, cl client.Client, logger logr.Logger, dataplane *operatorv1beta1.DataPlane) (patched bool, requeue bool, err error) {
	if len(dataplane.Spec.Extensions) == 0 {
		return false, false, nil
	}
	condition := k8sutils.NewConditionWithGeneration(consts.ResolvedRefsType, metav1.ConditionTrue, consts.ResolvedRefsReason, "", dataplane.GetGeneration())
	err = applyKonnectExtension(ctx, cl, dataplane)
	if err != nil {
		switch {
		case errors.Is(err, ErrCrossNamespaceReference):
			condition.Status = metav1.ConditionFalse
			condition.Reason = string(consts.RefNotPermittedReason)
			condition.Message = strings.ReplaceAll(err.Error(), "\n", " - ")
		case errors.Is(err, ErrKonnectExtensionNotFound):
			condition.Status = metav1.ConditionFalse
			condition.Reason = string(consts.InvalidExtensionRefReason)
			condition.Message = strings.ReplaceAll(err.Error(), "\n", " - ")
		case errors.Is(err, ErrClusterCertificateNotFound):
			condition.Status = metav1.ConditionFalse
			condition.Reason = string(consts.InvalidSecretRefReason)
			condition.Message = strings.ReplaceAll(err.Error(), "\n", " - ")
		default:
			return patched, true, err
		}
	}
	newDataPlane := dataplane.DeepCopy()
	k8sutils.SetCondition(condition, newDataPlane)
	patched, patchErr := patchDataPlaneStatus(ctx, cl, logger, newDataPlane)
	if patchErr != nil {
		return false, true, fmt.Errorf("failed patching status for DataPlane %s/%s: %w", dataplane.Namespace, dataplane.Name, patchErr)
	}
	return patched, false, err
}
