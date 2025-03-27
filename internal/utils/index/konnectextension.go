package index

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/kong/kubernetes-configuration/api/common/v1alpha1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

const (
	// IndexFieldKonnectExtensionOnAPIAuthConfiguration is the index field for KonnectExtension -> APIAuthConfiguration.
	IndexFieldKonnectExtensionOnAPIAuthConfiguration = "konnectExtensionAPIAuthConfigurationRef"
	// IndexFieldKonnectExtensionOnSecrets is the index field for KonnectExtension -> Secret.
	IndexFieldKonnectExtensionOnSecrets = "konnectExtensionSecretRef"
	// IndexFieldKonnectExtensionOnKonnectGatewayControlPlane is the index field for KonnectExtension -> KonnectGatewayControlPlane.
	IndexFieldKonnectExtensionOnKonnectGatewayControlPlane = "konnectExtensionKonnectGatewayControlPlaneRef"
)

// OptionsForKonnectExtension returns required Index options for KonnectExtension reconciler.
func OptionsForKonnectExtension() []Option {
	return []Option{
		{
			Object:         &konnectv1alpha1.KonnectExtension{},
			Field:          IndexFieldKonnectExtensionOnAPIAuthConfiguration,
			ExtractValueFn: konnectExtensionAPIAuthConfigurationRef,
		},
		{
			Object:         &konnectv1alpha1.KonnectExtension{},
			Field:          IndexFieldKonnectExtensionOnSecrets,
			ExtractValueFn: konnectExtensionSecretRef,
		},
		{
			Object:         &konnectv1alpha1.KonnectExtension{},
			Field:          IndexFieldKonnectExtensionOnKonnectGatewayControlPlane,
			ExtractValueFn: konnectExtensionControlPlaneRef,
		},
	}
}

func konnectExtensionAPIAuthConfigurationRef(object client.Object) []string {
	ext, ok := object.(*konnectv1alpha1.KonnectExtension)
	if !ok {
		return nil
	}

	if ext.Spec.Konnect.Configuration == nil {
		return nil
	}

	return []string{ext.Spec.Konnect.Configuration.APIAuthConfigurationRef.Name}
}

func konnectExtensionSecretRef(obj client.Object) []string {
	ext, ok := obj.(*konnectv1alpha1.KonnectExtension)
	if !ok {
		return nil
	}

	if ext.Spec.ClientAuth == nil ||
		ext.Spec.ClientAuth.CertificateSecret.CertificateSecretRef == nil {
		return nil
	}

	return []string{ext.Spec.ClientAuth.CertificateSecret.CertificateSecretRef.Name}
}

func konnectExtensionControlPlaneRef(obj client.Object) []string {
	ext, ok := obj.(*konnectv1alpha1.KonnectExtension)
	if !ok {
		return nil
	}

	if ext.Spec.Konnect.ControlPlane.Ref.Type != commonv1alpha1.ControlPlaneRefKonnectNamespacedRef {
		return nil
	}
	// TODO: add namespace to index when cross namespace reference is supported.
	return []string{ext.Spec.Konnect.ControlPlane.Ref.KonnectNamespacedRef.Name}
}
