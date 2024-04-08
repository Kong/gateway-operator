package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kong/kubernetes-testing-framework/pkg/utils/kubernetes/generators"
	"github.com/kr/pretty"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	operatorv1beta1 "github.com/kong/gateway-operator/apis/v1beta1"
	"github.com/kong/gateway-operator/pkg/consts"
	gatewayutils "github.com/kong/gateway-operator/pkg/utils/gateway"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
	testutils "github.com/kong/gateway-operator/pkg/utils/test"
	"github.com/kong/gateway-operator/test/helpers"
)

func TestGatewayEssentials(t *testing.T) {
	t.Parallel()
	namespace, cleaner := helpers.SetupTestEnv(t, GetCtx(), GetEnv())

	t.Log("deploying a GatewayClass resource")
	gatewayClass := testutils.GenerateGatewayClass()
	gatewayClass, err := GetClients().GatewayClient.GatewayV1().GatewayClasses().Create(GetCtx(), gatewayClass, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gatewayClass)

	t.Log("deploying Gateway resource")
	gatewayNN := types.NamespacedName{
		Name:      uuid.NewString(),
		Namespace: namespace.Name,
	}
	gateway := testutils.GenerateGateway(gatewayNN, gatewayClass)
	gateway, err = GetClients().GatewayClient.GatewayV1().Gateways(namespace.Name).Create(GetCtx(), gateway, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gateway)

	t.Log("verifying Gateway gets marked as Scheduled")
	require.Eventually(t, testutils.GatewayIsScheduled(t, GetCtx(), gatewayNN, clients), testutils.GatewaySchedulingTimeLimit, time.Second)

	t.Log("verifying Gateway gets marked as Programmed")
	require.Eventually(t, testutils.GatewayIsProgrammed(t, GetCtx(), gatewayNN, clients), testutils.GatewayReadyTimeLimit, time.Second)
	require.Eventually(t, testutils.GatewayListenersAreProgrammed(t, GetCtx(), gatewayNN, clients), testutils.GatewayReadyTimeLimit, time.Second)

	t.Log("verifying Gateway gets an IP address")
	require.Eventually(t, testutils.GatewayIPAddressExist(t, GetCtx(), gatewayNN, clients), testutils.SubresourceReadinessWait, time.Second)
	gateway = testutils.MustGetGateway(t, GetCtx(), gatewayNN, clients)
	gatewayIPAddress := gateway.Status.Addresses[0].Value

	t.Log("verifying that the DataPlane becomes Ready")
	require.Eventually(t, testutils.GatewayDataPlaneIsReady(t, GetCtx(), gateway, clients), testutils.SubresourceReadinessWait, time.Second)
	dataplanes := testutils.MustListDataPlanesForGateway(t, GetCtx(), gateway, clients)
	require.Len(t, dataplanes, 1)
	dataplane := dataplanes[0]

	t.Log("verifying that the ControlPlane becomes provisioned")
	require.Eventually(t, testutils.GatewayControlPlaneIsProvisioned(t, GetCtx(), gateway, clients), testutils.SubresourceReadinessWait, time.Second)
	controlplanes := testutils.MustListControlPlanesForGateway(t, GetCtx(), gateway, clients)
	require.Len(t, controlplanes, 1)
	controlplane := controlplanes[0]

	t.Log("verifying networkpolicies are created")
	require.Eventually(t, testutils.GatewayNetworkPoliciesExist(t, GetCtx(), gateway, clients), testutils.SubresourceReadinessWait, time.Second)

	t.Log("verifying connectivity to the Gateway")
	require.Eventually(t, Expect404WithNoRouteFunc(t, GetCtx(), "http://"+gatewayIPAddress), testutils.SubresourceReadinessWait, time.Second)

	dataplaneClient := GetClients().OperatorClient.ApisV1beta1().DataPlanes(namespace.Name)
	dataplaneNN := types.NamespacedName{Namespace: namespace.Name, Name: dataplane.Name}
	controlplaneClient := GetClients().OperatorClient.ApisV1beta1().ControlPlanes(namespace.Name)
	controlplaneNN := types.NamespacedName{Namespace: namespace.Name, Name: controlplane.Name}

	t.Log("verifying that dataplane has 1 ready replica")
	require.Eventually(t, testutils.DataPlaneHasNReadyPods(t, GetCtx(), dataplaneNN, clients, 1), time.Minute, time.Second)

	t.Log("verifying that controlplane has 1 ready replica")
	require.Eventually(t, testutils.ControlPlaneHasNReadyPods(t, GetCtx(), controlplaneNN, clients, 1), time.Minute, time.Second)

	t.Log("deleting controlplane")
	require.NoError(t, controlplaneClient.Delete(GetCtx(), controlplane.Name, metav1.DeleteOptions{}))

	t.Log("deleting dataplane")
	require.NoError(t, dataplaneClient.Delete(GetCtx(), dataplane.Name, metav1.DeleteOptions{}))

	t.Log("verifying Gateway gets marked as not Programmed")
	require.Eventually(t, testutils.Not(testutils.GatewayIsProgrammed(t, GetCtx(), gatewayNN, clients)), testutils.GatewayReadyTimeLimit, time.Second)
	require.Eventually(t, testutils.GatewayListenersAreProgrammed(t, GetCtx(), gatewayNN, clients), testutils.GatewayReadyTimeLimit, time.Second)

	t.Log("verifying that the ControlPlane becomes provisioned again")
	require.Eventually(t, testutils.GatewayControlPlaneIsProvisioned(t, GetCtx(), gateway, clients), 45*time.Second, time.Second)
	controlplanes = testutils.MustListControlPlanesForGateway(t, GetCtx(), gateway, clients)
	require.Len(t, controlplanes, 1)
	controlplane = controlplanes[0]

	t.Log("verifying that the DataPlane becomes provisioned again")
	require.Eventually(t, testutils.GatewayDataPlaneIsReady(t, GetCtx(), gateway, clients), 45*time.Second, time.Second)
	dataplanes = testutils.MustListDataPlanesForGateway(t, GetCtx(), gateway, clients)
	require.Len(t, dataplanes, 1)
	dataplane = dataplanes[0]

	t.Log("verifying Gateway gets marked as Programmed again")
	require.Eventually(t, testutils.GatewayIsProgrammed(t, GetCtx(), gatewayNN, clients), testutils.GatewayReadyTimeLimit, time.Second)
	require.Eventually(t, testutils.GatewayListenersAreProgrammed(t, GetCtx(), gatewayNN, clients), testutils.GatewayReadyTimeLimit, time.Second)

	t.Log("verifying Gateway gets an IP address again")
	require.Eventually(t, testutils.GatewayIPAddressExist(t, GetCtx(), gatewayNN, clients), testutils.SubresourceReadinessWait, time.Second)
	gateway = testutils.MustGetGateway(t, GetCtx(), gatewayNN, clients)
	gatewayIPAddress = gateway.Status.Addresses[0].Value

	t.Log("verifying connectivity to the Gateway")
	require.Eventually(t, Expect404WithNoRouteFunc(t, GetCtx(), "http://"+gatewayIPAddress), testutils.SubresourceReadinessWait, time.Second)

	t.Log("verifying services managed by the dataplane")
	var dataplaneService corev1.Service
	dataplaneName := types.NamespacedName{
		Namespace: dataplane.Namespace,
		Name:      dataplane.Name,
	}
	require.Eventually(t, testutils.DataPlaneHasActiveService(t, GetCtx(), dataplaneName, &dataplaneService, clients, client.MatchingLabels{
		consts.GatewayOperatorManagedByLabel: consts.DataPlaneManagedLabelValue,
		consts.DataPlaneServiceTypeLabel:     string(consts.DataPlaneIngressServiceLabelValue),
	}), time.Minute, time.Second)

	t.Log("deleting the dataplane service")
	require.NoError(t, GetClients().MgrClient.Delete(GetCtx(), &dataplaneService))

	t.Log("verifying services managed by the dataplane after deletion")
	require.Eventually(t, testutils.DataPlaneHasActiveService(t, GetCtx(), dataplaneName, &dataplaneService, clients, client.MatchingLabels{
		consts.GatewayOperatorManagedByLabel: consts.DataPlaneManagedLabelValue,
		consts.DataPlaneServiceTypeLabel:     string(consts.DataPlaneIngressServiceLabelValue),
	}), time.Minute, time.Second)
	services := testutils.MustListDataPlaneServices(t, GetCtx(), &dataplane, GetClients().MgrClient, client.MatchingLabels{
		consts.GatewayOperatorManagedByLabel: consts.DataPlaneManagedLabelValue,
		consts.DataPlaneServiceTypeLabel:     string(consts.DataPlaneIngressServiceLabelValue),
	})
	require.Len(t, services, 1)
	service := services[0]

	t.Log("verifying controlplane deployment updated with new dataplane service")
	require.Eventually(t, func() bool {
		controlDeployment := testutils.MustListControlPlaneDeployments(t, GetCtx(), &controlplane, clients)[0]
		container := k8sutils.GetPodContainerByName(&controlDeployment.Spec.Template.Spec, consts.ControlPlaneControllerContainerName)
		if container == nil {
			return false
		}
		for _, envvar := range container.Env {
			if envvar.Name == "CONTROLLER_PUBLISH_SERVICE" {
				return envvar.Value == fmt.Sprintf("%s/%s", service.Namespace, service.Name)
			}
		}
		return false
	}, time.Minute*2, time.Second)

	t.Log("deleting Gateway resource")
	require.NoError(t, GetClients().GatewayClient.GatewayV1().Gateways(namespace.Name).Delete(GetCtx(), gateway.Name, metav1.DeleteOptions{}))

	t.Log("verifying that DataPlane sub-resources are deleted")
	assert.Eventually(t, func() bool {
		_, err := GetClients().OperatorClient.ApisV1beta1().DataPlanes(namespace.Name).Get(GetCtx(), dataplane.Name, metav1.GetOptions{})
		return errors.IsNotFound(err)
	}, time.Minute, time.Second)

	t.Log("verifying that ControlPlane sub-resources are deleted")
	assert.Eventually(t, func() bool {
		_, err := GetClients().OperatorClient.ApisV1beta1().ControlPlanes(namespace.Name).Get(GetCtx(), controlplane.Name, metav1.GetOptions{})
		return errors.IsNotFound(err)
	}, time.Minute, time.Second)

	t.Log("verifying networkpolicies are deleted")
	require.Eventually(t, testutils.Not(testutils.GatewayNetworkPoliciesExist(t, GetCtx(), gateway, clients)), time.Minute, time.Second)

	t.Log("verifying that gateway itself is deleted")
	require.Eventually(t, testutils.GatewayNotExist(t, GetCtx(), gatewayNN, clients), time.Minute, time.Second)
}

