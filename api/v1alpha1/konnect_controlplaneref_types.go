package v1alpha1

const (
	ControlPlaneRefKonnectID            = "konnectID"
	ControlPlaneRefKonnectNamespacedRef = "konnectNamespacedRef"
	ControlPlaneRefKIC                  = "kIC"
)

type ControlPlaneRef struct {
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

// TODO(pmalek)
type KIC struct{}

type KonnectNamespacedRef struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	Namespace string `json:"namespace,omitempty"`
}
