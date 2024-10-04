package integration

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kong/kubernetes-testing-framework/pkg/clusters"
	"github.com/kong/kubernetes-testing-framework/pkg/utils/kubernetes/generators"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
	operatorv1beta1 "github.com/kong/gateway-operator/api/v1beta1"
	"github.com/kong/gateway-operator/controller/dataplane"
	"github.com/kong/gateway-operator/pkg/consts"
	testutils "github.com/kong/gateway-operator/pkg/utils/test"
	"github.com/kong/gateway-operator/test/helpers"

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
)

func TestKongPluginInstallationEssentials(t *testing.T) {
	namespace, cleaner := helpers.SetupTestEnv(t, GetCtx(), GetEnv())

	const registryUrl = "northamerica-northeast1-docker.pkg.dev/k8s-team-playground/"

	const pluginInvalidLayersImage = registryUrl + "plugin-example/invalid-layers"

	const pluginMyHeaderImage = registryUrl + "plugin-example/myheader"
	expectedHeadersForMyHeader := http.Header{"myheader": {"roar"}}

	const pluginMyHeader2Image = registryUrl + "plugin-example-private/myheader-2"
	expectedHeadersForMyHeader2 := http.Header{"newheader": {"amazing"}}

	t.Log("deploying an invalid KongPluginInstallation resource")
	kpiPublicNN := k8stypes.NamespacedName{
		Name:      "test-kpi",
		Namespace: namespace.Name,
	}
	kpiPublic := &operatorv1alpha1.KongPluginInstallation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kpiPublicNN.Name,
			Namespace: kpiPublicNN.Namespace,
		},
		Spec: operatorv1alpha1.KongPluginInstallationSpec{
			Image: pluginInvalidLayersImage,
		},
	}
	kpiPublic, err := GetClients().OperatorClient.ApisV1alpha1().KongPluginInstallations(namespace.Name).Create(GetCtx(), kpiPublic, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(kpiPublic)

	t.Log("waiting for the KongPluginInstallation resource to be rejected, because of the invalid image")
	checkKongPluginInstallationConditions(
		t, kpiPublicNN, metav1.ConditionFalse,
		fmt.Sprintf(`problem with the image: "%s" error: expected exactly one layer with plugin, found 2 layers`, pluginInvalidLayersImage),
	)

	t.Log("deploy Gateway with example service and HTTPRoute")
	ip, gatewayConfigNN, httpRouteNN := deployGatewayWithKPI(t, cleaner, namespace.Name)
	t.Log("attach broken KPI to the Gateway")
	attachKPI(t, gatewayConfigNN, kpiPublicNN)
	t.Log("ensure that status of the DataPlane is not ready with proper description of the issue")
	checkDataPlaneStatus(
		t, namespace.Name, metav1.ConditionFalse, dataplane.DataPlaneConditionReferencedResourcesNotAvailable,
		fmt.Sprintf("something wrong with referenced KongPluginInstallation %s, please check it", client.ObjectKeyFromObject(kpiPublic)),
	)

	t.Log("updating KongPluginInstallation resource to a valid image")
	kpiPublic, err = GetClients().OperatorClient.ApisV1alpha1().KongPluginInstallations(kpiPublicNN.Namespace).Get(GetCtx(), kpiPublicNN.Name, metav1.GetOptions{})
	kpiPublic.Spec.Image = pluginMyHeaderImage
	require.NoError(t, err)
	_, err = GetClients().OperatorClient.ApisV1alpha1().KongPluginInstallations(kpiPublicNN.Namespace).Update(GetCtx(), kpiPublic, metav1.UpdateOptions{})
	require.NoError(t, err)
	t.Log("waiting for the KongPluginInstallation resource to be accepted")
	checkKongPluginInstallationConditions(t, kpiPublicNN, metav1.ConditionTrue, "plugin successfully saved in cluster as ConfigMap")

	t.Log("waiting for the DataPlane that reference KongPluginInstallation to be ready")
	checkDataPlaneStatus(t, namespace.Name, metav1.ConditionTrue, consts.ResourceReadyReason, "")
	t.Log("attach configured KongPlugin with KongPluginInstallation to the HTTPRoute")
	attachKongPluginBasedOnKPIToRoute(t, cleaner, httpRouteNN, kpiPublicNN)

	t.Log("verify that plugin is properly configured and works")
	verifyCustomPlugins(t, ip, expectedHeadersForMyHeader)

	if registryCreds := GetKongPluginImageRegistryCredentialsForTests(); registryCreds != "" {
		// Create kpiPrivateNamespace with K8s client to check cross-namespace capabilities.
		t.Log("add additional KongPluginInstallation resource from a private image")
		kpiPrivateNN := k8stypes.NamespacedName{
			Name:      "test-kpi-private",
			Namespace: createRandomNamespace(t),
		}
		kpiPrivate := &operatorv1alpha1.KongPluginInstallation{
			ObjectMeta: metav1.ObjectMeta{
				Name:      kpiPrivateNN.Name,
				Namespace: kpiPrivateNN.Namespace,
			},
			Spec: operatorv1alpha1.KongPluginInstallationSpec{
				Image: pluginMyHeader2Image,
			},
		}
		require.NoError(t, err)
		_, err = GetClients().OperatorClient.ApisV1alpha1().KongPluginInstallations(kpiPrivateNN.Namespace).Create(GetCtx(), kpiPrivate, metav1.CreateOptions{})
		require.NoError(t, err)
		t.Log("waiting for the KongPluginInstallation resource to be reconciled and report unauthenticated request")
		checkKongPluginInstallationConditions(
			t, kpiPrivateNN, metav1.ConditionFalse, "response status code 403: denied: Unauthenticated request. Unauthenticated requests do not have permission",
		)

		t.Log("update KongPluginInstallation resource with credentials reference in other namespace than KongPluginInstallation")
		namespaceForSecret := createRandomNamespace(t)
		kpiPrivate, err = GetClients().OperatorClient.ApisV1alpha1().KongPluginInstallations(kpiPrivateNN.Namespace).Get(GetCtx(), kpiPrivateNN.Name, metav1.GetOptions{})
		require.NoError(t, err)
		const kindSecret = gatewayv1.Kind("Secret")
		secretRef := gatewayv1.SecretObjectReference{
			Kind:      lo.ToPtr(kindSecret),
			Namespace: lo.ToPtr(gatewayv1.Namespace(namespaceForSecret)),
			Name:      "kong-plugin-image-registry-credentials",
		}
		kpiPrivate.Spec.ImagePullSecretRef = &secretRef
		_, err = GetClients().OperatorClient.ApisV1alpha1().KongPluginInstallations(kpiPrivateNN.Namespace).Update(GetCtx(), kpiPrivate, metav1.UpdateOptions{})
		require.NoError(t, err)
		t.Log("waiting for the KongPluginInstallation resource to be reconciled and report missing ReferenceGrant for the Secret with credentials")
		checkKongPluginInstallationConditions(
			t, kpiPrivateNN, metav1.ConditionFalse, fmt.Sprintf("Secret %s/%s reference not allowed by any ReferenceGrant", *secretRef.Namespace, secretRef.Name),
		)
		attachKPI(t, gatewayConfigNN, kpiPrivateNN)
		checkDataPlaneStatus(
			t, namespace.Name, metav1.ConditionFalse, dataplane.DataPlaneConditionReferencedResourcesNotAvailable,
			fmt.Sprintf("something wrong with referenced KongPluginInstallation %s, please check it", client.ObjectKeyFromObject(kpiPrivate)),
		)

		t.Log("add missing ReferenceGrant for the Secret with credentials")
		refGrant := &gatewayv1beta1.ReferenceGrant{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kong-plugin-image-registry-credentials",
				Namespace: namespaceForSecret,
			},
			Spec: gatewayv1beta1.ReferenceGrantSpec{
				To: []gatewayv1beta1.ReferenceGrantTo{
					{
						Kind: kindSecret,
						Name: lo.ToPtr(secretRef.Name),
					},
				},
				From: []gatewayv1beta1.ReferenceGrantFrom{
					{
						Group:     gatewayv1.Group(operatorv1alpha1.SchemeGroupVersion.Group),
						Kind:      gatewayv1.Kind("KongPluginInstallation"),
						Namespace: gatewayv1.Namespace(kpiPrivate.Namespace),
					},
				},
			},
		}
		_, err = GetClients().GatewayClient.GatewayV1beta1().ReferenceGrants(namespaceForSecret).Create(GetCtx(), refGrant, metav1.CreateOptions{})
		require.NoError(t, err)

		t.Log("waiting for the KongPluginInstallation resource to be reconciled and report missing Secret with credentials")
		checkKongPluginInstallationConditions(
			t, kpiPrivateNN, metav1.ConditionFalse,
			fmt.Sprintf(`referenced Secret "%s/%s" not found`, *secretRef.Namespace, secretRef.Name),
		)
		checkDataPlaneStatus(
			t, namespace.Name, metav1.ConditionFalse, dataplane.DataPlaneConditionReferencedResourcesNotAvailable,
			fmt.Sprintf("something wrong with referenced KongPluginInstallation %s, please check it", client.ObjectKeyFromObject(kpiPrivate)),
		)

		t.Log("add missing Secret with credentials")
		secret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: string(secretRef.Name),
			},
			Type: corev1.SecretTypeDockerConfigJson,
			StringData: map[string]string{
				".dockerconfigjson": registryCreds,
			},
		}
		_, err = GetClients().K8sClient.CoreV1().Secrets(string(*secretRef.Namespace)).Create(GetCtx(), &secret, metav1.CreateOptions{})
		require.NoError(t, err)
		t.Log("waiting for the KongPluginInstallation resource to be reconciled successfully")
		checkKongPluginInstallationConditions(
			t, kpiPrivateNN, metav1.ConditionTrue, "plugin successfully saved in cluster as ConfigMap",
		)

		t.Log("waiting for the DataPlane that reference KongPluginInstallation to be ready")
		checkDataPlaneStatus(t, namespace.Name, metav1.ConditionTrue, consts.ResourceReadyReason, "")
		t.Log("attach configured KongPlugin to the HTTPRoute")
		attachKongPluginBasedOnKPIToRoute(t, cleaner, httpRouteNN, kpiPrivateNN)
		t.Log("verify that plugin is properly configured and works")
		verifyCustomPlugins(t, ip, expectedHeadersForMyHeader, expectedHeadersForMyHeader2)
	} else {
		t.Log("skipping private image test - no credentials provided")
	}
}

