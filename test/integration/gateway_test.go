//go:build integration_tests
// +build integration_tests

package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/kong/gateway-operator/pkg/vars"
)

var (
	// gatewaySchedulingTimeLimit is the maximum amount of time to wait for
	// a supported Gateway to be marked as Scheduled by the gateway controller.
	gatewaySchedulingTimeLimit = time.Second * 7

	// gatewayReadyTimeLimit is the maximum amount of time to wait for a
	// supported Gateway to be fully provisioned and marked as Ready by the
	// gateway controller.
	gatewayReadyTimeLimit = time.Minute * 2
)

func TestGatewayEssentials(t *testing.T) {
	namespace, cleaner := setup(t)
	defer func() { assert.NoError(t, cleaner.Cleanup(ctx)) }()

	t.Log("deploying a GatewayClass resource")
	gatewayClass := &gatewayv1alpha2.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: uuid.NewString(),
		},
		Spec: gatewayv1alpha2.GatewayClassSpec{
			ControllerName: gatewayv1alpha2.GatewayController(vars.ControllerName),
		},
	}
	gatewayClass, err := gatewayClient.GatewayV1alpha2().GatewayClasses().Create(ctx, gatewayClass, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gatewayClass)

	t.Log("deploying Gateway resource")
	gatewayNSN := types.NamespacedName{
		Name:      uuid.NewString(),
		Namespace: namespace.Name,
	}
	gateway := &gatewayv1alpha2.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: gatewayNSN.Namespace,
			Name:      gatewayNSN.Name,
		},
		Spec: gatewayv1alpha2.GatewaySpec{
			GatewayClassName: gatewayv1alpha2.ObjectName(gatewayClass.Name),
			Listeners: []gatewayv1alpha2.Listener{{
				Name:     "http",
				Protocol: gatewayv1alpha2.HTTPProtocolType,
				Port:     gatewayv1alpha2.PortNumber(80),
			}},
		},
	}
	gateway, err = gatewayClient.GatewayV1alpha2().Gateways(namespace.Name).Create(ctx, gateway, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gateway)

	t.Log("verifying Gateway gets marked as Scheduled")
	require.Eventually(t, gatewayIsScheduled(t, ctx, gatewayNSN), gatewaySchedulingTimeLimit, time.Second)

	t.Log("verifying Gateway gets marked as Ready")
	require.Eventually(t, gatewayIsReady(t, ctx, gatewayNSN), gatewayReadyTimeLimit, time.Second)

	t.Log("verifying Gateway gets an IP address")
	require.Eventually(t, gatewayIpAddressExist(t, ctx, gatewayNSN), subresourceReadinessWait, time.Second)
	gateway = mustGetGateway(t, gatewayNSN)
	gatewayIPAddress := gateway.Status.Addresses[0].Value

	t.Log("verifying that the DataPlane becomes provisioned")
	require.Eventually(t, gatewayDataPlaneIsProvisioned(t, gateway), subresourceReadinessWait, time.Second)
	dataplane := mustListDataPlanesForGateway(t, gateway)[0]

	t.Log("verifying that the ControlPlane becomes provisioned")
	require.Eventually(t, gatewayControlPlaneIsProvisioned(t, gateway), subresourceReadinessWait, time.Second)
	controlplane := mustListControlPlanesForGateway(t, gateway)[0]

	t.Log("verifying networkpolicies are created")
	require.Eventually(t, gatewayNetworkPoliciesExist(t, ctx, gateway), subresourceReadinessWait, time.Second)

	t.Log("verifying connectivity to the Gateway")

	require.Eventually(t, getResponseBodyContains(t, ctx, "http://"+gatewayIPAddress, defaultKongResponseBody), subresourceReadinessWait, time.Second)

	t.Log("deleting Gateway resource")
	require.NoError(t, gatewayClient.GatewayV1alpha2().Gateways(namespace.Name).Delete(ctx, gateway.Name, metav1.DeleteOptions{}))

	t.Log("verifying that DataPlane sub-resources are deleted")
	assert.Eventually(t, func() bool {
		_, err := operatorClient.ApisV1alpha1().DataPlanes(namespace.Name).Get(ctx, dataplane.Name, metav1.GetOptions{})
		return errors.IsNotFound(err)
	}, time.Minute, time.Second)

	t.Log("verifying that ControlPlane sub-resources are deleted")
	assert.Eventually(t, func() bool {
		_, err := operatorClient.ApisV1alpha1().ControlPlanes(namespace.Name).Get(ctx, controlplane.Name, metav1.GetOptions{})
		return errors.IsNotFound(err)
	}, time.Minute, time.Second)

	t.Log("verifying networkpolicies are deleted")
	require.Eventually(t, Not(gatewayNetworkPoliciesExist(t, ctx, gateway)), time.Minute, time.Second)
}
