package konnect

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

const (
	// IndexFieldKongCredentialHMACReferencesKongConsumer is the index name for KongCredentialHMAC -> Consumer.
	IndexFieldKongCredentialHMACReferencesKongConsumer = "kongCredentialsHMACConsumerRef"
	// IndexFieldKongCredentialHMACReferencesSecret is the index name for KongCredentialHMAC -> Secret.
	IndexFieldKongCredentialHMACReferencesSecret = "kongCredentialsHMACSecretRef"
)

// IndexOptionsForCredentialsHMAC returns required Index options for KongCredentialHMAC.
func IndexOptionsForCredentialsHMAC() []ReconciliationIndexOption {
	return []ReconciliationIndexOption{
		{
			IndexObject:  &configurationv1alpha1.KongCredentialHMAC{},
			IndexField:   IndexFieldKongCredentialHMACReferencesKongConsumer,
			ExtractValue: kongCredentialHMACReferencesConsumer,
		},
		{
			IndexObject:  &configurationv1alpha1.KongCredentialHMAC{},
			IndexField:   IndexFieldKongCredentialHMACReferencesSecret,
			ExtractValue: kongCredentialReferencesSecret[configurationv1alpha1.KongCredentialHMAC],
		},
	}
}

// kongCredentialHMACReferencesConsumer returns the name of referenced Consumer.
func kongCredentialHMACReferencesConsumer(obj client.Object) []string {
	cred, ok := obj.(*configurationv1alpha1.KongCredentialHMAC)
	if !ok {
		return nil
	}
	return []string{cred.Spec.ConsumerRef.Name}
}