// TestGatewayMultiple checks essential Gateway behavior with multiple Gateways of the same class. Ensure DataPlanes
// only serve routes attached to their Gateway.
func TestGatewayMultiple(t *testing.T) {
	t.Parallel()
	namespace, cleaner := helpers.SetupTestEnv(t, GetCtx(), GetEnv())
	gatewayV1Client := GetClients().GatewayClient.GatewayV1()

	t.Log("deploying a GatewayClass resource")
	gatewayClass := testutils.GenerateGatewayClass()
	gatewayClass, err := gatewayV1Client.GatewayClasses().Create(GetCtx(), gatewayClass, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gatewayClass)

	t.Log("deploying Gateway resources")
	gatewayOneNN := types.NamespacedName{
		Name:      uuid.NewString(),
		Namespace: namespace.Name,
	}
	gatewayTwoNN := types.NamespacedName{
		Name:      uuid.NewString(),
		Namespace: namespace.Name,
	}
	gatewayOne := testutils.GenerateGateway(gatewayOneNN, gatewayClass)
	gatewayOne, err = gatewayV1Client.Gateways(namespace.Name).Create(GetCtx(), gatewayOne, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gatewayOne)
	gatewayTwo := testutils.GenerateGateway(gatewayTwoNN, gatewayClass)
	gatewayTwo, err = gatewayV1Client.Gateways(namespace.Name).Create(GetCtx(), gatewayTwo, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gatewayTwo)

	t.Log("verifying Gateways marked as Scheduled")
	require.Eventually(t, testutils.GatewayIsScheduled(t, GetCtx(), gatewayOneNN, clients), testutils.GatewaySchedulingTimeLimit, time.Second)
	require.Eventually(t, testutils.GatewayIsScheduled(t, GetCtx(), gatewayTwoNN, clients), testutils.GatewaySchedulingTimeLimit, time.Second)

	t.Log("verifying Gateways marked as Programmed")
	require.Eventually(t, testutils.GatewayIsProgrammed(t, GetCtx(), gatewayOneNN, clients), testutils.GatewayReadyTimeLimit, time.Second)
	require.Eventually(t, testutils.GatewayListenersAreProgrammed(t, GetCtx(), gatewayOneNN, clients), testutils.GatewayReadyTimeLimit, time.Second)
	require.Eventually(t, testutils.GatewayIsProgrammed(t, GetCtx(), gatewayTwoNN, clients), testutils.GatewayReadyTimeLimit, time.Second)
	require.Eventually(t, testutils.GatewayListenersAreProgrammed(t, GetCtx(), gatewayTwoNN, clients), testutils.GatewayReadyTimeLimit, time.Second)

	t.Log("verifying Gateways get an IP address")
	require.Eventually(t, testutils.GatewayIPAddressExist(t, GetCtx(), gatewayOneNN, clients), testutils.SubresourceReadinessWait, time.Second)
	gatewayOne = testutils.MustGetGateway(t, GetCtx(), gatewayOneNN, clients)
	gatewayOneIPAddress := gatewayOne.Status.Addresses[0].Value
	gatewayTwo = testutils.MustGetGateway(t, GetCtx(), gatewayTwoNN, clients)
	gatewayTwoIPAddress := gatewayTwo.Status.Addresses[0].Value

	t.Log("verifying that the DataPlanes become Ready")
	require.Eventually(t, testutils.GatewayDataPlaneIsReady(t, GetCtx(), gatewayOne, clients), testutils.SubresourceReadinessWait, time.Second)
	dataplanesOne := testutils.MustListDataPlanesForGateway(t, GetCtx(), gatewayOne, clients)
	require.Len(t, dataplanesOne, 1)
	dataplaneOne := dataplanesOne[0]
	require.Eventually(t, testutils.GatewayDataPlaneIsReady(t, GetCtx(), gatewayTwo, clients), testutils.SubresourceReadinessWait, time.Second)
	dataplanesTwo := testutils.MustListDataPlanesForGateway(t, GetCtx(), gatewayTwo, clients)
	require.Len(t, dataplanesTwo, 1)
	dataplaneTwo := dataplanesTwo[0]

	t.Log("verifying that the ControlPlanes become provisioned")
	require.Eventually(t, testutils.GatewayControlPlaneIsProvisioned(t, GetCtx(), gatewayOne, clients), testutils.SubresourceReadinessWait, time.Second)
	controlplanesOne := testutils.MustListControlPlanesForGateway(t, GetCtx(), gatewayOne, clients)
	require.Len(t, controlplanesOne, 1)
	controlplaneOne := controlplanesOne[0]
	require.Eventually(t, testutils.GatewayControlPlaneIsProvisioned(t, GetCtx(), gatewayTwo, clients), testutils.SubresourceReadinessWait, time.Second)
	controlplanesTwo := testutils.MustListControlPlanesForGateway(t, GetCtx(), gatewayTwo, clients)
	require.Len(t, controlplanesTwo, 1)
	controlplaneTwo := controlplanesTwo[0]

	dataplaneOneNN := types.NamespacedName{Namespace: namespace.Name, Name: dataplaneOne.Name}
	controlplaneOneNN := types.NamespacedName{Namespace: namespace.Name, Name: controlplaneOne.Name}
	dataplaneTwoNN := types.NamespacedName{Namespace: namespace.Name, Name: dataplaneTwo.Name}
	controlplaneTwoNN := types.NamespacedName{Namespace: namespace.Name, Name: controlplaneTwo.Name}

	t.Log("verifying that dataplanes have 1 ready replica each")
	require.Eventually(t, testutils.DataPlaneHasNReadyPods(t, GetCtx(), dataplaneOneNN, clients, 1), time.Minute, time.Second)
	require.Eventually(t, testutils.DataPlaneHasNReadyPods(t, GetCtx(), dataplaneTwoNN, clients, 1), time.Minute, time.Second)

	t.Log("verifying that controlplanes have 1 ready replica each")
	require.Eventually(t, testutils.ControlPlaneHasNReadyPods(t, GetCtx(), controlplaneOneNN, clients, 1), time.Minute, time.Second)
	require.Eventually(t, testutils.ControlPlaneHasNReadyPods(t, GetCtx(), controlplaneTwoNN, clients, 1), time.Minute, time.Second)

	t.Log("verifying connectivity to the Gateway")
	require.Eventually(t, Expect404WithNoRouteFunc(t, GetCtx(), "http://"+gatewayOneIPAddress), testutils.SubresourceReadinessWait, time.Second)
	require.Eventually(t, Expect404WithNoRouteFunc(t, GetCtx(), "http://"+gatewayTwoIPAddress), testutils.SubresourceReadinessWait, time.Second)

	t.Log("verifying services are managed by their dataplanes")
	var dataplaneOneService corev1.Service
	dataplaneOneName := types.NamespacedName{
		Namespace: dataplaneOne.Namespace,
		Name:      dataplaneOne.Name,
	}
	var dataplaneTwoService corev1.Service
	dataplaneTwoName := types.NamespacedName{
		Namespace: dataplaneTwo.Namespace,
		Name:      dataplaneTwo.Name,
	}

	require.Eventually(t, testutils.DataPlaneHasActiveService(t, GetCtx(), dataplaneOneName, &dataplaneOneService, clients, client.MatchingLabels{
		consts.GatewayOperatorManagedByLabel: consts.DataPlaneManagedLabelValue,
		consts.DataPlaneServiceTypeLabel:     string(consts.DataPlaneIngressServiceLabelValue),
	}), time.Minute, time.Second)
	require.Eventually(t, testutils.DataPlaneHasActiveService(t, GetCtx(), dataplaneTwoName, &dataplaneTwoService, clients, client.MatchingLabels{
		consts.GatewayOperatorManagedByLabel: consts.DataPlaneManagedLabelValue,
		consts.DataPlaneServiceTypeLabel:     string(consts.DataPlaneIngressServiceLabelValue),
	}), time.Minute, time.Second)

	t.Log("deploying backend deployment (httpbin) of HTTPRoute")
	container := generators.NewContainer("httpbin", testutils.HTTPBinImage, 80)
	deployment := generators.NewDeploymentForContainer(container)
	deployment, err = GetEnv().Cluster().Client().AppsV1().Deployments(namespace.Name).Create(GetCtx(), deployment, metav1.CreateOptions{})
	require.NoError(t, err)

	t.Logf("exposing deployment %s via service", deployment.Name)
	service := generators.NewServiceForDeployment(deployment, corev1.ServiceTypeClusterIP)
	_, err = GetEnv().Cluster().Client().CoreV1().Services(namespace.Name).Create(GetCtx(), service, metav1.CreateOptions{})
	require.NoError(t, err)

	t.Logf("creating an httproute to access deployment %s via kong", deployment.Name)

	const pathOne = "/path-test-one"
	httpRouteOne := createHTTPRoute(gatewayOne, service, pathOne)
	require.EventuallyWithT(t,
		func(c *assert.CollectT) {
			result, err := gatewayV1Client.HTTPRoutes(namespace.Name).Create(GetCtx(), httpRouteOne, metav1.CreateOptions{})
			if err != nil {
				t.Logf("failed to deploy httproute: %v", err)
				c.Errorf("failed to deploy httproute: %v", err)
				return
			}
			cleaner.Add(result)
		},
		30*time.Second, time.Second,
	)

	const pathTwo = "/path-test-two"
	httpRouteTwo := createHTTPRoute(gatewayTwo, service, pathTwo)
	require.EventuallyWithT(t,
		func(c *assert.CollectT) {
			result, err := gatewayV1Client.HTTPRoutes(namespace.Name).Create(GetCtx(), httpRouteTwo, metav1.CreateOptions{})
			if err != nil {
				t.Logf("failed to deploy httproute: %v", err)
				c.Errorf("failed to deploy httproute: %v", err)
				return
			}
			cleaner.Add(result)
		},
		30*time.Second, time.Second,
	)

	t.Log("verifying connectivity to the HTTPRoute")

	require.Eventually(t, func() bool {
		url := fmt.Sprintf("http://%s%s", gatewayOneIPAddress, pathOne)
		bad := fmt.Sprintf("http://%s%s", gatewayOneIPAddress, pathTwo)
		req, err := http.NewRequestWithContext(GetCtx(), http.MethodGet, url, nil)
		if err != nil {
			return false
		}
		resp, err := httpc.Do(req)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		badReq, err := http.NewRequestWithContext(GetCtx(), http.MethodGet, bad, nil)
		if err != nil {
			return false
		}
		badResp, err := httpc.Do(badReq)
		if err != nil {
			return false
		}
		defer badResp.Body.Close()
		return resp.StatusCode == http.StatusOK && badResp.StatusCode == http.StatusNotFound
	}, time.Minute, time.Second)

	require.Eventually(t, func() bool {
		url := fmt.Sprintf("http://%s%s", gatewayTwoIPAddress, pathTwo)
		bad := fmt.Sprintf("http://%s%s", gatewayTwoIPAddress, pathOne)
		req, err := http.NewRequestWithContext(GetCtx(), http.MethodGet, url, nil)
		if err != nil {
			return false
		}
		resp, err := httpc.Do(req)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		badReq, err := http.NewRequestWithContext(GetCtx(), http.MethodGet, bad, nil)
		if err != nil {
			return false
		}
		badResp, err := httpc.Do(badReq)
		if err != nil {
			return false
		}
		defer badResp.Body.Close()
		return resp.StatusCode == http.StatusOK && badResp.StatusCode == http.StatusNotFound
	}, time.Minute, time.Second)

	t.Log("deleting Gateway resource")
	require.NoError(t, gatewayV1Client.Gateways(namespace.Name).Delete(GetCtx(), gatewayOne.Name, metav1.DeleteOptions{}))
	require.NoError(t, gatewayV1Client.Gateways(namespace.Name).Delete(GetCtx(), gatewayTwo.Name, metav1.DeleteOptions{}))

	t.Log("verifying that DataPlane sub-resources are deleted")
	assert.Eventually(t, func() bool {
		_, err := GetClients().OperatorClient.ApisV1beta1().DataPlanes(namespace.Name).Get(GetCtx(), dataplaneOne.Name, metav1.GetOptions{})
		return errors.IsNotFound(err)
	}, time.Minute, time.Second)
	assert.Eventually(t, func() bool {
		_, err := GetClients().OperatorClient.ApisV1beta1().DataPlanes(namespace.Name).Get(GetCtx(), dataplaneTwo.Name, metav1.GetOptions{})
		return errors.IsNotFound(err)
	}, time.Minute, time.Second)

	t.Log("verifying that ControlPlane sub-resources are deleted")
	assert.Eventually(t, func() bool {
		_, err := GetClients().OperatorClient.ApisV1beta1().ControlPlanes(namespace.Name).Get(GetCtx(), controlplaneOne.Name, metav1.GetOptions{})
		return errors.IsNotFound(err)
	}, time.Minute, time.Second)
	assert.Eventually(t, func() bool {
		_, err := GetClients().OperatorClient.ApisV1beta1().ControlPlanes(namespace.Name).Get(GetCtx(), controlplaneTwo.Name, metav1.GetOptions{})
		return errors.IsNotFound(err)
	}, time.Minute, time.Second)

	t.Log("verifying that gateways are deleted")
	require.Eventually(t, testutils.GatewayNotExist(t, GetCtx(), gatewayOneNN, clients), time.Minute, time.Second)
	require.Eventually(t, testutils.GatewayNotExist(t, GetCtx(), gatewayTwoNN, clients), time.Minute, time.Second)
}

