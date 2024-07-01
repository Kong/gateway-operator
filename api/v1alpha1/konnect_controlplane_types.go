package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	sdkkonnectgocomp "github.com/Kong/sdk-konnect-go/models/components"
)

func init() {
	SchemeBuilder.Register(&KonnectControlPlane{}, &KonnectControlPlaneList{})
}

// +genclient
// +kubebuilder:resource:scope=Namespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Programmed",description="The Resource is Programmed on Konnect",type=string,JSONPath=`.status.conditions[?(@.type=='Programmed')].status`
// +kubebuilder:printcolumn:name="ID",description="Konnect ID",type=string,JSONPath=`.status.id`
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.spec.konnectAPIAuthConfigurationRef) || has(oldSelf.spec.konnectAPIAuthConfigurationRef)", message="Konnect Configuration reference is immutable"

// KonnectControlPlane is the Schema for the konnectcontrolplanes API.
type KonnectControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KonnectControlPlaneSpec `json:"spec,omitempty"`

	Status KonnectEntityStatus `json:"status,omitempty"`
}

type KonnectControlPlaneSpec struct {
	sdkkonnectgocomp.CreateControlPlaneRequest `json:",inline"`

	KonnectAPIAuthConfigurationRef KonnectAPIAuthConfigurationRef `json:"konnectAPIAuthConfigurationRef,omitempty"`
}

// GetKonnectStatus returns the Konnect Status of the KonnectControlPlane.
func (c *KonnectControlPlane) GetStatus() *KonnectEntityStatus {
	return &c.Status
}

func (c KonnectControlPlane) GetTypeName() string {
	return "KonnectControlPlane"
}

func (c *KonnectControlPlane) SetKonnectLabels(labels map[string]string) {
	c.Spec.Labels = labels
}

func (c *KonnectControlPlane) GetKonnectAPIAuthConfigurationRef() KonnectAPIAuthConfigurationRef {
	return c.Spec.KonnectAPIAuthConfigurationRef
}

func (c *KonnectControlPlane) GetReconciliationWatchOptions(
	cl client.Client,
) []func(*ctrl.Builder) *ctrl.Builder {
	return []func(*ctrl.Builder) *ctrl.Builder{
		func(b *ctrl.Builder) *ctrl.Builder {
			return b.Watches(
				&KonnectAPIAuthConfiguration{},
				handler.EnqueueRequestsFromMapFunc(
					enqueueKonnectControlPlaneForKonnectAPIAuthConfiguration(cl),
				),
			)
		},
	}
}

func enqueueKonnectControlPlaneForKonnectAPIAuthConfiguration(
	cl client.Client,
) func(ctx context.Context, obj client.Object) []reconcile.Request {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		auth, ok := obj.(*KonnectAPIAuthConfiguration)
		if !ok {
			return nil
		}
		var l KonnectControlPlaneList
		if err := cl.List(ctx, &l); err != nil {
			return nil
		}
		var ret []reconcile.Request
		for _, cp := range l.Items {
			if cp.Spec.KonnectAPIAuthConfigurationRef.Name != auth.Name {
				continue
			}
			ret = append(ret, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: cp.Namespace,
					Name:      cp.Name,
				},
			})
		}
		return ret
	}
}

// +kubebuilder:object:root=true
type KonnectControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []KonnectControlPlane `json:"items"`
}
