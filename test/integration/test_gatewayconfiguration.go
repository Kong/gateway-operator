package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	operatorv1beta1 "github.com/kong/gateway-operator/api/v1beta1"
	gwtypes "github.com/kong/gateway-operator/internal/types"
	"github.com/kong/gateway-operator/pkg/consts"
	gatewayutils "github.com/kong/gateway-operator/pkg/utils/gateway"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
	testutils "github.com/kong/gateway-operator/pkg/utils/test"
	"github.com/kong/gateway-operator/pkg/vars"
	"github.com/kong/gateway-operator/test/helpers"
)

const (
	testEnvVar         = "KONG_INTEGRATION_TESTS"
	testEnvVal         = "TEST_VALUE"
	testEnvVarFromName = "KONG_INTEGRATION_TESTS_FROM"
	testEnvVarFromKV   = "dzhambul"
)

func TestGatewayConfigurationEssentials(t *testing.T) {
	t.Parallel()
	namespace, cleaner := helpers.SetupTestEnv(t, GetCtx(), GetEnv())

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace.Name,
			Name:      uuid.NewString(),
		},
		Data: map[string]string{
			testEnvVarFromKV: testEnvVarFromKV,
		},
	}
	configMap, err := GetEnv().Cluster().Client().CoreV1().ConfigMaps(namespace.Name).Create(GetCtx(), configMap, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(configMap)

	t.Log("deploying a GatewayConfiguration resource")
	gatewayConfig := &operatorv1beta1.GatewayConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace.Name,
			Name:      uuid.NewString(),
		},
		Spec: operatorv1beta1.GatewayConfigurationSpec{
			DataPlaneOptions: &operatorv1beta1.GatewayConfigDataPlaneOptions{
				Deployment: operatorv1beta1.DataPlaneDeploymentOptions{
					DeploymentOptions: operatorv1beta1.DeploymentOptions{
						PodTemplateSpec: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  consts.DataPlaneProxyContainerName,
										Image: helpers.GetDefaultDataPlaneImage(),
										Env: []corev1.EnvVar{
											{
												Name:  testEnvVar,
												Value: testEnvVal,
											},
											{
												Name: testEnvVarFromName,
												ValueFrom: &corev1.EnvVarSource{
													ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
														LocalObjectReference: corev1.LocalObjectReference{
															Name: configMap.Name,
														},
														Key: testEnvVarFromKV,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			ControlPlaneOptions: &operatorv1beta1.ControlPlaneOptions{
				Deployment: operatorv1beta1.ControlPlaneDeploymentOptions{
					PodTemplateSpec: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  consts.ControlPlaneControllerContainerName,
									Image: consts.DefaultControlPlaneImage,
									Env: []corev1.EnvVar{
										{
											Name:  testEnvVar,
											Value: testEnvVal,
										},
										{
											Name: testEnvVarFromName,
											ValueFrom: &corev1.EnvVarSource{
												ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: configMap.Name,
													},
													Key: testEnvVarFromKV,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	gatewayConfig, err = GetClients().OperatorClient.ApisV1beta1().GatewayConfigurations(namespace.Name).Create(GetCtx(), gatewayConfig, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gatewayConfig)

	t.Log("deploying a GatewayClass resource with the GatewayConfiguration attached via ParametersReference")
	gatewayClass := &gatewayv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: uuid.NewString(),
		},
		Spec: gatewayv1.GatewayClassSpec{
			ParametersRef: &gatewayv1.ParametersReference{
				Group:     gatewayv1.Group(operatorv1beta1.SchemeGroupVersion.Group),
				Kind:      gatewayv1.Kind("GatewayConfiguration"),
				Namespace: (*gatewayv1.Namespace)(&gatewayConfig.Namespace),
				Name:      gatewayConfig.Name,
			},
			ControllerName: gatewayv1.GatewayController(vars.ControllerName()),
		},
	}
	gatewayClass, err = GetClients().GatewayClient.GatewayV1().GatewayClasses().Create(GetCtx(), gatewayClass, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gatewayClass)

	t.Log("deploying Gateway resource")
	gateway := &gwtypes.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace.Name,
			Name:      uuid.NewString(),
		},
		Spec: gatewayv1.GatewaySpec{
			GatewayClassName: gatewayv1.ObjectName(gatewayClass.Name),
			Listeners: []gatewayv1.Listener{{
				Name:     "http",
				Protocol: gatewayv1.HTTPProtocolType,
				Port:     gatewayv1.PortNumber(80),
			}},
		},
	}
	gateway, err = GetClients().GatewayClient.GatewayV1().Gateways(namespace.Name).Create(GetCtx(), gateway, metav1.CreateOptions{})
	require.NoError(t, err)
	cleaner.Add(gateway)

	t.Log("verifying that the DataPlane receives the configuration override")
	require.Eventually(t, func() bool {
		dataplanes, err := gatewayutils.ListDataPlanesForGateway(GetCtx(), GetClients().MgrClient, gateway)
		if err != nil {
			return false
		}
		if len(dataplanes) != 1 {
			return false
		}
		dp := dataplanes[0]
		container := k8sutils.GetPodContainerByName(&dp.Spec.Deployment.PodTemplateSpec.Spec, consts.DataPlaneProxyContainerName)
		if container == nil {
			return false
		}
		for _, envVar := range container.Env {
			if envVar.Name == testEnvVar && envVar.Value == testEnvVal {
				return true
			}
		}
		return false
	}, testutils.GatewayReadyTimeLimit, time.Second)

	t.Log("verifying that the ControlPlane receives the configuration override")
	require.Eventually(t, func() bool {
		controlplanes, err := gatewayutils.ListControlPlanesForGateway(GetCtx(), GetClients().MgrClient, gateway)
		if err != nil {
			return false
		}
		if len(controlplanes) != 1 {
			return false
		}
		cp := controlplanes[0]
		container := k8sutils.GetPodContainerByName(&cp.Spec.Deployment.PodTemplateSpec.Spec, consts.ControlPlaneControllerContainerName)
		if container == nil {
			return false
		}
		for _, envVar := range container.Env {
			if envVar.Name == testEnvVar && envVar.Value == testEnvVal {
				return true
			}
		}
		return false
	}, testutils.ControlPlaneSchedulingTimeLimit, time.Second)

	t.Log("verifying that the DataPlane receives the configuration override")
	require.Eventually(t, func() bool {
		dataplanes, err := gatewayutils.ListDataPlanesForGateway(GetCtx(), GetClients().MgrClient, gateway)
		if err != nil {
			return false
		}
		if len(dataplanes) != 1 {
			return false
		}
		dp := dataplanes[0]
		container := k8sutils.GetPodContainerByName(&dp.Spec.Deployment.PodTemplateSpec.Spec, consts.DataPlaneProxyContainerName)
		if container == nil {
			return false
		}

		for _, envVar := range container.Env {
			if envVar.Name == testEnvVarFromName && envVar.ValueFrom.ConfigMapKeyRef.Key == testEnvVarFromKV {
				return true
			}
		}
		return false
	}, testutils.GatewayReadyTimeLimit, time.Second)

	t.Log("verifying that the ControlPlane receives the configuration override")
	require.Eventually(t, func() bool {
		controlplanes, err := gatewayutils.ListControlPlanesForGateway(GetCtx(), GetClients().MgrClient, gateway)
		if err != nil {
			return false
		}
		if len(controlplanes) != 1 {
			return false
		}
		cp := controlplanes[0]
		container := k8sutils.GetPodContainerByName(&cp.Spec.Deployment.PodTemplateSpec.Spec, consts.ControlPlaneControllerContainerName)
		if container == nil {
			return false
		}
		for _, envVar := range container.Env {
			if envVar.Name == testEnvVarFromName && envVar.ValueFrom.ConfigMapKeyRef.Key == testEnvVarFromKV {
				return true
			}
		}
		return false
	}, testutils.ControlPlaneSchedulingTimeLimit, time.Second)

	t.Log("removing the GatewayConfiguration attachment")
	require.Eventually(t, func() bool {
		gatewayClass, err = GetClients().GatewayClient.GatewayV1().GatewayClasses().Get(GetCtx(), gatewayClass.Name, metav1.GetOptions{})
		if err != nil {
			return false
		}

		gatewayClass.Spec.ParametersRef = nil
		gatewayClass, err = GetClients().GatewayClient.GatewayV1().GatewayClasses().Update(GetCtx(), gatewayClass, metav1.UpdateOptions{})
		return err == nil
	}, testutils.GatewaySchedulingTimeLimit, time.Second)

	t.Log("verifying that the DataPlane loses the configuration override")
	require.Eventually(t, func() bool {
		dataplanes, err := gatewayutils.ListDataPlanesForGateway(GetCtx(), GetClients().MgrClient, gateway)
		if err != nil {
			return false
		}
		if len(dataplanes) != 1 {
			return false
		}
		dp := dataplanes[0]
		container := k8sutils.GetPodContainerByName(&dp.Spec.Deployment.PodTemplateSpec.Spec, consts.DataPlaneProxyContainerName)
		if container == nil {
			return false
		}

		for _, envVar := range container.Env {
			if envVar.Name == testEnvVarFromName && envVar.ValueFrom.ConfigMapKeyRef.Key == testEnvVarFromKV {
				return false
			}
			if envVar.Name == testEnvVar && envVar.Value == testEnvVal {
				return false
			}
		}
		return true
	}, testutils.GatewayReadyTimeLimit, time.Second)

	t.Log("verifying that the ControlPlane receives the configuration override")
	require.Eventually(t, func() bool {
		controlplanes, err := gatewayutils.ListControlPlanesForGateway(GetCtx(), GetClients().MgrClient, gateway)
		if err != nil {
			return false
		}
		if len(controlplanes) != 1 {
			return false
		}
		cp := controlplanes[0]
		container := k8sutils.GetPodContainerByName(&cp.Spec.Deployment.PodTemplateSpec.Spec, consts.ControlPlaneControllerContainerName)
		if container == nil {
			return false
		}
		for _, envVar := range container.Env {
			if envVar.Name == testEnvVarFromName && envVar.ValueFrom.ConfigMapKeyRef.Key == testEnvVarFromKV {
				return false
			}
			if envVar.Name == testEnvVar && envVar.Value == testEnvVal {
				return false
			}
		}
		return true
	}, testutils.ControlPlaneSchedulingTimeLimit, time.Second)
}
