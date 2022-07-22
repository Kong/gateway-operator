package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/kong/gateway-operator/apis/v1alpha1"
	"github.com/kong/gateway-operator/internal/consts"
	k8sutils "github.com/kong/gateway-operator/internal/utils/kubernetes"
	k8sresources "github.com/kong/gateway-operator/internal/utils/kubernetes/resources"
)

// numReplicasWhenNoDataplane represents the desired number of replicas
// for the controlplane deployment when no dataplane is set.
const numReplicasWhenNoDataplane = 0

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

// ensureDataPlaneStatus ensures that the dataplane is in the correct state
// to carry on with the controlplane deployments reconciliation.
// Information about the missing dataplane is stored in the controlplane status.
func (r *ControlPlaneReconciler) ensureDataPlaneStatus(
	ctx context.Context,
	controlplane *operatorv1alpha1.ControlPlane,
) (controlPlaneChanged, dataplaneIsSet bool, err error) {
	updatedConditions := make([]metav1.Condition, 0)
	dataplaneIsSet = controlplane.Spec.DataPlane != nil && *controlplane.Spec.DataPlane != ""

	for _, condition := range controlplane.Status.Conditions {
		if condition.Type == string(ControlPlaneConditionTypeProvisioned) {
			switch {

			// Change state to NoDataplane if dataplane is not set.
			case !dataplaneIsSet && condition.Reason != ControlPlaneConditionReasonNoDataplane:
				condition = metav1.Condition{
					Type:               string(ControlPlaneConditionTypeProvisioned),
					Reason:             ControlPlaneConditionReasonNoDataplane,
					Status:             metav1.ConditionFalse,
					Message:            "DataPlane is not set",
					ObservedGeneration: controlplane.Generation,
					LastTransitionTime: metav1.Now(),
				}
				updatedConditions = append(updatedConditions, condition)

			// Change state from NoDataplane to PodsNotReady to start provisioning.
			case dataplaneIsSet && condition.Reason == ControlPlaneConditionReasonNoDataplane:
				condition = metav1.Condition{
					Type:               string(ControlPlaneConditionTypeProvisioned),
					Reason:             ControlPlaneConditionReasonPodsNotReady,
					Status:             metav1.ConditionFalse,
					Message:            "DataPlane was set, ControlPlane resource is scheduled for provisioning",
					ObservedGeneration: controlplane.Generation,
					LastTransitionTime: metav1.Now(),
				}
				updatedConditions = append(updatedConditions, condition)
			}
		}
	}

	if len(updatedConditions) > 0 {
		controlplane.Status.Conditions = updatedConditions
		return true, dataplaneIsSet, r.Status().Update(ctx, controlplane)
	}

	return false, dataplaneIsSet, nil
}

// -----------------------------------------------------------------------------
// ControlPlaneReconciler - Spec Management
// -----------------------------------------------------------------------------

