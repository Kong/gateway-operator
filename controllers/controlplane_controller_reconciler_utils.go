package controllers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
	"github.com/kong/gateway-operator/internal/consts"
	k8sutils "github.com/kong/gateway-operator/internal/utils/kubernetes"
)

// -----------------------------------------------------------------------------
// ControlPlaneReconciler - Status Management
// -----------------------------------------------------------------------------

func (r *ControlPlaneReconciler) ensureControlPlaneIsMarkedScheduled(
	ctx context.Context,
	controlplane *operatorv1alpha1.ControlPlane,
) (bool, error) {
	isScheduled := false
	for _, condition := range controlplane.Status.Conditions {
		if condition.Type == string(ControlPlaneConditionTypeProvisioned) {
			isScheduled = true
		}
	}

	if !isScheduled {
		controlplane.Status.Conditions = append(controlplane.Status.Conditions, metav1.Condition{
			Type:               string(ControlPlaneConditionTypeProvisioned),
			Reason:             ControlPlaneConditionReasonPodsNotReady,
			Status:             metav1.ConditionFalse,
			Message:            "ControlPlane resource is scheduled for provisioning",
			ObservedGeneration: controlplane.Generation,
			LastTransitionTime: metav1.Now(),
		})
		return true, r.Client.Status().Update(ctx, controlplane)
	}

	return false, nil
}

func (r *ControlPlaneReconciler) ensureControlPlaneIsMarkedProvisioned(
	ctx context.Context,
	controlplane *operatorv1alpha1.ControlPlane,
) error {
	updatedConditions := make([]metav1.Condition, 0)
	for _, condition := range controlplane.Status.Conditions {
		if condition.Type == string(ControlPlaneConditionTypeProvisioned) {
			condition.Status = metav1.ConditionTrue
			condition.Reason = ControlPlaneConditionReasonPodsReady
			condition.Message = "pods for all Deployments are ready"
			condition.ObservedGeneration = controlplane.Generation
			condition.LastTransitionTime = metav1.Now()
		}
		updatedConditions = append(updatedConditions, condition)
	}

	controlplane.Status.Conditions = updatedConditions
	return r.Status().Update(ctx, controlplane)
}

func (r *ControlPlaneReconciler) ensureControlPlaneHasDataplane(
	ctx context.Context,
	controlplane *operatorv1alpha1.ControlPlane,
) (dataplaneSet bool, err error) {

	if controlplane.Spec.DataPlane == nil || *controlplane.Spec.DataPlane == "" {
		updatedConditions := make([]metav1.Condition, 0)
		for _, condition := range controlplane.Status.Conditions {
			if condition.Type == string(ControlPlaneConditionTypeProvisioned) {
				if condition.Status != metav1.ConditionFalse || condition.Reason != ControlPlaneConditionReasonNoDataplane {
					condition.Status = metav1.ConditionFalse
					condition.Reason = ControlPlaneConditionReasonNoDataplane
					condition.Message = "dataplane is not specified"
					condition.ObservedGeneration = controlplane.Generation
					condition.LastTransitionTime = metav1.Now()
				}
			}
			updatedConditions = append(updatedConditions, condition)
		}
		if len(updatedConditions) > 0 {
			controlplane.Status.Conditions = updatedConditions
			return false, r.Status().Update(ctx, controlplane)
		}
	}

	return true, nil
}

// -----------------------------------------------------------------------------
// ControlPlaneReconciler - Owned Resource Management
// -----------------------------------------------------------------------------

// ensureDeploymentForControlPlane ensures that a Deployment is created for the
// ControlPlane resource. Deployment will remain in dormant state until
// corresponding dataplane is set.
func (r *ControlPlaneReconciler) ensureDeploymentForControlPlane(
	ctx context.Context,
	controlplane *operatorv1alpha1.ControlPlane,
) (bool, *appsv1.Deployment, error) {
	var replicasDormantState int32 = 0
	var replicasActiveState int32 = 1

	dataplaneProvided := controlplane.Spec.DataPlane != nil && *controlplane.Spec.DataPlane != ""

	deployments, err := k8sutils.ListDeploymentsForOwner(ctx, r.Client, consts.GatewayOperatorControlledLabel, consts.ControlPlaneManagedLabelValue, controlplane.Namespace, controlplane.UID)
	if err != nil {
		return false, nil, err
	}

	count := len(deployments)
	if count > 1 {
		return false, nil, fmt.Errorf("found %d deployments for ControlPlane currently unsupported: expected 1 or less", count)
	}

	if count == 1 {
		deactivateReplicas := !dataplaneProvided &&
			(deployments[0].Spec.Replicas == nil || *deployments[0].Spec.Replicas != replicasDormantState)
		if deactivateReplicas {
			deployments[0].Spec.Replicas = &replicasDormantState
			return true, &deployments[0], r.Client.Update(ctx, &deployments[0])
		}

		activateReplicas := dataplaneProvided &&
			(deployments[0].Spec.Replicas == nil || *deployments[0].Spec.Replicas == replicasDormantState)
		if activateReplicas {
			deployments[0].Spec.Replicas = &replicasActiveState
			return true, &deployments[0], r.Client.Update(ctx, &deployments[0])
		}

		return false, &deployments[0], nil
	}

	deployment := generateNewDeploymentForControlPlane(controlplane)
	k8sutils.SetOwnerForObject(deployment, controlplane)
	labelObjForControlPlane(deployment)

	if !dataplaneProvided {
		deployment.Spec.Replicas = &replicasDormantState
	}

	return true, deployment, r.Client.Create(ctx, deployment)
}
