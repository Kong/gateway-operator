//go:build integration_tests
// +build integration_tests

package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kong/gateway-operator/api/v1alpha1"
	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
)

func TestControlPlaneEssentials(t *testing.T) {
	namespace, cleaner := setup(t)
	defer func() { assert.NoError(t, cleaner.Cleanup(ctx)) }()

	dataplaneClient := operatorClient.V1alpha1().DataPlanes(namespace.Name)
	controlplaneClient := operatorClient.V1alpha1().ControlPlanes(namespace.Name)

	// Control plane needs a dataplane to exist to properly function.
	dataplane := &v1alpha1.DataPlane{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace.Name,
			Name:      uuid.NewString(),
		},
	}

	controlplaneName := types.NamespacedName{
		Namespace: namespace.Name,
		Name:      uuid.NewString(),
	}
	controlplane := &operatorv1alpha1.ControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: controlplaneName.Namespace,
			Name:      controlplaneName.Name,
		},
		Spec: operatorv1alpha1.ControlPlaneSpec{
			ControlPlaneDeploymentOptions: operatorv1alpha1.ControlPlaneDeploymentOptions{
				DataPlane: &dataplane.Name,
			},
		},
	}

	t.Log("deploying dataplane resource")
	dataplane, err := dataplaneClient.Create(ctx, dataplane, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(dataplane)

	t.Log("verifying deployments managed by the dataplane are ready")
	require.Eventually(t, func() bool {
		deployments := mustListDataPlaneDeployments(t, dataplane)
		return len(deployments) == 1 &&
			deployments[0].Status.AvailableReplicas >= deployments[0].Status.ReadyReplicas
	}, time.Minute, time.Second)

	t.Log("verifying services managed by the dataplane")
	require.Eventually(t, func() bool {
		services := mustListDataPlaneServices(t, dataplane)
		if len(services) == 1 {
			return true
		}
		return false
	}, 2*time.Minute, time.Second)

	t.Log("deploying controlplane resource")
	controlplane, err = controlplaneClient.Create(ctx, controlplane, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(controlplane)

	t.Log("verifying controlplane gets marked scheduled")
	require.Eventually(t, controlPlaneIsScheduled(t, controlplaneName), 2*time.Minute, time.Second)

	t.Log("verifying that the controlplane gets marked as provisioned")
	require.Eventually(t, controlPlaneIsProvisioned(t, controlplaneName), 2*time.Minute, time.Second)

	t.Log("verifying controlplane deployment has active replicas")
	require.Eventually(t, controlPlaneHasActiveReplicasDeployment(t, controlplaneName), 2*time.Minute, time.Second)
}

func TestDormantControlplane(t *testing.T) {
	namespace, cleaner := setup(t)
	defer func() { assert.NoError(t, cleaner.Cleanup(ctx)) }()

	dataplaneClient := operatorClient.V1alpha1().DataPlanes(namespace.Name)
	controlplaneClient := operatorClient.V1alpha1().ControlPlanes(namespace.Name)

	controlplaneName := types.NamespacedName{
		Namespace: namespace.Name,
		Name:      uuid.NewString(),
	}
	controlplane := &operatorv1alpha1.ControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: controlplaneName.Namespace,
			Name:      controlplaneName.Name,
		},
		Spec: operatorv1alpha1.ControlPlaneSpec{
			ControlPlaneDeploymentOptions: operatorv1alpha1.ControlPlaneDeploymentOptions{
				DeploymentOptions: operatorv1alpha1.DeploymentOptions{},
				DataPlane:         nil,
			},
		},
	}

	dataplane := &v1alpha1.DataPlane{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace.Name,
			Name:      uuid.NewString(),
		},
	}

	t.Log("deploying controlplane resource without dataplane attached")
	controlplane, err := controlplaneClient.Create(ctx, controlplane, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(controlplane)

	t.Log("verifying controlplane state reflects lack of dataplane")
	require.Eventually(t, controlPlaneDetectedNoDataplane(t, controlplaneName), time.Minute, time.Second)

	t.Log("verifying controlplane deployment has no active replicas")
	require.Eventually(t, Not(controlPlaneHasActiveReplicasDeployment(t, controlplaneName)), time.Minute, time.Second)

	t.Log("deploying dataplane resource")
	dataplane, err = dataplaneClient.Create(ctx, dataplane, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(dataplane)

	t.Log("verifying deployments managed by the dataplane are ready")
	require.Eventually(t, func() bool {
		deployments := mustListDataPlaneDeployments(t, dataplane)
		return len(deployments) == 1 &&
			deployments[0].Status.AvailableReplicas >= deployments[0].Status.ReadyReplicas
	}, time.Minute, time.Second)

	t.Log("verifying services managed by the dataplane")
	// var dataplaneService *corev1.Service
	require.Eventually(t, func() bool {
		services := mustListDataPlaneServices(t, dataplane)
		if len(services) == 1 {
			// dataplaneService = &services[0]
			return true
		}
		return false
	}, time.Minute, time.Second)

	t.Log("attaching dataplane to controlplane")
	controlplane, err = controlplaneClient.Get(ctx, controlplane.Name, metav1.GetOptions{})
	require.NoError(t, err)
	controlplane.Spec.DataPlane = &dataplane.Name
	controlplane, err = controlplaneClient.Update(ctx, controlplane, metav1.UpdateOptions{})
	require.NoError(t, err)

	t.Log("verifying controlplane is now provisioned")
	require.Eventually(t, controlPlaneIsProvisioned(t, controlplaneName), 2*time.Minute, time.Second)

	t.Log("verifying controlplane deployment has active replicas")
	require.Eventually(t, controlPlaneHasActiveReplicasDeployment(t, controlplaneName), 2*time.Minute, time.Second)

	t.Log("removing dataplane from controlplane")
	require.Eventually(t, func() bool {
		controlplane, err = controlplaneClient.Get(ctx, controlplane.Name, metav1.GetOptions{})
		require.NoError(t, err)
		controlplane.Spec.DataPlane = nil
		controlplane, err = controlplaneClient.Update(ctx, controlplane, metav1.UpdateOptions{})
		require.NoError(t, err)
		return true
	}, 2*time.Minute, time.Second)

	t.Log("verifying controlplane deployment has no active replicas")
	require.Eventually(t, Not(controlPlaneHasActiveReplicasDeployment(t, controlplaneName)), 2*time.Minute, time.Second)
}
