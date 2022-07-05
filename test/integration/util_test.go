//go:build integration_tests
// +build integration_tests

package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kong/kubernetes-testing-framework/pkg/clusters"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
	"github.com/kong/gateway-operator/internal/consts"
	k8sutils "github.com/kong/gateway-operator/internal/utils/kubernetes"
)

// setup is a helper function for tests which conveniently creates a cluster
// cleaner (to clean up test resources automatically after the test finishes)
// and creates a new namespace for the test to use. It also enables parallel
// testing.
func setup(t *testing.T) (*corev1.Namespace, *clusters.Cleaner) {
	t.Log("performing test setup")
	t.Parallel()
	cleaner := clusters.NewCleaner(env.Cluster())

	t.Log("creating a testing namespace")
	namespace, err := k8sClient.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: uuid.NewString(),
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.AddNamespace(namespace)

	return namespace, cleaner
}

// mustListDataPlaneDeployments is a helper function for tests that
// conveniently lists all deployments managed by a given dataplane.
func mustListDataPlaneDeployments(t *testing.T, dataplane *operatorv1alpha1.DataPlane) []v1.Deployment {
	deployments, err := k8sutils.ListDeploymentsForOwner(
		ctx,
		mgrClient,
		consts.GatewayOperatorControlledLabel,
		consts.DataPlaneManagedLabelValue,
		dataplane.Namespace,
		dataplane.UID,
	)
	require.NoError(t, err)
	return deployments
}

// mustListControlPlaneDeployments is a helper function for tests that
// conveniently lists all deployments managed by a given controlplane.
func mustListControlPlaneDeployments(t *testing.T, controlplane *operatorv1alpha1.ControlPlane) []v1.Deployment {
	deployments, err := k8sutils.ListDeploymentsForOwner(
		ctx,
		mgrClient,
		consts.GatewayOperatorControlledLabel,
		consts.ControlPlaneManagedLabelValue,
		controlplane.Namespace,
		controlplane.UID,
	)
	require.NoError(t, err)
	return deployments
}

// mustListServices is a helper function for tests that
// conveniently lists all services managed by a given dataplane.
func mustListDataPlaneServices(t *testing.T, dataplane *operatorv1alpha1.DataPlane) []corev1.Service {
	services, err := k8sutils.ListServicesForOwner(
		ctx,
		mgrClient,
		consts.GatewayOperatorControlledLabel,
		consts.DataPlaneManagedLabelValue,
		dataplane.Namespace,
		dataplane.UID,
	)
	require.NoError(t, err)
	return services
}

// controlPlanePredicate is a helper function for tests that returns a function
// that can be used to check if a ControlPlane has a certain state.
func controlPlanePredicate(
	t *testing.T,
	controlplane *operatorv1alpha1.ControlPlane,
	predicate func(controlplane *operatorv1alpha1.ControlPlane) bool,
) func() bool {
	controlplaneClient := operatorClient.V1alpha1().ControlPlanes(controlplane.Namespace)
	return func() bool {
		controlplane, err := controlplaneClient.Get(ctx, controlplane.Name, metav1.GetOptions{})
		require.NoError(t, err)
		return predicate(controlplane)
	}
}

// dataPlanePredicate is a helper function for tests that returns a function
// that can be used to check if a DataPlane has a certain state.
func dataPlanePredicate(
	t *testing.T,
	dataplane *operatorv1alpha1.DataPlane,
	predicate func(dataplane *operatorv1alpha1.DataPlane) bool,
) func() bool {
	dataPlaneClient := operatorClient.V1alpha1().DataPlanes(dataplane.Namespace)
	return func() bool {
		dataplane, err := dataPlaneClient.Get(ctx, dataplane.Name, metav1.GetOptions{})
		require.NoError(t, err)
		return predicate(dataplane)
	}
}