func deployGatewayWithKPI(
	t *testing.T, cleaner *clusters.Cleaner, namespace string,
) (gatewayIPAddress string, gatewayConfigNN, httpRouteNN k8stypes.NamespacedName) {
	// NOTE: Disable webhook for KIC, because it checks for the plugin in Kong Gateway and rejects,
	// thus it requires strict order of deployment which is not guaranteed.
	gatewayConfig := helpers.GenerateGatewayConfiguration(namespace, helpers.WithControlPlaneWebhookDisabled())
	t.Logf("deploying GatewayConfiguration %s/%s", gatewayConfig.Namespace, gatewayConfig.Name)
	gatewayConfig, err := GetClients().OperatorClient.ApisV1beta1().GatewayConfigurations(namespace).Create(GetCtx(), gatewayConfig, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gatewayConfig)

	gatewayClass := helpers.MustGenerateGatewayClass(t, gatewayv1.ParametersReference{
		Group:     gatewayv1.Group(operatorv1beta1.SchemeGroupVersion.Group),
		Kind:      gatewayv1.Kind("GatewayConfiguration"),
		Namespace: (*gatewayv1.Namespace)(&gatewayConfig.Namespace),
		Name:      gatewayConfig.Name,
	})
	t.Logf("deploying GatewayClass %s", gatewayClass.Name)
	gatewayClass, err = GetClients().GatewayClient.GatewayV1().GatewayClasses().Create(GetCtx(), gatewayClass, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gatewayClass)

	gatewayNSN := k8stypes.NamespacedName{
		Name:      uuid.NewString(),
		Namespace: namespace,
	}
	gateway := helpers.GenerateGateway(gatewayNSN, gatewayClass)
	t.Logf("deploying Gateway %s/%s", gateway.Namespace, gateway.Name)
	gateway, err = GetClients().GatewayClient.GatewayV1().Gateways(namespace).Create(GetCtx(), gateway, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gateway)

	t.Logf("verifying Gateway %s/%s gets an IP address", gateway.Namespace, gateway.Name)
	require.Eventually(t, testutils.GatewayIPAddressExist(t, GetCtx(), gatewayNSN, clients), 4*testutils.SubresourceReadinessWait, time.Second)
	gateway = testutils.MustGetGateway(t, GetCtx(), gatewayNSN, clients)

	t.Log("deploying backend deployment (httpbin) of HTTPRoute")
	container := generators.NewContainer("httpbin", testutils.HTTPBinImage, 80)
	deployment := generators.NewDeploymentForContainer(container)
	deployment, err = GetEnv().Cluster().Client().AppsV1().Deployments(namespace).Create(GetCtx(), deployment, metav1.CreateOptions{})
	require.NoError(t, err)

	t.Logf("exposing deployment %s via service", deployment.Name)
	service := generators.NewServiceForDeployment(deployment, corev1.ServiceTypeClusterIP)
	_, err = GetEnv().Cluster().Client().CoreV1().Services(namespace).Create(GetCtx(), service, metav1.CreateOptions{})
	require.NoError(t, err)

	httpRoute := helpers.GenerateHTTPRoute(namespace, gateway.Name, service.Name)
	t.Logf("creating httproute %s/%s to access deployment %s via kong", httpRoute.Namespace, httpRoute.Name, deployment.Name)
	require.EventuallyWithT(t,
		func(c *assert.CollectT) {
			result, err := GetClients().GatewayClient.GatewayV1().HTTPRoutes(namespace).Create(GetCtx(), httpRoute, metav1.CreateOptions{})
			if err != nil {
				t.Logf("failed to deploy httproute: %v", err)
				c.Errorf("failed to deploy httproute: %v", err)
				return
			}
			cleaner.Add(result)
		},
		testutils.DefaultIngressWait, testutils.WaitIngressTick,
	)

	return gateway.Status.Addresses[0].Value, client.ObjectKeyFromObject(gatewayConfig), client.ObjectKeyFromObject(httpRoute)
}

