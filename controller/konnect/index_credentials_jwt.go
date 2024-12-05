package konnect

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

const (
	// IndexFieldKongCredentialJWTReferencesKongConsumer is the index name for KongCredentialJWT -> Consumer.
	IndexFieldKongCredentialJWTReferencesKongConsumer = "kongCredentialsJWTConsumerRef"
)

// IndexOptionsForCredentialsJWT returns required Index options for KongCredentialJWT.
func IndexOptionsForCredentialsJWT() []ReconciliationIndexOption {
	return []ReconciliationIndexOption{
		{
			IndexObject:  &configurationv1alpha1.KongCredentialJWT{},
			IndexField:   IndexFieldKongCredentialJWTReferencesKongConsumer,
			ExtractValue: kongCredentialJWTReferencesConsumer,
		},
	}
}

// kongCredentialJWTReferencesConsumer returns the name of referenced Consumer.
func kongCredentialJWTReferencesConsumer(obj client.Object) []string {
	cred, ok := obj.(*configurationv1alpha1.KongCredentialJWT)
	if !ok {
		return nil
	}
	return []string{cred.Spec.ConsumerRef.Name}
}