func createHTTPRoute(parentRef metav1.Object, svc metav1.Object, path string) *gatewayv1.HTTPRoute {
	return &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: parentRef.GetNamespace(),
			Name:      uuid.NewString(),
			Annotations: map[string]string{
				"konghq.com/strip-path": "true",
			},
		},
		Spec: gatewayv1.HTTPRouteSpec{
			CommonRouteSpec: gatewayv1.CommonRouteSpec{
				ParentRefs: []gatewayv1.ParentReference{{
					Name: gatewayv1.ObjectName(parentRef.GetName()),
				}},
			},
			Rules: []gatewayv1.HTTPRouteRule{
				{
					Matches: []gatewayv1.HTTPRouteMatch{
						{
							Path: &gatewayv1.HTTPPathMatch{
								Type:  lo.ToPtr(gatewayv1.PathMatchPathPrefix),
								Value: lo.ToPtr(path),
							},
						},
					},
					BackendRefs: []gatewayv1.HTTPBackendRef{
						{
							BackendRef: gatewayv1.BackendRef{
								BackendObjectReference: gatewayv1.BackendObjectReference{
									Name: gatewayv1.ObjectName(svc.GetName()),
									Port: lo.ToPtr(gatewayv1.PortNumber(80)),
									Kind: lo.ToPtr(gatewayv1.Kind("Service")),
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestGatewayWithMultipleListeners(t *testing.T) {
	t.Parallel()
	namespace, cleaner := helpers.SetupTestEnv(t, ctx, env)

	t.Log("deploying a GatewayClass resource")
	gatewayClass := testutils.GenerateGatewayClass()
	gatewayClass, err := clients.GatewayClient.GatewayV1().GatewayClasses().Create(ctx, gatewayClass, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gatewayClass)

	t.Log("deploying Gateway resource")
	gatewayNN := types.NamespacedName{
		Name:      uuid.NewString(),
		Namespace: namespace.Name,
	}
	const port8080 = 8080
	gateway := testutils.GenerateGateway(gatewayNN, gatewayClass, func(gateway *gatewayv1.Gateway) {
		gateway.Spec.Listeners = append(gateway.Spec.Listeners,
			gatewayv1.Listener{
				Name:     "http2",
				Protocol: gatewayv1.HTTPProtocolType,
				Port:     gatewayv1.PortNumber(port8080),
			},
		)
	})
	gateway, err = clients.GatewayClient.GatewayV1().Gateways(namespace.Name).Create(ctx, gateway, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gateway)

	t.Log("verifying Gateway gets marked as Scheduled")
	require.Eventually(t, testutils.GatewayIsScheduled(t, ctx, gatewayNN, clients), testutils.GatewaySchedulingTimeLimit, time.Second)

	t.Log("verifying Gateway gets marked as Programmed")
	require.Eventually(t, testutils.GatewayIsProgrammed(t, ctx, gatewayNN, clients), testutils.GatewayReadyTimeLimit, time.Second)
	require.Eventually(t, testutils.GatewayListenersAreProgrammed(t, ctx, gatewayNN, clients), testutils.GatewayReadyTimeLimit, time.Second)

	t.Log("verifying Gateway gets the IP addresses")
	require.Eventually(t, testutils.GatewayIPAddressExist(t, ctx, gatewayNN, clients), testutils.SubresourceReadinessWait, time.Second)
	gateway = testutils.MustGetGateway(t, ctx, gatewayNN, clients)
	gatewayIPAddress := gateway.Status.Addresses[0].Value

	t.Log("verifying that the DataPlane becomes Ready")
	require.Eventually(t, testutils.GatewayDataPlaneIsReady(t, ctx, gateway, clients), testutils.SubresourceReadinessWait, time.Second)
	dataplanes := testutils.MustListDataPlanesForGateway(t, ctx, gateway, clients)
	require.Len(t, dataplanes, 1)
	dataplane := dataplanes[0]
	dataplaneNN := types.NamespacedName{Namespace: namespace.Name, Name: dataplane.Name}

	t.Log("verifying that dataplane has 1 ready replica")
	require.Eventually(t, testutils.DataPlaneHasNReadyPods(t, ctx, dataplaneNN, clients, 1), time.Minute, time.Second)

	t.Log("verifying that the ControlPlane becomes provisioned")
	require.Eventually(t, testutils.GatewayControlPlaneIsProvisioned(t, ctx, gateway, clients), testutils.SubresourceReadinessWait, time.Second)
	controlplanes := testutils.MustListControlPlanesForGateway(t, ctx, gateway, clients)
	require.Len(t, controlplanes, 1)
	controlplane := controlplanes[0]
	controlplaneNN := types.NamespacedName{Namespace: namespace.Name, Name: controlplane.Name}

	t.Log("verifying that controlplane has 1 ready replica")
	require.Eventually(t, testutils.ControlPlaneHasNReadyPods(t, ctx, controlplaneNN, clients, 1), time.Minute, time.Second)

	t.Log("verifying networkpolicies are created")
	require.Eventually(t, testutils.GatewayNetworkPoliciesExist(t, ctx, gateway, clients), testutils.SubresourceReadinessWait, time.Second)

	t.Log("verifying connectivity to the Gateway")
	require.Eventually(t, Expect404WithNoRouteFunc(t, ctx, fmt.Sprintf("http://%s:80", gatewayIPAddress)), testutils.SubresourceReadinessWait, time.Second)
	require.Eventually(t, Expect404WithNoRouteFunc(t, ctx, fmt.Sprintf("http://%s:%d", gatewayIPAddress, port8080)), testutils.SubresourceReadinessWait, time.Second)
}

func TestScalingDataPlaneThroughGatewayConfiguration(t *testing.T) {
	t.Parallel()
	namespace, cleaner := helpers.SetupTestEnv(t, GetCtx(), GetEnv())

	gatewayConfigurationName := uuid.NewString()
	t.Logf("deploying the GatewayConfiguration %s", gatewayConfigurationName)
	gatewayConfiguration := testutils.GenerateGatewayConfiguration(types.NamespacedName{Namespace: namespace.Name, Name: gatewayConfigurationName})
	gatewayConfiguration, err := GetClients().OperatorClient.ApisV1beta1().GatewayConfigurations(namespace.Name).Create(GetCtx(), gatewayConfiguration, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gatewayConfiguration)

	gatewayClass := testutils.GenerateGatewayClass()
	gatewayClass.Spec.ParametersRef = &gatewayv1.ParametersReference{
		Group:     "gateway-operator.konghq.com",
		Kind:      "GatewayConfiguration",
		Name:      gatewayConfigurationName,
		Namespace: (*gatewayv1.Namespace)(&namespace.Name),
	}
	t.Logf("deploying the GatewayClass %s", gatewayClass.Name)
	gatewayClass, err = GetClients().GatewayClient.GatewayV1().GatewayClasses().Create(GetCtx(), gatewayClass, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gatewayClass)

	t.Log("deploying Gateway resource")
	gatewayNN := types.NamespacedName{
		Name:      uuid.NewString(),
		Namespace: namespace.Name,
	}
	gateway := testutils.GenerateGateway(gatewayNN, gatewayClass)
	gateway, err = GetClients().GatewayClient.GatewayV1().Gateways(namespace.Name).Create(GetCtx(), gateway, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gateway)

	t.Log("verifying Gateway gets marked as Scheduled")
	require.Eventually(t, testutils.GatewayIsScheduled(t, GetCtx(), gatewayNN, clients), testutils.GatewaySchedulingTimeLimit, time.Second)

	t.Log("verifying Gateway gets marked as Programmed")
	require.Eventually(t, testutils.GatewayIsProgrammed(t, GetCtx(), gatewayNN, clients), testutils.GatewayReadyTimeLimit, time.Second)
	require.Eventually(t, testutils.GatewayListenersAreProgrammed(t, GetCtx(), gatewayNN, clients), testutils.GatewayReadyTimeLimit, time.Second)

	t.Log("verifying that the ControlPlane becomes provisioned")
	require.Eventually(t, testutils.GatewayControlPlaneIsProvisioned(t, GetCtx(), gateway, clients), testutils.SubresourceReadinessWait, time.Second)

	t.Log("verifying that the DataPlane becomes ready")
	require.Eventually(t, testutils.GatewayDataPlaneIsReady(t, GetCtx(), gateway, clients), testutils.SubresourceReadinessWait, time.Second)

	testCases := []struct {
		name                       string
		dataplaneDeploymentOptions operatorv1beta1.DeploymentOptions
		expectedReplicasCount      int
	}{
		{
			name: "replicas=3",
			dataplaneDeploymentOptions: operatorv1beta1.DeploymentOptions{
				Replicas: lo.ToPtr[int32](3),
			},
			expectedReplicasCount: 3,
		},
		{
			name: "replicas=0",
			dataplaneDeploymentOptions: operatorv1beta1.DeploymentOptions{
				Replicas: lo.ToPtr[int32](0),
			},
			expectedReplicasCount: 0,
		},
		{
			name: "replicas=5",
			dataplaneDeploymentOptions: operatorv1beta1.DeploymentOptions{
				Replicas: lo.ToPtr[int32](5),
			},
			expectedReplicasCount: 5,
		},
		{
			name: "replicas=1",
			dataplaneDeploymentOptions: operatorv1beta1.DeploymentOptions{
				Replicas: lo.ToPtr[int32](1),
			},
			expectedReplicasCount: 1,
		},
		{
			name: "horizontal scaling with minReplicas=3",
			dataplaneDeploymentOptions: operatorv1beta1.DeploymentOptions{
				Scaling: &operatorv1beta1.Scaling{
					HorizontalScaling: &operatorv1beta1.HorizontalScaling{
						MinReplicas: lo.ToPtr[int32](3),
						MaxReplicas: 5,
					},
				},
			},
			expectedReplicasCount: 3,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deploymentOptions := tc.dataplaneDeploymentOptions
			gatewayConfiguration, err := GetClients().OperatorClient.ApisV1beta1().GatewayConfigurations(namespace.Name).Get(GetCtx(), gatewayConfigurationName, metav1.GetOptions{})
			require.NoError(t, err)
			gatewayConfiguration.Spec.DataPlaneOptions.Deployment.DeploymentOptions = deploymentOptions
			t.Logf("changing the GatewayConfiguration to change dataplane deploymentOptions to %v", deploymentOptions)
			_, err = GetClients().OperatorClient.ApisV1beta1().GatewayConfigurations(namespace.Name).Update(GetCtx(), gatewayConfiguration, metav1.UpdateOptions{})
			require.NoError(t, err)

			t.Log("verifying the deployment managed by the controlplane is ready")
			controlplanes := testutils.MustListControlPlanesForGateway(t, GetCtx(), gateway, clients)
			require.Len(t, controlplanes, 1)
			controlplaneNN := client.ObjectKeyFromObject(&controlplanes[0])
			require.Eventually(t, testutils.ControlPlaneHasActiveDeployment(t,
				GetCtx(),
				controlplaneNN,
				clients), testutils.ControlPlaneCondDeadline, testutils.ControlPlaneCondTick)

			t.Logf("verifying the deployment managed by the dataplane is ready and has %d available dataplane replicas", tc.expectedReplicasCount)
			dataplanes := testutils.MustListDataPlanesForGateway(t, GetCtx(), gateway, clients)
			require.Len(t, dataplanes, 1)
			dataplane := dataplanes[0]
			dataplaneNN := client.ObjectKeyFromObject(&dataplane)
			require.Eventually(t, testutils.DataPlaneHasActiveDeployment(t,
				GetCtx(),
				dataplaneNN,
				&appsv1.Deployment{},
				client.MatchingLabels{
					consts.GatewayOperatorManagedByLabel: consts.DataPlaneManagedLabelValue,
				},
				clients), testutils.DataPlaneCondDeadline, testutils.DataPlaneCondTick)
			require.Eventually(t, testutils.DataPlaneHasNReadyPods(t, GetCtx(), dataplaneNN, clients, tc.expectedReplicasCount), time.Minute, time.Second)
		})
	}
}

func TestGatewayDataPlaneNetworkPolicy(t *testing.T) {
	t.Parallel()
	namespace, cleaner := helpers.SetupTestEnv(t, GetCtx(), GetEnv())

	var err error
	gatewayConfigurationName := uuid.NewString()
	t.Log("deploying a GatewayConfiguration resource")
	gatewayConfiguration := testutils.GenerateGatewayConfiguration(types.NamespacedName{Namespace: namespace.Name, Name: gatewayConfigurationName})
	gatewayConfiguration, err = GetClients().OperatorClient.ApisV1beta1().GatewayConfigurations(namespace.Name).Create(GetCtx(), gatewayConfiguration, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gatewayConfiguration)

	t.Log("deploying a GatewayClass resource")
	gatewayClass := testutils.GenerateGatewayClass()
	gatewayClass.Spec.ParametersRef = &gatewayv1.ParametersReference{
		Group:     "gateway-operator.konghq.com",
		Kind:      "GatewayConfiguration",
		Name:      gatewayConfigurationName,
		Namespace: (*gatewayv1.Namespace)(&namespace.Name),
	}
	gatewayClass, err = GetClients().GatewayClient.GatewayV1().GatewayClasses().Create(GetCtx(), gatewayClass, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gatewayClass)

	t.Log("deploying Gateway resource")
	gatewayNN := types.NamespacedName{
		Name:      uuid.NewString(),
		Namespace: namespace.Name,
	}
	gateway := testutils.GenerateGateway(gatewayNN, gatewayClass)
	gateway, err = GetClients().GatewayClient.GatewayV1().Gateways(namespace.Name).Create(GetCtx(), gateway, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gateway)

	t.Log("verifying Gateway gets marked as Scheduled")
	require.Eventually(t, testutils.GatewayIsScheduled(t, GetCtx(), gatewayNN, clients), testutils.GatewaySchedulingTimeLimit, time.Second)

	t.Log("verifying Gateway gets marked as Programmed")
	require.Eventually(t, testutils.GatewayIsProgrammed(t, GetCtx(), gatewayNN, clients), testutils.GatewayReadyTimeLimit, time.Second)
	require.Eventually(t, testutils.GatewayListenersAreProgrammed(t, GetCtx(), gatewayNN, clients), testutils.GatewayReadyTimeLimit, time.Second)

	t.Log("verifying that the DataPlane becomes provisioned")
	require.Eventually(t, testutils.GatewayDataPlaneIsReady(t, GetCtx(), gateway, clients), testutils.SubresourceReadinessWait, time.Second)
	dataplanes := testutils.MustListDataPlanesForGateway(t, GetCtx(), gateway, clients)
	require.Len(t, dataplanes, 1)
	dataplane := dataplanes[0]

	t.Log("verifying that the ControlPlane becomes provisioned")
	require.Eventually(t, testutils.GatewayControlPlaneIsProvisioned(t, GetCtx(), gateway, clients), testutils.SubresourceReadinessWait, time.Second)
	controlplanes := testutils.MustListControlPlanesForGateway(t, GetCtx(), gateway, clients)
	require.Len(t, controlplanes, 1)
	controlplane := controlplanes[0]

	t.Log("verifying DataPlane's NetworkPolicies is created")
	require.Eventually(t, testutils.GatewayNetworkPoliciesExist(t, GetCtx(), gateway, clients), testutils.SubresourceReadinessWait, time.Second)
	networkpolicies := testutils.MustListNetworkPoliciesForGateway(t, GetCtx(), gateway, clients)
	require.Len(t, networkpolicies, 1)
	networkPolicy := networkpolicies[0]
	require.Equal(t, map[string]string{"app": dataplane.Name}, networkPolicy.Spec.PodSelector.MatchLabels)

	t.Log("verifying that the DataPlane's Pod Admin API is network restricted to ControlPlane Pods")
	var expectLimitedAdminAPI networkPolicyIngressRuleDecorator
	expectLimitedAdminAPI.withProtocolPort(corev1.ProtocolTCP, consts.DataPlaneAdminAPIPort)
	expectLimitedAdminAPI.withPeerMatchLabels(
		map[string]string{"app": controlplane.Name},
		map[string]string{"kubernetes.io/metadata.name": dataplane.Namespace},
	)

	t.Log("verifying that the DataPlane's proxy ingress traffic is allowed")
	var expectAllowProxyIngress networkPolicyIngressRuleDecorator
	expectAllowProxyIngress.withProtocolPort(corev1.ProtocolTCP, consts.DataPlaneProxyPort)
	expectAllowProxyIngress.withProtocolPort(corev1.ProtocolTCP, consts.DataPlaneProxySSLPort)

	t.Log("verifying that the DataPlane's metrics ingress traffic is allowed")
	var expectAllowMetricsIngress networkPolicyIngressRuleDecorator
	expectAllowMetricsIngress.withProtocolPort(corev1.ProtocolTCP, consts.DataPlaneMetricsPort)

	t.Log("verifying DataPlane's NetworkPolicies ingress rules correctness")
	require.Contains(t, networkPolicy.Spec.Ingress, expectLimitedAdminAPI.Rule)
	require.Contains(t, networkPolicy.Spec.Ingress, expectAllowProxyIngress.Rule)
	require.Contains(t, networkPolicy.Spec.Ingress, expectAllowMetricsIngress.Rule)

	t.Log("deleting DataPlane's NetworkPolicies")
	require.NoError(t,
		GetClients().K8sClient.NetworkingV1().
			NetworkPolicies(networkPolicy.Namespace).
			Delete(GetCtx(), networkPolicy.Name, metav1.DeleteOptions{}),
	)

	t.Log("verifying NetworkPolicies are recreated")
	require.Eventually(t, testutils.GatewayNetworkPoliciesExist(t, GetCtx(), gateway, clients), testutils.SubresourceReadinessWait, time.Second)
	networkpolicies = testutils.MustListNetworkPoliciesForGateway(t, GetCtx(), gateway, clients)
	require.Len(t, networkpolicies, 1)
	networkPolicy = networkpolicies[0]
	t.Logf("NetworkPolicy generation %d", networkPolicy.Generation)

	t.Log("verifying DataPlane's NetworkPolicies ingress rules correctness")
	require.Contains(t, networkPolicy.Spec.Ingress, expectLimitedAdminAPI.Rule)
	require.Contains(t, networkPolicy.Spec.Ingress, expectAllowProxyIngress.Rule)
	require.Contains(t, networkPolicy.Spec.Ingress, expectAllowMetricsIngress.Rule)

	t.Run("verifying DataPlane's NetworkPolicies get updated after customizing kong proxy listen port through GatewayConfiguration", func(t *testing.T) {
		// TODO: https://github.com/Kong/gateway-operator/issues/695
		t.Skip("re-enable once https://github.com/Kong/gateway-operator/issues/695 is fixed")

		gwcClient := GetClients().OperatorClient.ApisV1beta1().GatewayConfigurations(namespace.Name)

		setGatewayConfigurationEnvProxyPort(t, gatewayConfiguration, 8005, 8999)
		gatewayConfiguration, err = gwcClient.Update(GetCtx(), gatewayConfiguration, metav1.UpdateOptions{})
		require.NoError(t, err)

		t.Log("ingress rules get updated with configured proxy listen port")
		var expectedUpdatedProxyListenPort networkPolicyIngressRuleDecorator
		expectedUpdatedProxyListenPort.withProtocolPort(corev1.ProtocolTCP, 8005)
		expectedUpdatedProxyListenPort.withProtocolPort(corev1.ProtocolTCP, 8999)
		require.Eventually(t,
			testutils.GatewayNetworkPolicyForGatewayContainsRules(t, GetCtx(), gateway, clients, expectedUpdatedProxyListenPort.Rule),
			testutils.SubresourceReadinessWait, time.Second)
		var notExpectedUpdatedProxyListenPort networkPolicyIngressRuleDecorator
		notExpectedUpdatedProxyListenPort.withProtocolPort(corev1.ProtocolTCP, consts.DataPlaneProxyPort)
		require.Eventually(t,
			testutils.Not(
				testutils.GatewayNetworkPolicyForGatewayContainsRules(t, GetCtx(), gateway, clients, notExpectedUpdatedProxyListenPort.Rule),
			),
			testutils.SubresourceReadinessWait, time.Second)

		t.Log("ingress rules get updated with configured admin listen port")
		setGatewayConfigurationEnvAdminAPIPort(t, gatewayConfiguration, 8555)
		_, err = gwcClient.Update(GetCtx(), gatewayConfiguration, metav1.UpdateOptions{})
		require.NoError(t, err)

		var expectedUpdatedLimitedAdminAPI networkPolicyIngressRuleDecorator
		expectedUpdatedLimitedAdminAPI.withProtocolPort(corev1.ProtocolTCP, 8555)
		expectedUpdatedLimitedAdminAPI.withPeerMatchLabels(
			map[string]string{"app": controlplane.Name},
			map[string]string{"kubernetes.io/metadata.name": controlplane.Namespace},
		)
		if !assert.Eventually(t,
			testutils.GatewayNetworkPolicyForGatewayContainsRules(t, GetCtx(), gateway, clients, expectedUpdatedLimitedAdminAPI.Rule),
			2*testutils.SubresourceReadinessWait, time.Second,
			"NetworkPolicy didn't get updated with port 8555 after a corresponding change to GatewayConfiguration") {
			networkpolicies, err := gatewayutils.ListNetworkPoliciesForGateway(GetCtx(), GetClients().MgrClient, gateway)
			require.NoError(t, err)
			t.Log("DataPlane's NetworkPolicies")
			for _, np := range networkpolicies {
				t.Logf("%# v\n", pretty.Formatter(np))
			}
		}

		var notExpectedUpdatedLimitedAdminAPI networkPolicyIngressRuleDecorator
		notExpectedUpdatedLimitedAdminAPI.withProtocolPort(corev1.ProtocolTCP, consts.DataPlaneAdminAPIPort)
		notExpectedUpdatedLimitedAdminAPI.withPeerMatchLabels(
			map[string]string{"app": controlplane.Name},
			map[string]string{"kubernetes.io/metadata.name": controlplane.Namespace},
		)
		require.Eventually(t,
			testutils.Not(testutils.GatewayNetworkPolicyForGatewayContainsRules(t, GetCtx(), gateway, clients, notExpectedUpdatedLimitedAdminAPI.Rule)),
			testutils.SubresourceReadinessWait, time.Second)
	})

	t.Run("verifying DataPlane's NetworkPolicies get deleted after Gateway is deleted", func(t *testing.T) {
		t.Log("deleting Gateway resource")
		require.NoError(t, GetClients().GatewayClient.GatewayV1().Gateways(namespace.Name).Delete(GetCtx(), gateway.Name, metav1.DeleteOptions{}))

		t.Log("verifying networkpolicies are deleted")
		require.Eventually(t, testutils.Not(testutils.GatewayNetworkPoliciesExist(t, GetCtx(), gateway, clients)), time.Minute, time.Second)
	})
}

func TestGatewayProvisionDataPlaneFail(t *testing.T) {
	t.Parallel()
	if !webhookEnabled {
		t.Skip("Skipping because webhook is not enabled")
	}

	namespace, cleaner := helpers.SetupTestEnv(t, GetCtx(), GetEnv())

	t.Logf("creating another service with no endpoints for wrong webhooks")
	serviceClient := env.Cluster().Client().CoreV1().Services(namespace.Name)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace.Name,
			Name:      "fake-webhook",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "fake",
			},
			Ports: []corev1.ServicePort{
				{Port: int32(443), TargetPort: intstr.FromInt(8443)},
			},
		},
	}
	_, err := serviceClient.Create(ctx, service, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(service)

	t.Log("creating a wrong webhook for DataPlane with a Service having no endpoints")
	webhookClient := env.Cluster().Client().AdmissionregistrationV1().ValidatingWebhookConfigurations()
	wrongWebhook := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gateway-operator-wrong-validators.test.konghq.com",
		},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{
			{
				Name: "wrong-dataplane-webhook.test.konghq.com",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Namespace: namespace.Name,
						Name:      "fake-webhook",
					},
				},
				Rules: []admissionregistrationv1.RuleWithOperations{
					{
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"gateway-operator.konghq.com"},
							APIVersions: []string{"v1beta1"},
							Resources:   []string{"dataplanes"},
							Scope:       lo.ToPtr(admissionregistrationv1.NamespacedScope),
						},
						Operations: []admissionregistrationv1.OperationType{
							admissionregistrationv1.Create,
							admissionregistrationv1.Update,
						},
					},
				},
				AdmissionReviewVersions: []string{"v1", "v1beta1"},
				SideEffects:             lo.ToPtr(admissionregistrationv1.SideEffectClassNone),
				TimeoutSeconds:          lo.ToPtr(int32(1)),
			},
		},
	}
	_, err = webhookClient.Create(ctx, wrongWebhook, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(wrongWebhook)

	t.Log("creating a Gateway and verify that it does not get Programmed")
	t.Log("deploying a GatewayClass resource")
	gatewayClass := testutils.GenerateGatewayClass()
	gatewayClass, err = GetClients().GatewayClient.GatewayV1().GatewayClasses().Create(GetCtx(), gatewayClass, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gatewayClass)

	t.Log("deploying Gateway resource")
	gatewayNN := types.NamespacedName{
		Name:      uuid.NewString(),
		Namespace: namespace.Name,
	}
	gateway := testutils.GenerateGateway(gatewayNN, gatewayClass)
	gateway, err = GetClients().GatewayClient.GatewayV1().Gateways(namespace.Name).Create(GetCtx(), gateway, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gateway)

	t.Log("verifying Gateway gets marked as Scheduled")
	require.Eventually(t, testutils.GatewayIsScheduled(t, GetCtx(), gatewayNN, clients), testutils.GatewaySchedulingTimeLimit, time.Second)

	t.Log("verifying Gateway does not get marked as Programmed")
	require.Never(t, testutils.GatewayIsProgrammed(t, GetCtx(), gatewayNN, clients), testutils.GatewayReadyTimeLimit, time.Second)

	t.Log("deleting the wrong webhook for dataplane")
	err = webhookClient.Delete(ctx, wrongWebhook.Name, metav1.DeleteOptions{})
	require.NoError(t, err)

	t.Log("verifying Gateway gets marked as Programmed")
	require.Eventually(t, testutils.GatewayIsProgrammed(t, GetCtx(), gatewayNN, clients), testutils.GatewayReadyTimeLimit, time.Second)
}

