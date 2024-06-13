package dataplane

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1beta1 "github.com/kong/gateway-operator/api/v1beta1"
	"github.com/kong/gateway-operator/controller/pkg/address"
	"github.com/kong/gateway-operator/controller/pkg/log"
	"github.com/kong/gateway-operator/pkg/consts"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
)

// -----------------------------------------------------------------------------
// DataPlaneReconciler - Status Management
// -----------------------------------------------------------------------------

// ensureDataPlaneReadinessStatus ensures the readiness Status fields of DataPlane are set.
func ensureDataPlaneReadinessStatus(
	dataplane *operatorv1beta1.DataPlane,
	dataplaneDeploymentStatus appsv1.DeploymentStatus,
) {
	dataplane.Status.Replicas = dataplaneDeploymentStatus.Replicas
	dataplane.Status.ReadyReplicas = dataplaneDeploymentStatus.ReadyReplicas
}

func (r *Reconciler) ensureDataPlaneServiceStatus(
	ctx context.Context,
	log logr.Logger,
	dataplane *operatorv1beta1.DataPlane,
	dataplaneServiceName string,
) (bool, error) {
	shouldUpdate := false
	if dataplane.Status.Service != dataplaneServiceName {
		dataplane.Status.Service = dataplaneServiceName
		shouldUpdate = true
	}

	if shouldUpdate {
		_, err := patchDataPlaneStatus(ctx, r.Client, log, dataplane)
		return true, err
	}
	return false, nil
}

// ensureDataPlaneAddressesStatus ensures that provided DataPlane's status addresses
// are as expected and pathes its status if there's a difference between the
// current state and what's expected.
// It returns a boolean indicating if the patch has been trigerred and an error.
func (r *Reconciler) ensureDataPlaneAddressesStatus(
	ctx context.Context,
	log logr.Logger,
	dataplane *operatorv1beta1.DataPlane,
	dataplaneService *corev1.Service,
) (bool, error) {
	addresses, err := address.AddressesFromService(dataplaneService)
	if err != nil {
		return false, fmt.Errorf("failed getting addresses for service %s: %w", dataplaneService, err)
	}

	// Compare the lengths prior to cmp.Equal() because cmp.Equal() will return
	// false when comparing nil slice and 0 length slice.
	if len(addresses) != len(dataplane.Status.Addresses) ||
		!cmp.Equal(addresses, dataplane.Status.Addresses) {
		dataplane.Status.Addresses = addresses
		_, err := patchDataPlaneStatus(ctx, r.Client, log, dataplane)
		return true, err
	}

	return false, nil
}

// isSameDataPlaneCondition returns true if two `metav1.Condition`s
// indicates the same condition of a `DataPlane` resource.
func isSameDataPlaneCondition(condition1, condition2 metav1.Condition) bool {
	return condition1.Type == condition2.Type &&
		condition1.Status == condition2.Status &&
		condition1.Reason == condition2.Reason &&
		condition1.Message == condition2.Message
}

func (r *Reconciler) ensureDataPlaneIsMarkedNotReady(
	ctx context.Context,
	log logr.Logger,
	dataplane *operatorv1beta1.DataPlane,
	reason consts.ConditionReason, message string,
) error {
	notReadyCondition := metav1.Condition{
		Type:               string(consts.ReadyType),
		Status:             metav1.ConditionFalse,
		Reason:             string(reason),
		Message:            message,
		ObservedGeneration: dataplane.Generation,
		LastTransitionTime: metav1.Now(),
	}

	conditionFound := false
	shouldUpdate := false
	for i, condition := range dataplane.Status.Conditions {
		// update the condition if condition has type `Ready`, and the condition is not the same.
		if condition.Type == string(consts.ReadyType) {
			conditionFound = true
			// update the slice if the condition is not the same as we expected.
			if !isSameDataPlaneCondition(notReadyCondition, condition) {
				dataplane.Status.Conditions[i] = notReadyCondition
				shouldUpdate = true
			}
		}
	}

	if !conditionFound {
		// append a new condition if Ready condition is not found.
		dataplane.Status.Conditions = append(dataplane.Status.Conditions, notReadyCondition)
		shouldUpdate = true
	}

	if shouldUpdate {
		_, err := patchDataPlaneStatus(ctx, r.Client, log, dataplane)
		return err
	}
	return nil
}

