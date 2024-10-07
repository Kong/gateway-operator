package conditions

// TODO(pmalek): move this to Konnect API directory so that it's part of the API contract.
// https://github.com/Kong/kubernetes-configuration/issues/14

const (
	// KonnectEntityProgrammedConditionType is the type of the condition that
	// indicates whether the entity has been programmed in Konnect.
	KonnectEntityProgrammedConditionType = "Programmed"

	// KonnectEntityProgrammedReasonProgrammed is the reason for the Programmed condition.
	// It is set when the entity has been programmed in Konnect.
	KonnectEntityProgrammedReasonProgrammed = "Programmed"
	// KonnectEntityProgrammedReasonKonnectAPIOpFailed is the reason for the Programmed condition.
	// It is set when the entity has failed to be programmed in Konnect.
	KonnectEntityProgrammedReasonKonnectAPIOpFailed = "KonnectAPIOpFailed"
	// KonnectEntityProgrammedReasonFailedToResolveConsumerGroupRefs is the reason for the Programmed condition.
	// It is set when one or more KongConsumerGroup references could not be resolved.
	KonnectEntityProgrammedReasonFailedToResolveConsumerGroupRefs = "FailedToResolveConsumerGroupRefs"
	// KonnectEntityProgrammedReasonFailedToReconcileConsumerGroupsWithKonnect is the reason for the Programmed condition.
	// It is set when one or more KongConsumerGroup references could not be reconciled with Konnect.
	KonnectEntityProgrammedReasonFailedToReconcileConsumerGroupsWithKonnect = "FailedToReconcileConsumerGroupsWithKonnect"

	// KonnectGatewayControlPlaneProgrammedReasonFailedToSetControlPlaneGroupMembers
	// is the reason for the Programmed condition. It is set when the control plane
	// group members could not be set.
	KonnectGatewayControlPlaneProgrammedReasonFailedToSetControlPlaneGroupMembers = "FailedToSetControlPlaneGroupMembers"
)

const (
	// KonnectEntityAPIAuthConfigurationResolvedRefConditionType is the type of the
	// condition that indicates whether the APIAuth configuration reference is
	// valid and points to an existing APIAuth configuration.
	KonnectEntityAPIAuthConfigurationResolvedRefConditionType = "APIAuthResolvedRef"

	// KonnectEntityAPIAuthConfigurationResolvedRefReasonResolvedRef is the reason
	// used with the APIAuthResolvedRef condition type indicating that the APIAuth
	// configuration reference has been resolved.
	KonnectEntityAPIAuthConfigurationResolvedRefReasonResolvedRef = "ResolvedRef"
	// KonnectEntityAPIAuthConfigurationResolvedRefReasonRefNotFound is the reason
	// used with the APIAuthResolvedRef condition type indicating that the APIAuth
	// configuration reference could not be resolved.
	KonnectEntityAPIAuthConfigurationResolvedRefReasonRefNotFound = "RefNotFound"
	// KonnectEntityAPIAuthConfigurationResolvedRefReasonRefNotFound is the reason
	// used with the APIAuthResolvedRef condition type indicating that the APIAuth
	// configuration reference is invalid and could not be resolved.
	// Condition message can contain more information about the error.
	KonnectEntityAPIAuthConfigurationResolvedRefReasonRefInvalid = "RefInvalid"
)

const (
	// KonnectEntityAPIAuthConfigurationValidConditionType is the type of the
	// condition that indicates whether the referenced APIAuth configuration is
	// valid.
	KonnectEntityAPIAuthConfigurationValidConditionType = "APIAuthValid"

	// KonnectEntityAPIAuthConfigurationReasonValid is the reason used with the
	// APIAuthRefValid condition type indicating that the APIAuth configuration
	// referenced by the entity is valid.
	KonnectEntityAPIAuthConfigurationReasonValid = "Valid"
	// KonnectEntityAPIAuthConfigurationReasonInvalid is the reason used with the
	// APIAuthRefValid condition type indicating that the APIAuth configuration
	// referenced by the entity is invalid.
	KonnectEntityAPIAuthConfigurationReasonInvalid = "Invalid"
)

