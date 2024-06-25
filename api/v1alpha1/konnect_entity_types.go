package v1alpha1

type KonnectEntityStatus struct {
	KonnectID string `json:"id,omitempty"`
}

// GetStatusID returns the ID field of the KonnectEntityStatus struct
func (in KonnectEntityStatus) GetStatusID() string {
	return in.KonnectID
}
