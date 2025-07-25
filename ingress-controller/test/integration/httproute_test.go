//go:build integration_tests

package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kong/go-kong/kong"
	ktfkong "github.com/kong/kubernetes-testing-framework/pkg/clusters/addons/kong"
	"github.com/kong/kubernetes-testing-framework/pkg/utils/kubernetes/generators"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayclient "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	"github.com/kong/kubernetes-configuration/pkg/clientset"

	"github.com/kong/kong-operator/ingress-controller/internal/annotations"
	"github.com/kong/kong-operator/ingress-controller/internal/gatewayapi"
	"github.com/kong/kong-operator/ingress-controller/internal/util"
	"github.com/kong/kong-operator/ingress-controller/internal/util/builder"
	"github.com/kong/kong-operator/ingress-controller/test"
	"github.com/kong/kong-operator/ingress-controller/test/internal/helpers"
)

var emptyHeaderSet = make(map[string]string)

func TestHTTPRouteEssentials(t *testing.T) {
	ctx := t.Context()

	ns, cleaner := helpers.Setup(ctx, t, env)

	t.Log("getting a gateway client")
	gatewayClient, err := gatewayclient.NewForConfig(env.Cluster().Config())
	require.NoError(t, err)

	t.Log("deploying a new gatewayClass")
	gatewayClassName := uuid.NewString()
	gwc, err := helpers.DeployGatewayClass(ctx, gatewayClient, gatewayClassName)
	require.NoError(t, err)
	cleaner.Add(gwc)

	t.Log("deploying a new gateway")
	gatewayName := uuid.NewString()
	gateway, err := helpers.DeployGateway(ctx, gatewayClient, ns.Name, gatewayClassName, func(gw *gatewayapi.Gateway) {
		gw.Name = gatewayName
		// add a UDP listener to check the HTTPRoute does not get attached to it.
		gw.Spec.Listeners = append(gw.Spec.Listeners, gatewayapi.Listener{
			Name:     "udp",
			Protocol: gatewayapi.UDPProtocolType,
			Port:     gatewayapi.PortNumber(ktfkong.DefaultUDPServicePort),
		})
	})
	require.NoError(t, err)
	cleaner.Add(gateway)

	t.Log("deploying a minimal HTTP container deployment to test Ingress routes")
	container := generators.NewContainer("httpbin", test.HTTPBinImage, test.HTTPBinPort)
	deployment := generators.NewDeploymentForContainer(container)
	deployment, err = env.Cluster().Client().AppsV1().Deployments(ns.Name).Create(ctx, deployment, metav1.CreateOptions{})
	require.NoError(t, err)

	t.Logf("exposing deployment %s via service", deployment.Name)
	service := generators.NewServiceForDeployment(deployment, corev1.ServiceTypeLoadBalancer)
	_, err = env.Cluster().Client().CoreV1().Services(ns.Name).Create(ctx, service, metav1.CreateOptions{})
	require.NoError(t, err)

	kongplugin := &configurationv1.KongPlugin{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns.Name,
			Name:      "correlation",
		},
		PluginName: "correlation-id",
		Config: apiextensionsv1.JSON{
			Raw: []byte(`{"header_name":"reqid", "echo_downstream": true}`),
		},
	}
	pluginClient, err := clientset.NewForConfig(env.Cluster().Config())
	require.NoError(t, err)
	kongplugin, err = pluginClient.ConfigurationV1().KongPlugins(ns.Name).Create(ctx, kongplugin, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(kongplugin)

	t.Logf("creating an httproute to access deployment %s via kong", deployment.Name)
	httpPort := gatewayapi.PortNumber(80)
	pathMatchPrefix := gatewayapi.PathMatchPathPrefix
	pathMatchRegularExpression := gatewayapi.PathMatchRegularExpression
	pathMatchExact := gatewayapi.PathMatchExact
	headerMatchRegex := gatewayapi.HeaderMatchRegularExpression
	httpRoute := &gatewayapi.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name: uuid.NewString(),
			Annotations: map[string]string{
				annotations.AnnotationPrefix + annotations.StripPathKey: "true",
				annotations.AnnotationPrefix + annotations.PluginsKey:   "correlation",
			},
		},
		Spec: gatewayapi.HTTPRouteSpec{
			CommonRouteSpec: gatewayapi.CommonRouteSpec{
				ParentRefs: []gatewayapi.ParentReference{{
					Name: gatewayapi.ObjectName(gateway.Name),
				}},
			},
			Rules: []gatewayapi.HTTPRouteRule{{
				Matches: []gatewayapi.HTTPRouteMatch{
					{
						Path: &gatewayapi.HTTPPathMatch{
							Type:  &pathMatchPrefix,
							Value: kong.String("/test-http-route-essentials"),
						},
					},
					{
						Path: &gatewayapi.HTTPPathMatch{
							Type:  &pathMatchRegularExpression,
							Value: kong.String(`/2/test-http-route-essentials/regex/\d{3}`),
						},
					},
					{
						Path: &gatewayapi.HTTPPathMatch{
							Type:  &pathMatchExact,
							Value: kong.String(`/3/exact-test-http-route-essentials`),
						},
					},
				},
				BackendRefs: []gatewayapi.HTTPBackendRef{{
					BackendRef: gatewayapi.BackendRef{
						BackendObjectReference: gatewayapi.BackendObjectReference{
							Name: gatewayapi.ObjectName(service.Name),
							Port: &httpPort,
							Kind: util.StringToGatewayAPIKindPtr("Service"),
						},
					},
				}},
			}},
		},
	}
	httpRoute, err = gatewayClient.GatewayV1().HTTPRoutes(ns.Name).Create(ctx, httpRoute, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(httpRoute)

	t.Log("verifying that the Gateway gets linked to the route via status")
	callback := helpers.GetGatewayIsLinkedCallback(ctx, t, gatewayClient, gatewayapi.HTTPProtocolType, ns.Name, httpRoute.Name)
	require.Eventually(t, callback, ingressWait, waitTick)
	t.Log("verifying that the httproute contains 'Programmed' condition")
	require.Eventually(t,
		helpers.GetVerifyProgrammedConditionCallback(t, gatewayClient, gatewayapi.HTTPProtocolType, ns.Name, httpRoute.Name, metav1.ConditionTrue),
		ingressWait, waitTick,
	)

	t.Log("waiting for routes from HTTPRoute to become operational")
	helpers.EventuallyGETPath(t, proxyHTTPURL, proxyHTTPURL.Host, "/test-http-route-essentials", nil, http.StatusOK, "<title>httpbin.org</title>", emptyHeaderSet, ingressWait, waitTick)
	helpers.EventuallyGETPath(t, proxyHTTPURL, proxyHTTPURL.Host, "/test-http-route-essentials/base64/wqt5b8q7ccK7IGRhbiBib3NocWEgYmlyIGphdm9iaW1peiB5b8q7cWRpci4K",
		nil, http.StatusOK, "«yoʻq» dan boshqa bir javobimiz yoʻqdir.", emptyHeaderSet, ingressWait, waitTick)
	helpers.EventuallyGETPath(t, proxyHTTPURL, proxyHTTPURL.Host, "/2/test-http-route-essentials/regex/999", nil, http.StatusOK, "<title>httpbin.org</title>", emptyHeaderSet, ingressWait, waitTick)
	helpers.EventuallyGETPath(t, proxyHTTPURL, proxyHTTPURL.Host, "/3/exact-test-http-route-essentials", nil, http.StatusOK, "<title>httpbin.org</title>", emptyHeaderSet, ingressWait, waitTick)
	helpers.EventuallyGETPath(t, proxyHTTPURL, proxyHTTPURL.Host, "/3/exact-test-http-route-essentialsNO", nil, http.StatusNotFound, "no Route matched", emptyHeaderSet, ingressWait, waitTick)

	require.EventuallyWithT(t, func(c *assert.CollectT) {
		req := helpers.MustHTTPRequest(t, http.MethodGet, proxyHTTPURL.Host, "/test-http-route-essentials", nil)
		resp, err := helpers.DefaultHTTPClient(helpers.WithResolveHostTo(proxyHTTPURL.Host)).Do(req)
		if !assert.NoError(c, err) {
			return
		}
		defer resp.Body.Close()
		_, ok := resp.Header["Reqid"]
		assert.True(c, ok, "expected header Reqid to be present in response")
	}, ingressWait, waitTick)

	t.Run("header regex match", func(t *testing.T) {
		httpRoute, err = gatewayClient.GatewayV1().HTTPRoutes(ns.Name).Get(ctx, httpRoute.Name, metav1.GetOptions{})
		require.NoError(t, err)

		httpRoute.Spec.Rules[0].Matches = append(httpRoute.Spec.Rules[0].Matches, gatewayapi.HTTPRouteMatch{
			Headers: []gatewayapi.HTTPHeaderMatch{
				{
					Type:  &headerMatchRegex,
					Value: `^audio/.*`,
					Name:  "Content-Type",
				},
			},
		})
		httpRoute, err = gatewayClient.GatewayV1().HTTPRoutes(ns.Name).Update(ctx, httpRoute, metav1.UpdateOptions{})
		require.NoError(t, err)

		t.Log("verifying HTTPRoute header match")
		helpers.EventuallyGETPath(t, proxyHTTPURL, proxyHTTPURL.Host, "/", nil, http.StatusOK, "<title>httpbin.org</title>", map[string]string{"Content-Type": "audio/mp3"}, ingressWait, waitTick)
	})

	t.Run("HTTPRoute query param match", func(t *testing.T) {
		RunWhenKongExpressionRouter(ctx, t)

		httpRoute, err = gatewayClient.GatewayV1().HTTPRoutes(ns.Name).Get(ctx, httpRoute.Name, metav1.GetOptions{})
		require.NoError(t, err)

		httpRoute.Spec.Rules[0].Matches = append(httpRoute.Spec.Rules[0].Matches, gatewayapi.HTTPRouteMatch{
			QueryParams: []gatewayapi.HTTPQueryParamMatch{
				{
					Type:  lo.ToPtr(gatewayapi.QueryParamMatchExact),
					Name:  "foo",
					Value: "bar",
				},
			},
		})
		httpRoute, err = gatewayClient.GatewayV1().HTTPRoutes(ns.Name).Update(ctx, httpRoute, metav1.UpdateOptions{})
		require.NoError(t, err)

		t.Log("verifying HTTPRoute query param match")
		helpers.EventuallyGETPath(t, proxyHTTPURL, proxyHTTPURL.Host, "/?foo=bar", nil, http.StatusOK, "<title>httpbin.org</title>", nil, ingressWait, waitTick)
	})

	t.Log("verifying that the HTTPRoute has the Condition 'Accepted' set to 'True'")
	require.Eventually(t, HTTPRouteMatchesAcceptedCallback(t, gatewayClient, httpRoute, true, gatewayapi.RouteReasonAccepted), statusWait, waitTick)

	t.Log("verifying that the Gateway listener have the proper attachedRoutes")
	require.Eventually(t, ListenersHaveNAttachedRoutesCallback(t, gatewayClient, ns.Name, gatewayName, map[string]int32{
		"http":  1,
		"https": 1,
		"udp":   0,
	}), statusWait, waitTick)

	t.Log("removing the parentrefs from the HTTPRoute")
	oldParentRefs := httpRoute.Spec.ParentRefs
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		httpRoute, err = gatewayClient.GatewayV1().HTTPRoutes(ns.Name).Get(ctx, httpRoute.Name, metav1.GetOptions{})
		if !assert.NoError(c, err) {
			return
		}
		httpRoute.Spec.ParentRefs = nil
		httpRoute, err = gatewayClient.GatewayV1().HTTPRoutes(ns.Name).Update(ctx, httpRoute, metav1.UpdateOptions{})
		assert.NoError(c, err)
	}, time.Minute, time.Second)

	t.Log("verifying that the Gateway gets unlinked from the route via status")
	callback = helpers.GetGatewayIsUnlinkedCallback(ctx, t, gatewayClient, gatewayapi.HTTPProtocolType, ns.Name, httpRoute.Name)
	require.Eventually(t, callback, ingressWait, waitTick)

	t.Log("verifying that the data-plane configuration from the HTTPRoute gets dropped with the parentRefs now removed")
	helpers.EventuallyGETPath(t, proxyHTTPURL, proxyHTTPURL.Host, "/test-http-route-essentials", nil, http.StatusNotFound, "", emptyHeaderSet, ingressWait, waitTick)

	t.Log("putting the parentRefs back")
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		httpRoute, err = gatewayClient.GatewayV1().HTTPRoutes(ns.Name).Get(ctx, httpRoute.Name, metav1.GetOptions{})
		if !assert.NoError(c, err) {
			return
		}
		httpRoute.Spec.ParentRefs = oldParentRefs
		httpRoute, err = gatewayClient.GatewayV1().HTTPRoutes(ns.Name).Update(ctx, httpRoute, metav1.UpdateOptions{})
		assert.NoError(c, err)
	}, time.Minute, time.Second)

	t.Log("verifying that the Gateway gets linked to the route via status")
	callback = helpers.GetGatewayIsLinkedCallback(ctx, t, gatewayClient, gatewayapi.HTTPProtocolType, ns.Name, httpRoute.Name)
	require.Eventually(t, callback, ingressWait, waitTick)

	t.Log("verifying that putting the parentRefs back results in the routes becoming available again")
	helpers.EventuallyGETPath(t, proxyHTTPURL, proxyHTTPURL.Host, "/test-http-route-essentials", nil, http.StatusOK, "<title>httpbin.org</title>", emptyHeaderSet, ingressWait, waitTick)

	t.Log("deleting the GatewayClass")
	require.NoError(t, gatewayClient.GatewayV1().GatewayClasses().Delete(ctx, gatewayClassName, metav1.DeleteOptions{}))

	t.Log("verifying that the Gateway gets unlinked from the route via status")
	callback = helpers.GetGatewayIsUnlinkedCallback(ctx, t, gatewayClient, gatewayapi.HTTPProtocolType, ns.Name, httpRoute.Name)
	require.Eventually(t, callback, ingressWait, waitTick)
	t.Log("verifying that the data-plane configuration from the HTTPRoute gets dropped with the GatewayClass now removed")
	helpers.EventuallyGETPath(t, proxyHTTPURL, proxyHTTPURL.Host, "/test-http-route-essentials", nil, http.StatusNotFound, "", emptyHeaderSet, ingressWait, waitTick)

	t.Log("putting the GatewayClass back")
	gwc, err = helpers.DeployGatewayClass(ctx, gatewayClient, gatewayClassName)
	require.NoError(t, err)
	cleaner.Add(gwc)

	t.Log("verifying that the Gateway gets linked to the route via status")
	callback = helpers.GetGatewayIsLinkedCallback(ctx, t, gatewayClient, gatewayapi.HTTPProtocolType, ns.Name, httpRoute.Name)
	require.Eventually(t, callback, ingressWait, waitTick)

	t.Log("verifying that creating the GatewayClass again triggers reconciliation of HTTPRoutes and the route becomes available again")
	helpers.EventuallyGETPath(t, proxyHTTPURL, proxyHTTPURL.Host, "/test-http-route-essentials", nil, http.StatusOK, "<title>httpbin.org</title>", emptyHeaderSet, ingressWait, waitTick)

	t.Log("deleting the Gateway")
	require.NoError(t, gatewayClient.GatewayV1().Gateways(ns.Name).Delete(ctx, gatewayName, metav1.DeleteOptions{}))

	t.Log("verifying that the Gateway gets unlinked from the route via status")
	callback = helpers.GetGatewayIsUnlinkedCallback(ctx, t, gatewayClient, gatewayapi.HTTPProtocolType, ns.Name, httpRoute.Name)
	require.Eventually(t, callback, ingressWait, waitTick)

	t.Log("verifying that the data-plane configuration from the HTTPRoute gets dropped with the Gateway now removed")
	helpers.EventuallyGETPath(t, proxyHTTPURL, proxyHTTPURL.Host, "/test-http-route-essentials", nil, http.StatusNotFound, "", emptyHeaderSet, ingressWait, waitTick)

	t.Log("putting the Gateway back")
	gateway, err = helpers.DeployGateway(ctx, gatewayClient, ns.Name, gatewayClassName, func(gw *gatewayapi.Gateway) {
		gw.Name = gatewayName
	})
	require.NoError(t, err)

	t.Log("verifying that the Gateway gets linked to the route via status")
	callback = helpers.GetGatewayIsLinkedCallback(ctx, t, gatewayClient, gatewayapi.HTTPProtocolType, ns.Name, httpRoute.Name)
	require.Eventually(t, callback, ingressWait, waitTick)

	t.Log("verifying that creating the Gateway again triggers reconciliation of HTTPRoutes and the route becomes available again")
	helpers.EventuallyGETPath(t, proxyHTTPURL, proxyHTTPURL.Host, "/test-http-route-essentials", nil, http.StatusOK, "<title>httpbin.org</title>", emptyHeaderSet, ingressWait, waitTick)

	t.Log("deleting both GatewayClass and Gateway rapidly")
	require.NoError(t, gatewayClient.GatewayV1().GatewayClasses().Delete(ctx, gwc.Name, metav1.DeleteOptions{}))
	require.NoError(t, gatewayClient.GatewayV1().Gateways(ns.Name).Delete(ctx, gateway.Name, metav1.DeleteOptions{}))

	t.Log("verifying that the Gateway gets unlinked from the route via status")
	callback = helpers.GetGatewayIsUnlinkedCallback(ctx, t, gatewayClient, gatewayapi.HTTPProtocolType, ns.Name, httpRoute.Name)
	require.Eventually(t, callback, ingressWait, waitTick)

	t.Log("verifying that the data-plane configuration from the HTTPRoute does not get orphaned with the GatewayClass and Gateway gone")
	helpers.EventuallyGETPath(t, proxyHTTPURL, proxyHTTPURL.Host, "/test-http-route-essentials", nil, http.StatusNotFound, "", emptyHeaderSet, ingressWait, waitTick)

	t.Log("testing port matching....")
	t.Log("putting the Gateway back")
	_, err = helpers.DeployGateway(ctx, gatewayClient, ns.Name, gatewayClassName, func(gw *gatewayapi.Gateway) {
		gw.Name = gatewayName
	})
	require.NoError(t, err)
	t.Log("putting the GatewayClass back")
	_, err = helpers.DeployGatewayClass(ctx, gatewayClient, gatewayClassName)
	require.NoError(t, err)

	httpRrouteName := httpRoute.Name

	// Set the Port in ParentRef which does not have a matching listener in Gateway.
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		httpRoute, err = gatewayClient.GatewayV1().HTTPRoutes(ns.Name).Get(ctx, httpRrouteName, metav1.GetOptions{})
		if !assert.NoError(c, err) {
			return
		}
		httpRoute.Spec.ParentRefs[0].Port = lo.ToPtr(gatewayapi.PortNumber(81))
		httpRoute, err = gatewayClient.GatewayV1().HTTPRoutes(ns.Name).Update(ctx, httpRoute, metav1.UpdateOptions{})
		assert.NoError(c, err)
	}, time.Minute, time.Second)

	t.Log("verifying that the HTTPRoute has the Condition 'Accepted' set to 'False' when it specified a port not existent in Gateway")
	require.Eventually(t, HTTPRouteMatchesAcceptedCallback(t, gatewayClient, httpRoute, false, gatewayapi.RouteReasonNoMatchingParent), statusWait, waitTick)
}