func checkKongPluginInstallationConditions(
	t *testing.T,
	namespacedName k8stypes.NamespacedName,
	conditionStatus metav1.ConditionStatus,
	expectedMessage string,
) {
	t.Helper()
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		kpi, err := GetClients().OperatorClient.ApisV1alpha1().KongPluginInstallations(namespacedName.Namespace).Get(GetCtx(), namespacedName.Name, metav1.GetOptions{})
		if !assert.NoError(c, err) {
			return
		}
		if !assert.NotEmpty(c, kpi.Status.Conditions) {
			return
		}
		status := kpi.Status.Conditions[0]
		assert.EqualValues(c, operatorv1alpha1.KongPluginInstallationConditionStatusAccepted, status.Type)
		assert.EqualValues(c, conditionStatus, status.Status)
		if conditionStatus == metav1.ConditionTrue {
			assert.EqualValues(c, operatorv1alpha1.KongPluginInstallationReasonReady, status.Reason)
		} else {
			assert.EqualValues(c, operatorv1alpha1.KongPluginInstallationReasonFailed, status.Reason)
		}
		assert.Contains(c, status.Message, expectedMessage)
	}, 15*time.Second, time.Second)
}

func attachKPI(t *testing.T, gatewayConfigNN k8stypes.NamespacedName, kpiNN k8stypes.NamespacedName) {
	t.Helper()
	gatewayConfig, err := GetClients().OperatorClient.ApisV1beta1().GatewayConfigurations(gatewayConfigNN.Namespace).Get(GetCtx(), gatewayConfigNN.Name, metav1.GetOptions{})
	require.NoError(t, err)
	gatewayConfig.Spec.DataPlaneOptions.PluginsToInstall = append(gatewayConfig.Spec.DataPlaneOptions.PluginsToInstall, operatorv1beta1.NamespacedName(kpiNN))
	_, err = GetClients().OperatorClient.ApisV1beta1().GatewayConfigurations(gatewayConfigNN.Namespace).Update(GetCtx(), gatewayConfig, metav1.UpdateOptions{})
	require.NoError(t, err)
}

