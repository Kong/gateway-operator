package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	sdkkonnectgocomp "github.com/Kong/sdk-konnect-go/models/components"
)

func init() {
	SchemeBuilder.Register(&KonnectControlPlane{}, &KonnectControlPlaneList{})
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// KonnectControlPlane is the Schema for the konnectcontrolplanes API.
type KonnectControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec sdkkonnectgo.ControlPlanes `json:"spec,omitempty"`
	// Spec sdkkonnectgoops.GetControlPlaneRequest `json:"spec,omitempty"`
	Spec sdkkonnectgocomp.CreateControlPlaneRequest `json:"spec,omitempty"`

	Status KonnectEntityStatus `json:"status,omitempty"`
}

func (c *KonnectControlPlane) GetStatusID() string {
	return c.Status.KonnectID
}

// +kubebuilder:object:root=true
type KonnectControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []KonnectControlPlane `json:"items"`
}
