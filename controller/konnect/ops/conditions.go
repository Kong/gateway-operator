package ops

import (
	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"

	kcfgconsts "github.com/kong/kubernetes-configuration/api/common/consts"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

type entityType interface {
	SetConditions([]metav1.Condition)
	GetConditions() []metav1.Condition
	GetGeneration() int64
}

// SetKonnectEntityProgrammedCondition sets the KonnectEntityProgrammed condition to true
// on the provided object.
func SetKonnectEntityProgrammedCondition(
	obj entityType,
) {
	_setKonnectEntityProgrammedConditon(
		obj,
		metav1.ConditionTrue,
		konnectv1alpha1.KonnectEntityProgrammedReasonProgrammed,
		"",
	)
}

// SetKonnectEntityProgrammedConditionFalse sets the KonnectEntityProgrammed condition
// to false on the provided object.
func SetKonnectEntityProgrammedConditionFalse(
	obj entityType,
	reason kcfgconsts.ConditionReason,
	err error,
) {
	// Clear the instance field from the error to avoid requeueing the resource
	// because of the trace ID in the instance field is different for each request.
	err = ClearInstanceFromError(err)

	_setKonnectEntityProgrammedConditon(
		obj,
		metav1.ConditionFalse,
		reason,
		err.Error(),
	)
}

func _setKonnectEntityProgrammedConditon(
	obj entityType,
	status metav1.ConditionStatus,
	reason kcfgconsts.ConditionReason,
	msg string,
) {
	k8sutils.SetCondition(
		k8sutils.NewConditionWithGeneration(
			konnectv1alpha1.KonnectEntityProgrammedConditionType,
			status,
			reason,
			msg,
			obj.GetGeneration(),
		),
		obj,
	)
}

const (
	// ControlPlaneGroupMembersReferenceResolvedConditionType sets the condition for control plane groups
	// to show whether all of its members are programmed and attached to the group.
	ControlPlaneGroupMembersReferenceResolvedConditionType = "MembersReferenceResolved"
	// ControlPlaneGroupMembersReferenceResolvedReasonResolved indicates that all members of the control plane group
	// are created and attached to the group in Konnect.
	ControlPlaneGroupMembersReferenceResolvedReasonResolved kcfgconsts.ConditionReason = "Resolved"
	// ControlPlaneGroupMembersReferenceResolvedReasonPartialNotResolved indicates that some members of the control plane group
	// are not resolved (not found or not created in Konnect).
	ControlPlaneGroupMembersReferenceResolvedReasonPartialNotResolved kcfgconsts.ConditionReason = "SomeMemberNotResolved"
	// ControlPlaneGroupMembersReferenceResolvedReasonFailedToSet indicates that error happened on setting control plane as
	// member of the control plane.
	ControlPlaneGroupMembersReferenceResolvedReasonFailedToSet kcfgconsts.ConditionReason = "SetGroupMemberFailed"
)

// SetControlPlaneGroupMembersReferenceResolvedCondition sets MembersReferenceResolved condition of control plane to True.
func SetControlPlaneGroupMembersReferenceResolvedCondition(
	cpGroup *konnectv1alpha1.KonnectGatewayControlPlane,
) {
	_setControlPlaneGroupMembersReferenceResolvedCondition(
		cpGroup,
		metav1.ConditionTrue,
		ControlPlaneGroupMembersReferenceResolvedReasonResolved,
		"",
	)
}

// SetControlPlaneGroupMembersReferenceResolvedConditionFalse sets MembersReferenceResolved condition of control plane to False.
func SetControlPlaneGroupMembersReferenceResolvedConditionFalse(
	cpGroup *konnectv1alpha1.KonnectGatewayControlPlane,
	reason kcfgconsts.ConditionReason,
	msg string,
) {
	_setControlPlaneGroupMembersReferenceResolvedCondition(
		cpGroup,
		metav1.ConditionFalse,
		reason,
		msg,
	)
}

func _setControlPlaneGroupMembersReferenceResolvedCondition(
	cpGroup *konnectv1alpha1.KonnectGatewayControlPlane,
	status metav1.ConditionStatus,
	reason kcfgconsts.ConditionReason,
	msg string,
) {
	if cpGroup.Spec.ClusterType == nil || *cpGroup.Spec.ClusterType != sdkkonnectcomp.CreateControlPlaneRequestClusterTypeClusterTypeControlPlaneGroup {
		return
	}
	k8sutils.SetCondition(
		k8sutils.NewConditionWithGeneration(
			ControlPlaneGroupMembersReferenceResolvedConditionType,
			status,
			reason,
			msg,
			cpGroup.GetGeneration(),
		),
		cpGroup,
	)
}