func attachKongPluginBasedOnKPIToRoute(t *testing.T, cleaner *clusters.Cleaner, httpRouteNN, kpiNN k8stypes.NamespacedName) {
	t.Helper()

	kongPluginName := kpiNN.Name + "-plugin"
	// To have it in the same namespace as the HTTPRoute to which it is attached.
	kongPluginNamespace := httpRouteNN.Namespace
	kongPlugin := configurationv1.KongPlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kongPluginName,
			Namespace: kongPluginNamespace,
		},
		PluginName: kpiNN.Name,
	}
	_, err := GetClients().ConfigurationClient.ConfigurationV1().KongPlugins(kongPluginNamespace).Create(GetCtx(), &kongPlugin, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(&kongPlugin)

	t.Logf("attaching KongPlugin %s to HTTPRoute %s", kongPluginName, httpRouteNN)
	require.Eventually(t,
		testutils.HTTPRouteUpdateEventually(t, GetCtx(), httpRouteNN, clients, func(h *gatewayv1.HTTPRoute) {
			const kpAnnotation = "konghq.com/plugins"
			h.Annotations[kpAnnotation] = strings.Join(
				append(strings.Split(h.Annotations[kpAnnotation], ","), kongPluginName), ",",
			)
		}),
		time.Minute, 250*time.Millisecond,
	)
}

