//go:build integration_tests
// +build integration_tests

package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kong/kubernetes-testing-framework/pkg/clusters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kong/gateway-operator/api/v1alpha1"
)

func TestDataplaneEssentials(t *testing.T) {
	t.Log("setting up cleanup")
	cleaner := clusters.NewCleaner(env.Cluster())
	defer func() { assert.NoError(t, cleaner.Cleanup(ctx)) }()

	t.Log("creating a testing namespace")
	namespace, err := k8sClient.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: uuid.NewString(),
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.AddNamespace(namespace)

	t.Log("deploying dataplane resource")
	dataplane := &v1alpha1.DataPlane{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace.Name,
			Name:      uuid.NewString(),
		},
	}
	dataplane, err = operatorClient.V1alpha1().DataPlanes(namespace.Name).Create(ctx, dataplane, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(dataplane)

	t.Log("waiting for dataplane deployment")
	require.Eventually(t, func() bool {
		deployment, err := k8sClient.AppsV1().Deployments(namespace.Name).Get(ctx, dataplane.Name, metav1.GetOptions{})
		return err == nil && deployment.Status.ReadyReplicas == deployment.Status.Replicas
	}, time.Minute, time.Second)
}