// ensureDataPlaneIngressServiceAnnotationsUpdated updates annotations of existing ingress service
// owned by the `DataPlane`. It first removes outdated annotations and then update annotations
// in current spec of `DataPlane`.
func ensureDataPlaneIngressServiceAnnotationsUpdated(
	dataplane *operatorv1beta1.DataPlane, existingAnnotations map[string]string, generatedAnnotations map[string]string,
) (bool, map[string]string, error) {
	// Remove annotations applied from previous version of DataPlane but removed in the current version.
	// Should be done before updating new annotations, because the updating process will overwrite the annotation
	// to save last applied annotations.
	outdatedAnnotations, err := extractOutdatedDataPlaneIngressServiceAnnotations(dataplane, existingAnnotations)
	if err != nil {
		return true, existingAnnotations, fmt.Errorf("failed to extract outdated annotations: %w", err)
	}
	var shouldUpdate bool
	for k := range outdatedAnnotations {
		if _, ok := existingAnnotations[k]; ok {
			delete(existingAnnotations, k)
			shouldUpdate = true
		}
	}
	if generatedAnnotations != nil && existingAnnotations == nil {
		existingAnnotations = map[string]string{}
	}
	// set annotations by current specified ingress service annotations.
	for k, v := range generatedAnnotations {
		if existingAnnotations[k] != v {
			existingAnnotations[k] = v
			shouldUpdate = true
		}
	}
	return shouldUpdate, existingAnnotations, nil
}

// dataPlaneIngressServiceIsReady returns:
//   - true for DataPlanes that do not have the Ingress Service type set as LoadBalancer
//   - true for DataPlanes that have the Ingress Service type set as LoadBalancer and
//     which have at least one IP or Hostname in their Ingress Service Status
//   - false otherwise.
func dataPlaneIngressServiceIsReady(dataplaneIngressService *corev1.Service) bool {
	// If the DataPlane ingress Service is not of a LoadBalancer type then
	// report the DataPlane as Ready.
	// We don't check DataPlane spec to see if the Service is of type LoadBalancer
	// because we might be relying on the default Service type which might change.
	if dataplaneIngressService.Spec.Type != corev1.ServiceTypeLoadBalancer {
		return true
	}

	ingressStatuses := dataplaneIngressService.Status.LoadBalancer.Ingress
	// If there are ingress statuses attached to the ingress Service, check
	// if there are IPs of Hostnames specified.
	// If that's the case, the DataPlane is Ready.
	for _, ingressStatus := range ingressStatuses {
		if ingressStatus.Hostname != "" || ingressStatus.IP != "" {
			return true
		}
	}
	// Otherwise the DataPlane is not Ready.
	return false
}

// patchDataPlaneStatus patches the resource status only when there are changes
// that requires it.
func patchDataPlaneStatus(ctx context.Context, cl client.Client, logger logr.Logger, updated *operatorv1beta1.DataPlane) (bool, error) {
	current := &operatorv1beta1.DataPlane{}

	err := cl.Get(ctx, client.ObjectKeyFromObject(updated), current)
	if err != nil && !k8serrors.IsNotFound(err) {
		return false, err
	}

	if k8sutils.NeedsUpdate(current, updated) ||
		addressesChanged(current, updated) ||
		readinessChanged(current, updated) ||
		current.Status.Service != updated.Status.Service ||
		current.Status.Selector != updated.Status.Selector {

		log.Debug(logger, "patching DataPlane status", updated, "status", updated.Status)
		return true, cl.Status().Patch(ctx, updated, client.MergeFrom(current))
	}

	return false, nil
}

// addressesChanged returns a boolean indicating whether the addresses in provided
// DataPlane stauses differ.
func addressesChanged(current, updated *operatorv1beta1.DataPlane) bool {
	return !cmp.Equal(current.Status.Addresses, updated.Status.Addresses)
}

func readinessChanged(current, updated *operatorv1beta1.DataPlane) bool {
	return current.Status.ReadyReplicas != updated.Status.ReadyReplicas ||
		current.Status.Replicas != updated.Status.Replicas
}