func setGatewayConfigurationEnvProxyPort(t *testing.T, gatewayConfiguration *operatorv1beta1.GatewayConfiguration, proxyPort int, proxySSLPort int) {
	t.Helper()

	dpOptions := gatewayConfiguration.Spec.DataPlaneOptions
	if dpOptions == nil {
		dpOptions = &operatorv1beta1.GatewayConfigDataPlaneOptions{}
	}
	if dpOptions.Deployment.PodTemplateSpec == nil {
		dpOptions.Deployment.PodTemplateSpec = &corev1.PodTemplateSpec{}
	}

	container := k8sutils.GetPodContainerByName(&dpOptions.Deployment.PodTemplateSpec.Spec, consts.DataPlaneProxyContainerName)
	require.NotNil(t, container)

	container.Env = SetEnvValueByName(container.Env,
		"KONG_PROXY_LISTEN",
		fmt.Sprintf("0.0.0.0:%d reuseport backlog=16384, 0.0.0.0:%d http2 ssl reuseport backlog=16384", proxyPort, proxySSLPort),
	)
	container.Env = SetEnvValueByName(container.Env,
		"KONG_PORT_MAPS",
		fmt.Sprintf("80:%d, 443:%d", proxyPort, proxySSLPort),
	)

	gatewayConfiguration.Spec.DataPlaneOptions = dpOptions
}

