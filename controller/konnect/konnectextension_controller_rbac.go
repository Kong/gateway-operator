package konnect

// -----------------------------------------------------------------------------
// KonnectExtensionReconciler - RBAC
// -----------------------------------------------------------------------------

// +kubebuilder:rbac:groups=gateway-operator.konghq.com,resources=dataplanes,verbs=get;list;watch
// +kubebuilder:rbac:groups=gateway-operator.konghq.com,resources=konnectextensions,verbs=get;list;watch
// +kubebuilder:rbac:groups=gateway-operator.konghq.com,resources=konnectextensions/finalizers,verbs=update;patch

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="konnect.konghq.com",resources="konnectgatewaycontrolplanes",verbs=get;list;watch
