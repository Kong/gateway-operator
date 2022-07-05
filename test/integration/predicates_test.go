//go:build integration_tests
// +build integration_tests

package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
)

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