func (r *ControlPlaneReconciler) ensureDataPlaneConfiguration(
	ctx context.Context,
	controlplane *operatorv1alpha1.ControlPlane,
	dataplaneServiceName string,
) error {
	changed := setControlPlaneEnvOnDataPlaneChange(
		&controlplane.Spec.ControlPlaneDeploymentOptions,
		controlplane.Namespace,
		dataplaneServiceName,
	)
	if changed {
		return r.Client.Update(ctx, controlplane)
	}
	return nil
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
	serviceAccountName string,
) (bool, *appsv1.Deployment, error) {
	dataplaneIsSet := controlplane.Spec.DataPlane != nil && *controlplane.Spec.DataPlane != ""

	deployments, err := k8sutils.ListDeploymentsForOwner(ctx,
		r.Client,
		consts.GatewayOperatorControlledLabel,
		consts.ControlPlaneManagedLabelValue,
		controlplane.Namespace,
		controlplane.UID,
	)
	if err != nil {
		return false, nil, err
	}

	count := len(deployments)
	if count > 1 {
		// if there is more than one Deployment owned by the same ControlPlane,
		// delete all of them and recreate only one as follows below
		if err := r.Client.DeleteAllOf(ctx, &appsv1.Deployment{},
			client.InNamespace(controlplane.Namespace),
			client.MatchingLabels{consts.GatewayOperatorControlledLabel: consts.ControlPlaneManagedLabelValue},
		); err != nil {
			return false, nil, err
		}
	}

	generatedDeployment := generateNewDeploymentForControlPlane(controlplane, serviceAccountName)
	k8sutils.SetOwnerForObject(generatedDeployment, controlplane)
	labelObjForControlPlane(generatedDeployment)

	if count == 1 {
		var updated bool
		existingDeployment := &deployments[0]
		updated, existingDeployment.ObjectMeta = k8sutils.EnsureObjectMetaIsUpdated(existingDeployment.ObjectMeta, generatedDeployment.ObjectMeta)
		replicas := existingDeployment.Spec.Replicas
		switch {

		// Dataplane was just unset, so we need to scale down the Deployment.
		case !dataplaneIsSet && (replicas == nil || *replicas != numReplicasWhenNoDataplane):
			existingDeployment.Spec.Replicas = pointer.Int32(numReplicasWhenNoDataplane)
			updated = true

		// Dataplane was just set, so we need to scale up the Deployment
		// and ensure the env variables that might have been changed in
		// deployment are updated.
		case dataplaneIsSet && (replicas != nil && *replicas == numReplicasWhenNoDataplane):
			existingDeployment.Spec.Replicas = nil
			if len(existingDeployment.Spec.Template.Spec.Containers[0].Env) > 0 {
				existingDeployment.Spec.Template.Spec.Containers[0].Env = controlplane.Spec.Env
			}
			updated = true
		}
		if updated {
			return true, existingDeployment, r.Client.Update(ctx, existingDeployment)
		}
		return false, existingDeployment, nil
	}

	if !dataplaneIsSet {
		generatedDeployment.Spec.Replicas = pointer.Int32(numReplicasWhenNoDataplane)
	}

	return true, generatedDeployment, r.Client.Create(ctx, generatedDeployment)
}

func (r *ControlPlaneReconciler) ensureServiceAccountForControlPlane(
	ctx context.Context,
	controlplane *operatorv1alpha1.ControlPlane,
) (sa *corev1.ServiceAccount, err error) {
	serviceAccounts, err := k8sutils.ListServiceAccountsForOwner(ctx, r.Client, consts.GatewayOperatorControlledLabel, consts.ControlPlaneManagedLabelValue, controlplane.Namespace, controlplane.UID)
	if err != nil {
		return nil, err
	}

	count := len(serviceAccounts)
	if count > 1 {
		// if there is more than one ServiceAccount owned by the same ControlPlane,
		// delete all of them and recreate only one as follows below
		if err := r.Client.DeleteAllOf(ctx, &corev1.ServiceAccount{},
			client.InNamespace(controlplane.Namespace),
			client.MatchingLabels{consts.GatewayOperatorControlledLabel: consts.ControlPlaneManagedLabelValue},
		); err != nil {
			return nil, err
		}
	}

	generatedServiceAccount := k8sresources.GenerateNewServiceAccountForControlPlane(controlplane.Namespace, controlplane.Name)
	k8sutils.SetOwnerForObject(generatedServiceAccount, controlplane)
	labelObjForControlPlane(generatedServiceAccount)

	if count == 1 {
		var updated bool
		existingServiceAccount := &serviceAccounts[0]
		if updated, existingServiceAccount.ObjectMeta = k8sutils.EnsureObjectMetaIsUpdated(existingServiceAccount.ObjectMeta, generatedServiceAccount.ObjectMeta); updated {
			return existingServiceAccount, r.Client.Update(ctx, existingServiceAccount)
		}
		return existingServiceAccount, nil
	}

	return generatedServiceAccount, r.Client.Create(ctx, generatedServiceAccount)
}