const (
	// ControlPlaneRefValidConditionType is the type of the condition that indicates
	// whether the ControlPlane reference is valid and points to an existing
	// ControlPlane.
	ControlPlaneRefValidConditionType = "ControlPlaneRefValid"

	// ControlPlaneRefReasonValid is the reason used with the ControlPlaneRefValid
	// condition type indicating that the ControlPlane reference is valid.
	ControlPlaneRefReasonValid = "Valid"
	// ControlPlaneRefReasonInvalid is the reason used with the ControlPlaneRefValid
	// condition type indicating that the ControlPlane reference is invalid.
	ControlPlaneRefReasonInvalid = "Invalid"
)

const (
	// KongServiceRefValidConditionType is the type of the condition that indicates
	// whether the KongService reference is valid and points to an existing
	// KongService.
	KongServiceRefValidConditionType = "KongServiceRefValid"

	// KongServiceRefReasonValid is the reason used with the KongServiceRefValid
	// condition type indicating that the KongService reference is valid.
	KongServiceRefReasonValid = "Valid"
	// KongServiceRefReasonInvalid is the reason used with the KongServiceRefValid
	// condition type indicating that the KongService reference is invalid.
	KongServiceRefReasonInvalid = "Invalid"
)

const (
	// KongConsumerRefValidConditionType is the type of the condition that indicates
	// whether the KongConsumer reference is valid and points to an existing
	// KongConsumer.
	KongConsumerRefValidConditionType = "KongConsumerRefValid"

	// KongConsumerRefReasonValid is the reason used with the KongConsumerRefValid
	// condition type indicating that the KongConsumer reference is valid.
	KongConsumerRefReasonValid = "Valid"
	// KongConsumerRefReasonInvalid is the reason used with the KongConsumerRefValid
	// condition type indicating that the KongConsumer reference is invalid.
	KongConsumerRefReasonInvalid = "Invalid"
)

const (
	// KongConsumerGroupRefsValidConditionType is the type of the condition that indicates
	// whether the KongConsumerGroups referenced by the entity are valid and all point to
	// existing KongConsumerGroups.
	KongConsumerGroupRefsValidConditionType = "KongConsumerGroupRefsValid"

	// KongConsumerGroupRefsReasonValid is the reason used with the KongConsumerGroupRefsValid
	// condition type indicating that all KongConsumerGroup references are valid.
	KongConsumerGroupRefsReasonValid = "Valid"
	// KongConsumerGroupRefsReasonInvalid is the reason used with the KongConsumerGroupRefsValid
	// condition type indicating that one or more KongConsumerGroup references are invalid.
	KongConsumerGroupRefsReasonInvalid = "Invalid"
)

const (
	// KongUpstreamRefValidConditionType is the type of the condition that indicates
	// whether the KongUpstream reference is valid and points to an existing KongUpstream.
	KongUpstreamRefValidConditionType = "KongUpstreamRefValid"

	// KongUpstreamRefReasonValid is the reason used with the KongUpstreamRefValid
	// condition type indicating that the KongUpstream reference is valid.
	KongUpstreamRefReasonValid = "Valid"
	// KongUpstreamRefReasonInvalid is the reason used with the KongUpstreamRefValid
	// condition type indicating that the KongUpstream reference is invalid.
	KongUpstreamRefReasonInvalid = "Invalid"
)

const (
	// KeySetRefValidConditionType is the type of the condition that indicates
	// whether the KeySet reference is valid and points to an existing
	// KeySet.
	KeySetRefValidConditionType = "KeySetRefValid"

	// KeySetRefReasonValid is the reason used with the KeySetRefValid
	// condition type indicating that the KeySet reference is valid.
	KeySetRefReasonValid = "Valid"
	// KeySetRefReasonInvalid is the reason used with the KeySetRefValid
	// condition type indicating that the KeySet reference is invalid.
	KeySetRefReasonInvalid = "Invalid"
)

const (
	// KongCertificateRefValidConditionType is the type of the condition that indicates
	// whether the KongCertificate reference is valid and points to an existing KongCertificate
	KongCertificateRefValidConditionType = "KongCertificateRefValid"

	// KongCertificateRefReasonValid is the reason used with the KongCertificateRefValid
	// condition type indicating that the KongCertificate reference is valid.
	KongCertificateRefReasonValid = "Valid"
	// KongCertificateRefReasonInvalid is the reason used with the KongCertificateRefValid
	// condition type indicating that the KongCertificate reference is invalid.
	KongCertificateRefReasonInvalid = "Invalid"
)
