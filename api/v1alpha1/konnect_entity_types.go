package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type KonnectEntityStatus struct {
	// KonnectID is the unique identifier of the Konnect entity.
	KonnectID string `json:"id,omitempty"`

	// Conditions describe the status of the Konnect entity.
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=8
	// +kubebuilder:default={{type: "Programmed", status: "Unknown", reason:"Pending", message:"Waiting for controller", lastTransitionTime: "1970-01-01T00:00:00Z"}}
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// GetStatusID returns the ID field of the KonnectEntityStatus struct
func (in KonnectEntityStatus) GetStatusID() string {
	return in.KonnectID
}

// GetConditions returns the Status Conditions
func (in KonnectEntityStatus) GetConditions() []metav1.Condition {
	return in.Conditions
}

// SetConditions sets the Status Conditions
func (in *KonnectEntityStatus) SetConditions(conditions []metav1.Condition) {
	in.Conditions = conditions
}
