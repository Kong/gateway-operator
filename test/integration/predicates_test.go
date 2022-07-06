//go:build integration_tests
// +build integration_tests

package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
	"github.com/kong/gateway-operator/controllers"
)

// newControlPlanePredicate is a helper function for tests that returns a function
// that can be used to check if a ControlPlane has a certain state.
func newControlPlanePredicate(
	t *testing.T,
	controlplaneName types.NamespacedName,
	predicate func(controlplane *operatorv1alpha1.ControlPlane) bool,
) func() bool {
	controlplaneClient := operatorClient.V1alpha1().ControlPlanes(controlplaneName.Namespace)
	return func() bool {
		controlplane, err := controlplaneClient.Get(ctx, controlplaneName.Name, metav1.GetOptions{})
		require.NoError(t, err)
		return predicate(controlplane)
	}
}

// TODO (jrsmroz): Use types.NamespacedName
// dataPlanePredicate is a helper function for tests that returns a function
// that can be used to check if a DataPlane has a certain state.
func dataPlanePredicate(
	t *testing.T,
	dataplaneNamespace, dataplaneName string,
	predicate func(dataplane *operatorv1alpha1.DataPlane) bool,
) func() bool {
	dataPlaneClient := operatorClient.V1alpha1().DataPlanes(dataplaneNamespace)
	return func() bool {
		dataplane, err := dataPlaneClient.Get(ctx, dataplaneName, metav1.GetOptions{})
		require.NoError(t, err)
		return predicate(dataplane)
	}
}

func controlPlaneIsScheduled(t *testing.T, controlplane types.NamespacedName) func() bool {
	return newControlPlanePredicate(t, controlplane, func(c *operatorv1alpha1.ControlPlane) bool {
		for _, condition := range c.Status.Conditions {
			if condition.Type == string(controllers.ControlPlaneConditionTypeProvisioned) {
				return true
			}
		}
		return false
	})
}

func controlPlaneDetectedNoDataplane(t *testing.T, controlplane types.NamespacedName) func() bool {
	return newControlPlanePredicate(t, controlplane, func(c *operatorv1alpha1.ControlPlane) bool {
		for _, condition := range c.Status.Conditions {
			if condition.Type == string(controllers.ControlPlaneConditionTypeProvisioned) &&
				condition.Status == metav1.ConditionFalse &&
				condition.Reason == controllers.ControlPlaneConditionReasonNoDataplane {
				return true
			}
		}
		return false
	})
}

func controlPlaneIsProvisioned(t *testing.T, controlplane types.NamespacedName) func() bool {
	return newControlPlanePredicate(t, controlplane, func(c *operatorv1alpha1.ControlPlane) bool {
		for _, condition := range c.Status.Conditions {
			if condition.Type == string(controllers.ControlPlaneConditionTypeProvisioned) &&
				condition.Status == metav1.ConditionTrue {
				return true
			}
		}
		return false
	})
}

func controlPlaneHasDeployment(t *testing.T, controlplane types.NamespacedName) func() bool {
	return newControlPlanePredicate(t, controlplane, func(c *operatorv1alpha1.ControlPlane) bool {
		deployments := mustListControlPlaneDeployments(t, c)
		return len(deployments) == 1
	})
}

func controlPlaneHasActiveReplicasDeployment(t *testing.T, controlplane types.NamespacedName) func() bool {
	return newControlPlanePredicate(t, controlplane, func(c *operatorv1alpha1.ControlPlane) bool {
		deployments := mustListControlPlaneDeployments(t, c)
		return len(deployments) == 1 &&
			*deployments[0].Spec.Replicas > 0 &&
			deployments[0].Status.AvailableReplicas >= deployments[0].Status.ReadyReplicas
	})
}

func Not(fn func() bool) func() bool {
	return func() bool {
		return !fn()
	}
}
