package konnect

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

const (
	// IndexFieldKongDataPlaneClientCertificateOnKonnectGatewayControlPlane is the index field for KongDataPlaneCertificate -> KonnectGatewayControlPlane.
	IndexFieldKongDataPlaneClientCertificateOnKonnectGatewayControlPlane = "dataPlaneCertificateKonnectGatewayControlPlaneRef"
)

// IndexOptionsForKongDataPlaneCertificate returns required Index options for KongConsumer reconciler.
func IndexOptionsForKongDataPlaneCertificate() []ReconciliationIndexOption {
	return []ReconciliationIndexOption{
		{
			IndexObject:  &configurationv1alpha1.KongDataPlaneClientCertificate{},
			IndexField:   IndexFieldKongDataPlaneClientCertificateOnKonnectGatewayControlPlane,
			ExtractValue: kongDataPlaneCertificateReferencesKonnectGatewayControlPlane,
		},
	}
}

func kongDataPlaneCertificateReferencesKonnectGatewayControlPlane(object client.Object) []string {
	dpCert, ok := object.(*configurationv1alpha1.KongDataPlaneClientCertificate)
	if !ok {
		return nil
	}

	return controlPlaneKonnectNamespacedRefAsSlice(dpCert)
}
