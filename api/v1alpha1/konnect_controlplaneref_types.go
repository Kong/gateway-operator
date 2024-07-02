package v1alpha1

// KonnectControlPlaneRef is the Schema for the konnectcontrolplanes API.
type KonnectControlPlaneRef struct {
	// Type can be one of:
	// - KonnectID
	// - KonnectNamespacedRef
	// - KIC
	Type string `json:"type,omitempty"`

	// TODO(pmalek)
	KonnectID *string `json:"konnectID,omitempty"`

	KonnectNamespacedRef *KonnectNamespacedRef `json:"konnectNamespacedRef,omitempty"`

	// TODO(pmalek)
	KIC *KIC `json:"kic,omitempty"`
}

// KIC is the Schema for the konnectcontrolplanes API.
// TODO(pmalek)
type KIC struct{}

// KonnectNamespacedRef is the Schema for the konnectnamespacedrefs API.
type KonnectNamespacedRef struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	Namespace string `json:"namespace,omitempty"`
}
