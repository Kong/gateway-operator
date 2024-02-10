package resources_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"

	kgoerrors "github.com/kong/gateway-operator/internal/errors"
	"github.com/kong/gateway-operator/internal/versions"
	"github.com/kong/gateway-operator/pkg/utils/kubernetes/resources"
	"github.com/kong/gateway-operator/pkg/utils/kubernetes/resources/clusterroles"
)

func TestClusterroleHelpers(t *testing.T) {
	testCases := []struct {
		controlplane        string
		image               string
		expectedClusterRole func() *rbacv1.ClusterRole
		expectedError       error
	}{
		{
			controlplane: "test_3.0",
			image:        "kong/kubernetes-ingress-controller:3.0",
			expectedClusterRole: func() *rbacv1.ClusterRole {
				cr := clusterroles.GenerateNewClusterRoleForControlPlane_ge3_0("test_3.0")
				resources.LabelObjectAsControlPlaneManaged(cr)
				return cr
			},
		},
		{
			controlplane: "test_2.12",
			image:        "kong/kubernetes-ingress-controller:2.12",
			expectedClusterRole: func() *rbacv1.ClusterRole {
				cr := clusterroles.GenerateNewClusterRoleForControlPlane_lt3_0_ge2_12("test_2.12")
				resources.LabelObjectAsControlPlaneManaged(cr)
				return cr
			},
		},
		{
			controlplane: "test_2.11",
			image:        "kong/kubernetes-ingress-controller:2.11",
			expectedClusterRole: func() *rbacv1.ClusterRole {
				cr := clusterroles.GenerateNewClusterRoleForControlPlane_lt2_12_ge2_11("test_2.11")
				resources.LabelObjectAsControlPlaneManaged(cr)
				return cr
			},
		},
		{
			controlplane: "test_development_untagged",
			image:        "test/development",
			expectedClusterRole: func() *rbacv1.ClusterRole {
				cr := clusterroles.GenerateNewClusterRoleForControlPlane_lt3_0_ge2_12("test_development_untagged")
				resources.LabelObjectAsControlPlaneManaged(cr)
				return cr
			},
			expectedError: versions.ErrExpectedSemverVersion,
		},
		{
			controlplane: "test_empty",
			image:        "kong/kubernetes-ingress-controller",
			expectedClusterRole: func() *rbacv1.ClusterRole {
				cr := clusterroles.GenerateNewClusterRoleForControlPlane_lt3_0_ge2_12("test_empty")
				resources.LabelObjectAsControlPlaneManaged(cr)
				return cr
			},
			expectedError: versions.ErrExpectedSemverVersion,
		},
		{
			// TODO: https://github.com/Kong/gateway-operator/issues/1029
			controlplane: "test_unsupported",
			image:        "kong/kubernetes-ingress-controller:1.0",
			expectedClusterRole: func() *rbacv1.ClusterRole {
				cr := clusterroles.GenerateNewClusterRoleForControlPlane_ge3_0("test_unsupported")
				resources.LabelObjectAsControlPlaneManaged(cr)
				return cr
			},
		},
		{
			controlplane:  "test_invalid_tag",
			image:         "test/development:main",
			expectedError: kgoerrors.ErrInvalidSemverVersion,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.controlplane, func(t *testing.T) {
			clusterRole, err := resources.GenerateNewClusterRoleForControlPlane(tc.controlplane, tc.image)
			if tc.expectedError != nil {
				require.ErrorIs(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedClusterRole(), clusterRole)
			}
		})
	}
}