func TestHTTPRouteMultipleServices(t *testing.T) {
	ctx := t.Context()

	ns, cleaner := helpers.Setup(ctx, t, env)

	t.Log("getting a gateway client")
	gatewayClient, err := gatewayclient.NewForConfig(env.Cluster().Config())
	require.NoError(t, err)

	t.Log("deploying a new gatewayClass")
	gatewayClassName := uuid.NewString()
	gwc, err := helpers.DeployGatewayClass(ctx, gatewayClient, gatewayClassName)
	require.NoError(t, err)
	cleaner.Add(gwc)

	t.Log("deploying a new gateway")
	gatewayName := uuid.NewString()
	gateway, err := helpers.DeployGateway(ctx, gatewayClient, ns.Name, gatewayClassName, func(gw *gatewayapi.Gateway) {
		gw.Name = gatewayName
	})
	require.NoError(t, err)
	cleaner.Add(gateway)

	t.Log("deploying a minimal HTTP container deployment to test Ingress routes")
	container1 := generators.NewContainer("httpbin", test.HTTPBinImage, test.HTTPBinPort)
	deployment1 := generators.NewDeploymentForContainer(container1)
	deployment1, err = env.Cluster().Client().AppsV1().Deployments(ns.Name).Create(ctx, deployment1, metav1.CreateOptions{})
	require.NoError(t, err)

	t.Log("deploying an extra minimal HTTP container deployment to test multiple backendRefs")
	container2 := generators.NewContainer("nginx", "nginx", 80)
	deployment2 := generators.NewDeploymentForContainer(container2)
	deployment2, err = env.Cluster().Client().AppsV1().Deployments(ns.Name).Create(ctx, deployment2, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(deployment1)
	cleaner.Add(deployment2)

	t.Logf("exposing deployment %s via service", deployment1.Name)
	service1 := generators.NewServiceForDeployment(deployment1, corev1.ServiceTypeLoadBalancer)
	_, err = env.Cluster().Client().CoreV1().Services(ns.Name).Create(ctx, service1, metav1.CreateOptions{})
	require.NoError(t, err)

	t.Logf("exposing deployment %s via service", deployment2.Name)
	service2 := generators.NewServiceForDeployment(deployment2, corev1.ServiceTypeLoadBalancer)
	_, err = env.Cluster().Client().CoreV1().Services(ns.Name).Create(ctx, service2, metav1.CreateOptions{})
	require.NoError(t, err)
	// service3 has an annotation the others don't. we expect the controller to skip rules that have different annotations
	service3 := generators.NewServiceForDeployment(deployment2, corev1.ServiceTypeLoadBalancer)
	service3.Annotations = map[string]string{annotations.AnnotationPrefix + annotations.HostHeaderKey: "example.com"}
	service3.Name = "nginx-host"
	_, err = env.Cluster().Client().CoreV1().Services(ns.Name).Create(ctx, service3, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(service1)
	cleaner.Add(service2)
	cleaner.Add(service3)

	t.Log("adding an HTTPRoute with multi-backend rules")
	var httpbinWeight int32 = 75
	var nginxWeight int32 = 25
	httpPort := gatewayapi.PortNumber(80)
	pathMatchPrefix := gatewayapi.PathMatchPathPrefix
	httpRoute := &gatewayapi.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name: uuid.NewString(),
			Annotations: map[string]string{
				annotations.AnnotationPrefix + annotations.StripPathKey: "true",
				annotations.AnnotationPrefix + annotations.PluginsKey:   "correlation",
			},
		},
		Spec: gatewayapi.HTTPRouteSpec{
			CommonRouteSpec: gatewayapi.CommonRouteSpec{
				ParentRefs: []gatewayapi.ParentReference{{
					Name: gatewayapi.ObjectName(gateway.Name),
				}},
			},
			Rules: []gatewayapi.HTTPRouteRule{
				{
					Matches: []gatewayapi.HTTPRouteMatch{
						{
							Path: &gatewayapi.HTTPPathMatch{
								Type:  &pathMatchPrefix,
								Value: kong.String("/test-http-route-multiple-services"),
							},
						},
					},
					BackendRefs: []gatewayapi.HTTPBackendRef{
						{
							BackendRef: gatewayapi.BackendRef{
								BackendObjectReference: gatewayapi.BackendObjectReference{
									Name: gatewayapi.ObjectName(service1.Name),
									Port: &httpPort,
									Kind: util.StringToGatewayAPIKindPtr("Service"),
								},
								Weight: &httpbinWeight,
							},
						},
						{
							BackendRef: gatewayapi.BackendRef{
								BackendObjectReference: gatewayapi.BackendObjectReference{
									Name: gatewayapi.ObjectName(service2.Name),
									Port: &httpPort,
									Kind: util.StringToGatewayAPIKindPtr("Service"),
								},
								Weight: &nginxWeight,
							},
						},
					},
				},
				{
					Matches: []gatewayapi.HTTPRouteMatch{
						{
							Path: &gatewayapi.HTTPPathMatch{
								Type:  &pathMatchPrefix,
								Value: kong.String("/test-http-route-multiple-services-broken"),
							},
						},
					},
					BackendRefs: []gatewayapi.HTTPBackendRef{
						{
							BackendRef: gatewayapi.BackendRef{
								BackendObjectReference: gatewayapi.BackendObjectReference{
									Name: gatewayapi.ObjectName(service1.Name),
									Port: &httpPort,
									Kind: util.StringToGatewayAPIKindPtr("Service"),
								},
								Weight: &httpbinWeight,
							},
						},
						{
							BackendRef: gatewayapi.BackendRef{
								BackendObjectReference: gatewayapi.BackendObjectReference{
									Name: gatewayapi.ObjectName(service3.Name),
									Port: &httpPort,
									Kind: util.StringToGatewayAPIKindPtr("Service"),
								},
								Weight: &nginxWeight,
							},
						},
					},
				},
			},
		},
	}
	httpRoute, err = gatewayClient.GatewayV1().HTTPRoutes(ns.Name).Create(ctx, httpRoute, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(httpRoute)

	t.Log("verifying that both backends are ready to receive traffic")
	httpbinRespContent := "<title>httpbin.org</title>"
	nginxRespContent := "<title>Welcome to nginx!</title>"
	helpers.EventuallyGETPath(t, proxyHTTPURL, proxyHTTPURL.Host, "/test-http-route-multiple-services", nil, http.StatusOK, httpbinRespContent, emptyHeaderSet, ingressWait, waitTick)
	helpers.EventuallyGETPath(t, proxyHTTPURL, proxyHTTPURL.Host, "/test-http-route-multiple-services", nil, http.StatusOK, nginxRespContent, emptyHeaderSet, ingressWait, waitTick)

	t.Log("verifying that both backends receive requests according to weighted distribution")
	httpbinRespName := "httpbin-resp"
	nginxRespName := "nginx-resp"
	toleranceDelta := 0.2
	expectedRespRatio := map[string]int{
		httpbinRespName: int(httpbinWeight),
		nginxRespName:   int(nginxWeight),
	}
	weightedLoadBalancingTestConfig := helpers.CountHTTPResponsesConfig{
		Method:      http.MethodGet,
		Host:        proxyHTTPURL.Host,
		Path:        "test-http-route-multiple-services",
		Headers:     emptyHeaderSet,
		Duration:    5 * time.Second,
		RequestTick: 50 * time.Millisecond,
	}
	respCounter := helpers.CountHTTPGetResponses(t,
		proxyHTTPURL,
		weightedLoadBalancingTestConfig,
		helpers.MatchRespByStatusAndContent(httpbinRespName, http.StatusOK, httpbinRespContent),
		helpers.MatchRespByStatusAndContent(nginxRespName, http.StatusOK, nginxRespContent),
	)
	assert.InDeltaMapValues(t,
		helpers.DistributionOfMapValues(respCounter),
		helpers.DistributionOfMapValues(expectedRespRatio),
		toleranceDelta,
		"Response distribution does not match expected distribution within %f%% delta,"+
			" request-count=%v, expected-ratio=%v",
		toleranceDelta*100, respCounter, expectedRespRatio,
	)

	t.Log("verifying that misconfigured service rules are _not_ routed")
	helpers.EventuallyGETPath(t, proxyHTTPURL, proxyHTTPURL.Host, "/test-http-route-multiple-services-broken", nil, http.StatusNotFound, "", emptyHeaderSet, ingressWait, waitTick)
}

func TestHTTPRouteFilterHosts(t *testing.T) {
	ctx := t.Context()

	ns, cleaner := helpers.Setup(ctx, t, env)

	listenerHostname := gatewayapi.Hostname("test.specific.io")

	t.Log("getting a gateway client")
	gatewayClient, err := gatewayclient.NewForConfig(env.Cluster().Config())
	require.NoError(t, err)

	t.Log("deploying a new gatewayClass")
	gatewayClassName := uuid.NewString()
	gwc, err := helpers.DeployGatewayClass(ctx, gatewayClient, gatewayClassName)
	require.NoError(t, err)
	cleaner.Add(gwc)

	t.Log("deploying a new gateway with specified hostname")
	gatewayName := uuid.NewString()
	gateway, err := helpers.DeployGateway(ctx, gatewayClient, ns.Name, gatewayClassName, func(gw *gatewayapi.Gateway) {
		gw.Name = gatewayName
		for i := range gw.Spec.Listeners {
			gw.Spec.Listeners[i].Hostname = &listenerHostname
		}
	})
	require.NoError(t, err)
	cleaner.Add(gateway)

	t.Log("deploying a minimal HTTP container deployment to test Ingress routes")
	container := generators.NewContainer("httpbin", test.HTTPBinImage, test.HTTPBinPort)
	deployment := generators.NewDeploymentForContainer(container)
	deployment, err = env.Cluster().Client().AppsV1().Deployments(ns.Name).Create(ctx, deployment, metav1.CreateOptions{})
	require.NoError(t, err)

	t.Logf("exposing deployment %s via service", deployment.Name)
	service := generators.NewServiceForDeployment(deployment, corev1.ServiceTypeLoadBalancer)
	_, err = env.Cluster().Client().CoreV1().Services(ns.Name).Create(ctx, service, metav1.CreateOptions{})
	require.NoError(t, err)

	t.Logf("creating an httproute with a same hostname and another unmatched hostname")
	httpRoute := &gatewayapi.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name: uuid.NewString(),
			Annotations: map[string]string{
				annotations.AnnotationPrefix + annotations.StripPathKey: "true",
			},
		},
		Spec: gatewayapi.HTTPRouteSpec{
			CommonRouteSpec: gatewayapi.CommonRouteSpec{
				ParentRefs: []gatewayapi.ParentReference{{
					Name: gatewayapi.ObjectName(gateway.Name),
				}},
			},
			Hostnames: []gatewayapi.Hostname{
				gatewayapi.Hostname("test.specific.io"),
				gatewayapi.Hostname("another.specific.io"),
			},
			Rules: []gatewayapi.HTTPRouteRule{{
				Matches: []gatewayapi.HTTPRouteMatch{
					builder.NewHTTPRouteMatch().WithPathPrefix("/test-http-route-filter-hosts").Build(),
				},
				BackendRefs: []gatewayapi.HTTPBackendRef{
					builder.NewHTTPBackendRef(service.Name).WithPort(80).Build(),
				},
			}},
		},
	}
	hClient := gatewayClient.GatewayV1().HTTPRoutes(ns.Name)
	httpRoute, err = hClient.Create(ctx, httpRoute, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(httpRoute)

	// testGetByHost tries to get the test path with specified host in request,
	// and returns true if 200 returned.
	testGetByHost := func(t *testing.T, host string) error {
		t.Helper()
		req := helpers.MustHTTPRequest(t, http.MethodGet, host, "/test-http-route-filter-hosts", nil)
		resp, err := helpers.DefaultHTTPClient(helpers.WithResolveHostTo(proxyHTTPURL.Host)).Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("expected status code 200, got %d", resp.StatusCode)
		}
		return nil
	}

	t.Logf("test host matched hostname in listeners")
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.NoError(c, testGetByHost(t, "test.specific.io"))
	}, ingressWait, waitTick)
	t.Logf("test host matched in httproute, but not in listeners")
	require.Error(t, testGetByHost(t, "another.specific.io"))

	t.Logf("update hostnames in httproute to wildcard")
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		httpRoute, err = hClient.Get(ctx, httpRoute.Name, metav1.GetOptions{})
		if !assert.NoErrorf(c, err, "failed getting the HTTPRoute %s", httpRoute.Name) {
			return
		}
		httpRoute.Spec.Hostnames = []gatewayapi.Hostname{
			gatewayapi.Hostname("*.specific.io"),
		}
		httpRoute, err = hClient.Update(ctx, httpRoute, metav1.UpdateOptions{})
		assert.NoErrorf(c, err, "failed updating the HTTPRoute %s", httpRoute.Name)
	}, test.RequestTimeout, 100*time.Millisecond)
	t.Logf("test host matched hostname in listeners")
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.NoError(c, testGetByHost(t, "test.specific.io"))
	}, ingressWait, waitTick)
	t.Logf("test host matched in httproute, but not in listeners")
	require.Error(t, testGetByHost(t, "another2.specific.io"))

	t.Logf("update hostname in httproute to an unmatched host")
	httpRoute, err = hClient.Get(ctx, httpRoute.Name, metav1.GetOptions{})
	require.NoError(t, err)
	httpRoute.Spec.Hostnames = []gatewayapi.Hostname{
		gatewayapi.Hostname("another.specific.io"),
	}
	httpRoute, err = hClient.Update(ctx, httpRoute, metav1.UpdateOptions{})
	require.NoError(t, err)
	t.Logf("status of httproute should contain an 'Accepted' condition with 'False' status")
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		currentHTTPRoute, err := hClient.Get(ctx, httpRoute.Name, metav1.GetOptions{})
		if !assert.NoError(c, err) {
			return
		}
		for _, parent := range currentHTTPRoute.Status.Parents {
			for _, condition := range parent.Conditions {
				if assert.True(
					c,
					condition.Type == string(gatewayapi.RouteReasonAccepted) && condition.Status == metav1.ConditionFalse,
				) {
					return
				}
			}
		}
	}, ingressWait, waitTick,
		func() string {
			currentHTTPRoute, err := hClient.Get(ctx, httpRoute.Name, metav1.GetOptions{})
			if err != nil {
				return err.Error()
			}
			return fmt.Sprintf("current status of HTTPRoute %s/%s:%v", httpRoute.Namespace, httpRoute.Name, currentHTTPRoute.Status)
		}())
	t.Logf("test host matched in httproute, but not in listeners")
	require.Error(t, testGetByHost(t, "another.specific.io"))
}
