package konnect

import (
	"context"
	"errors"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getKonnectAPIAuthRefNN(_ context.Context, _ client.Client, ext *konnectv1alpha1.KonnectExtension) (types.NamespacedName, error) {
	// In case the KonnectConfiguration is not set, we fetch the KonnectGatewayControlPlane
	// and get the KonnectConfiguration from there. KonnectGatewayControlPlane reference and KonnectConfiguration
	// are mutually exclusive in the KonnectExtension API.
	if ext.Spec.KonnectConfiguration == nil {
		// TODO: https://github.com/Kong/gateway-operator/issues/889
		return types.NamespacedName{}, errors.New("KonnectGatewayControlPlane references not supported yet")
	}

	// TODO: handle cross namespace refs
	return types.NamespacedName{
		Namespace: ext.Namespace,
		Name:      ext.Spec.KonnectConfiguration.APIAuthConfigurationRef.Name,
	}, nil
}

func getCertificateSecret(ctx context.Context, cl client.Client, ext konnectv1alpha1.KonnectExtension) (*corev1.Secret, error) {
	var certificateSecret corev1.Secret
	switch *ext.Spec.DataPlaneClientAuth.CertificateSecret.Provisioning {
	case konnectv1alpha1.ManualSecretProvisioning:
		// No need to check CertificateSecretRef is nil, as it is enforced at the CRD level.
		if err := cl.Get(ctx, types.NamespacedName{
			Namespace: ext.Namespace,
			Name:      ext.Spec.DataPlaneClientAuth.CertificateSecret.CertificateSecretRef.Name,
		}, &certificateSecret); err != nil {
			return &certificateSecret, err
		}
	default:
		return nil, errors.New("automatic secret provisioning not supported yet")
	}
	return &certificateSecret, nil
}

func konnectClusterTypeToCRDClusterType(clusterType sdkkonnectcomp.ControlPlaneClusterType) konnectv1alpha1.KonnectExtensionClusterType {
	switch clusterType {
	case sdkkonnectcomp.ControlPlaneClusterTypeClusterTypeControlPlane:
		return konnectv1alpha1.ClusterTypeControlPlane
	case sdkkonnectcomp.ControlPlaneClusterTypeClusterTypeK8SIngressController:
		return konnectv1alpha1.ClusterTypeK8sIngressController
	default:
		// default never happens as the validation is at the CRD level
		return ""
	}
}
