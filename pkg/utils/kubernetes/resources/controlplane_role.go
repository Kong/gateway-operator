package resources

import (
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
)

func GenerateNewRoleForControlPlane(controlplaneName string, namespace string, rules []rbacv1.PolicyRule) *rbacv1.Role {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: k8sutils.TrimGenerateName(fmt.Sprintf("%s-", controlplaneName)),
			Namespace:    namespace,
			Labels: map[string]string{
				"app": controlplaneName,
			},
		},
		Rules: rules,
	}

	LabelObjectAsControlPlaneManaged(role)
	return role
}
