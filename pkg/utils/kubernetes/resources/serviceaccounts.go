package resources

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kong/gateway-operator/pkg/consts"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
)

// -----------------------------------------------------------------------------
// ServiceAccount generators
// -----------------------------------------------------------------------------

// GenerateNewServiceAccountForControlPlane is a helper to generate a ServiceAccount
// to be used by the controlplane deployment.
func GenerateNewServiceAccountForControlPlane(namespace, controlplaneName string) *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: k8sutils.TrimGenerateName(fmt.Sprintf("%s-%s-", consts.ControlPlanePrefix, controlplaneName)),
			Namespace:    namespace,
			Labels: map[string]string{
				"app": controlplaneName,
			},
		},
	}
	LabelObjectAsControlPlaneManaged(sa)

	return sa
}

// GenerateNewServiceAccountForCertificateConfig is a helper to generate a ServiceAccount
// to be used by the certificateConfig jobs
func GenerateNewServiceAccountForCertificateConfig(namespace, name, labelValue string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": labelValue,
			},
		},
	}
}