func (r *ControlPlaneReconciler) ensureClusterRoleForControlPlane(
	ctx context.Context,
	controlplane *operatorv1alpha1.ControlPlane,
) (cr *rbacv1.ClusterRole, err error) {
	clusterRoles, err := k8sutils.ListClusterRolesForOwner(ctx, r.Client, consts.GatewayOperatorControlledLabel, consts.ControlPlaneManagedLabelValue, controlplane.UID)
	if err != nil {
		return nil, err
	}

	count := len(clusterRoles)
	if count > 1 {
		// if there is more than one ClusterRole owned by the same ControlPlane,
		// delete all of them and recreate only one as follows below
		if err := r.Client.DeleteAllOf(ctx, &rbacv1.ClusterRole{},
			client.InNamespace(controlplane.Namespace),
			client.MatchingLabels{consts.GatewayOperatorControlledLabel: consts.ControlPlaneManagedLabelValue},
		); err != nil {
			return nil, err
		}
	}

	generatedClusterRole, err := k8sresources.GenerateNewClusterRoleForControlPlane(controlplane.Name, controlplane.Spec.ContainerImage)
	if err != nil {
		return nil, err
	}
	k8sutils.SetOwnerForObject(generatedClusterRole, controlplane)
	labelObjForControlPlane(generatedClusterRole)

	if count == 1 {
		var updated bool
		existingClusterRoles := &clusterRoles[0]
		if updated, existingClusterRoles.ObjectMeta = k8sutils.EnsureObjectMetaIsUpdated(existingClusterRoles.ObjectMeta, generatedClusterRole.ObjectMeta); updated {
			return existingClusterRoles, r.Client.Update(ctx, existingClusterRoles)
		}
		return existingClusterRoles, nil
	}

	return generatedClusterRole, r.Client.Create(ctx, generatedClusterRole)
}

func (r *ControlPlaneReconciler) ensureClusterRoleBindingForControlPlane(
	ctx context.Context,
	controlplane *operatorv1alpha1.ControlPlane,
	serviceAccountName string,
	clusterRoleName string,
) (crb *rbacv1.ClusterRoleBinding, err error) {
	clusterRoleBindings, err := k8sutils.ListClusterRoleBindingsForOwner(ctx, r.Client, consts.GatewayOperatorControlledLabel, consts.ControlPlaneManagedLabelValue, controlplane.UID)
	if err != nil {
		return nil, err
	}

	count := len(clusterRoleBindings)
	if count > 1 {
		// if there is more than one ClusterRoleBinding owned by the same ControlPlane,
		// delete all of them and recreate only one as follows below
		if err := r.Client.DeleteAllOf(ctx, &rbacv1.ClusterRoleBinding{},
			client.InNamespace(controlplane.Namespace),
			client.MatchingLabels{consts.GatewayOperatorControlledLabel: consts.ControlPlaneManagedLabelValue},
		); err != nil {
			return nil, err
		}
	}

	generatedClusterRoleBinding := k8sresources.GenerateNewClusterRoleBindingForControlPlane(controlplane.Namespace, controlplane.Name, serviceAccountName, clusterRoleName)
	k8sutils.SetOwnerForObject(generatedClusterRoleBinding, controlplane)
	labelObjForControlPlane(generatedClusterRoleBinding)

	if count == 1 {
		var updated bool
		existingClusterRoleBinding := &clusterRoleBindings[0]
		if updated, existingClusterRoleBinding.ObjectMeta = k8sutils.EnsureObjectMetaIsUpdated(existingClusterRoleBinding.ObjectMeta, generatedClusterRoleBinding.ObjectMeta); updated {
			return existingClusterRoleBinding, r.Client.Update(ctx, existingClusterRoleBinding)
		}
		return existingClusterRoleBinding, nil
	}

	return generatedClusterRoleBinding, r.Client.Create(ctx, generatedClusterRoleBinding)
}
