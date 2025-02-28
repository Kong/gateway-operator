package resources

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonv1alpha1 "github.com/kong/kubernetes-configuration/api/common/v1alpha1"
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

// GenerateKongDataPlaneClientCertificate generates a KongDataPlaneClientCertificate object setting
// the provided controlPlaneRef as the certificate controlPlaneRef. The cert parameter is the actual certificate
// pushed into Konnect.
func GenerateKongDataPlaneClientCertificate(name, namespace string, controlPlaneRef *commonv1alpha1.ControlPlaneRef, cert string, opts ...func(dpCert *configurationv1alpha1.KongDataPlaneClientCertificate)) configurationv1alpha1.KongDataPlaneClientCertificate {
	dpCert := configurationv1alpha1.KongDataPlaneClientCertificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: configurationv1alpha1.KongDataPlaneClientCertificateSpec{
			KongDataPlaneClientCertificateAPISpec: configurationv1alpha1.KongDataPlaneClientCertificateAPISpec{
				Cert: cert,
			},
			ControlPlaneRef: controlPlaneRef,
		},
	}

	for _, opt := range opts {
		opt(&dpCert)
	}

	return dpCert
}
