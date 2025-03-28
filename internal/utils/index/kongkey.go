package index

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/kong/kubernetes-configuration/api/common/v1alpha1"
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

const (
	// IndexFieldKongKeyOnKongKeySetReference is the index field for KongKey-> KongKeySet.
	IndexFieldKongKeyOnKongKeySetReference = "kongKeySetRef"

	// IndexFieldKongKeyOnKonnectGatewayControlPlane is the index field for KongKey -> KonnectGatewayControlPlane.
	IndexFieldKongKeyOnKonnectGatewayControlPlane = "kongKeyKonnectGatewayControlPlaneRef"
)

// OptionsForKongKey returns required Index options for KongKey reconclier.
func OptionsForKongKey(cl client.Client) []Option {
	return []Option{
		{
			Object:         &configurationv1alpha1.KongKey{},
			Field:          IndexFieldKongKeyOnKongKeySetReference,
			ExtractValueFn: kongKeySetRefFromKongKey,
		},
		{
			Object:         &configurationv1alpha1.KongKey{},
			Field:          IndexFieldKongKeyOnKonnectGatewayControlPlane,
			ExtractValueFn: indexKonnectGatewayControlPlaneRef[configurationv1alpha1.KongKey](cl),
		},
	}
}

// kongKeySetRefFromKongKey returns namespace/name of referenced KongKeySet in KongKey spec.
func kongKeySetRefFromKongKey(obj client.Object) []string {
	key, ok := obj.(*configurationv1alpha1.KongKey)
	if !ok {
		return nil
	}

	if key.Spec.KeySetRef == nil ||
		key.Spec.KeySetRef.Type != commonv1alpha1.ObjectRefTypeNamespacedRef ||
		key.Spec.KeySetRef.NamespacedRef == nil {
		return nil
	}

	return []string{key.GetNamespace() + "/" + key.Spec.KeySetRef.NamespacedRef.Name}
}