func checkDataPlaneStatus(
	t *testing.T,
	namespace string,
	expectedConditionStatus metav1.ConditionStatus,
	expectedConditionReason consts.ConditionReason,
	expectedConditionMessage string,
) {
	t.Helper()
	var dp operatorv1beta1.DataPlane
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		dps, err := GetClients().OperatorClient.ApisV1beta1().DataPlanes(namespace).List(GetCtx(), metav1.ListOptions{})
		if !assert.NoError(c, err) {
			return
		}
		if assert.Len(c, dps.Items, 1) {
			dp = dps.Items[0]
		}
		if !assert.Len(c, dp.Status.Conditions, 1) {
			return
		}

		condition := dp.Status.Conditions[0]
		assert.EqualValues(c, consts.ReadyType, condition.Type)
		assert.EqualValues(c, expectedConditionStatus, condition.Status)
		assert.EqualValues(c, expectedConditionReason, condition.Reason)
		assert.Equal(c, expectedConditionMessage, condition.Message)
	}, 2*time.Minute, time.Second)
}

func verifyCustomPlugins(t *testing.T, ip string, expectedHeaders ...http.Header) {
	t.Helper()
	httpClient, err := helpers.CreateHTTPClient(nil, "")
	require.NoError(t, err)
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		resp, err := httpClient.Get(fmt.Sprintf("http://%s/test", ip))
		if !assert.NoError(c, err) {
			return
		}
		defer resp.Body.Close()
		if !assert.Equal(c, http.StatusOK, resp.StatusCode) {
			return
		}
		for _, h := range expectedHeaders {
			for k, v := range h {
				assert.Equal(c, v, resp.Header.Values(k))
			}
		}
	}, 15*time.Second, time.Second)
}

func createRandomNamespace(t *testing.T) string {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: uuid.NewString(),
		},
	}
	_, err := GetClients().K8sClient.CoreV1().Namespaces().Create(GetCtx(), namespace, metav1.CreateOptions{})
	require.NoError(t, err)
	return namespace.Name
}