func setGatewayConfigurationEnvAdminAPIPort(t *testing.T, gatewayConfiguration *operatorv1beta1.GatewayConfiguration, adminAPIPort int) {
	t.Helper()

	dpOptions := gatewayConfiguration.Spec.DataPlaneOptions
	if dpOptions == nil {
		dpOptions = &operatorv1beta1.GatewayConfigDataPlaneOptions{}
	}

	container := k8sutils.GetPodContainerByName(&dpOptions.Deployment.PodTemplateSpec.Spec, consts.DataPlaneProxyContainerName)
	require.NotNil(t, container)

	container.Env = SetEnvValueByName(container.Env,
		"KONG_ADMIN_LISTEN",
		fmt.Sprintf("0.0.0.0:%d ssl reuseport backlog=16384", adminAPIPort),
	)

	gatewayConfiguration.Spec.DataPlaneOptions = dpOptions
}

type networkPolicyIngressRuleDecorator struct {
	Rule networkingv1.NetworkPolicyIngressRule
}

func (d *networkPolicyIngressRuleDecorator) withProtocolPort(protocol corev1.Protocol, port int) { //nolint:unparam
	portIntStr := intstr.FromInt(port)
	d.Rule.Ports = append(d.Rule.Ports, networkingv1.NetworkPolicyPort{
		Protocol: &protocol,
		Port:     &portIntStr,
	})
}

func (d *networkPolicyIngressRuleDecorator) withPeerMatchLabels(
	podSelector map[string]string,
	namespaceSelector map[string]string,
) {
	d.Rule.From = append(d.Rule.From, networkingv1.NetworkPolicyPeer{
		PodSelector: &metav1.LabelSelector{
			MatchLabels: podSelector,
		},
		NamespaceSelector: &metav1.LabelSelector{
			MatchLabels: namespaceSelector,
		},
	})
}
