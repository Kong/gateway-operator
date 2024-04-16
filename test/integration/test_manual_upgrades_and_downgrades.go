package integration

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func TestManualGatewayUpgradesAndDowngrades(t *testing.T) {
	t.Parallel()
	namespace, cleaner := helpers.SetupTestEnv(t, GetCtx(), GetEnv())

	originalControlPlaneImageName := "kong/kubernetes-ingress-controller"
	originalControlPlaneImageVersion := "3.1.2"
	originalControlPlaneImage := fmt.Sprintf("%s:%s", originalControlPlaneImageName, originalControlPlaneImageVersion)

	originalDataPlaneImageName := helpers.GetDefaultDataPlaneBaseImage()
	originalDataPlaneImageVersion := "2.7.0"
	originalDataPlaneImage := fmt.Sprintf("%s:%s", originalDataPlaneImageName, originalDataPlaneImageVersion)

	newDataPlaneImageVersion := "2.8.0"
	newDataPlaneImage := fmt.Sprintf("%s:%s", originalDataPlaneImageName, newDataPlaneImageVersion)

	t.Log("deploying a GatewayConfiguration resource")
	gatewayConfig := &operatorv1beta1.GatewayConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace.Name,
			Name:      uuid.NewString(),
		},
		Spec: operatorv1beta1.GatewayConfigurationSpec{
			ControlPlaneOptions: &operatorv1beta1.ControlPlaneOptions{
				Deployment: operatorv1beta1.ControlPlaneDeploymentOptions{
					PodTemplateSpec: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  consts.ControlPlaneControllerContainerName,
									Image: fmt.Sprintf("%s:%s", originalControlPlaneImageName, originalControlPlaneImageVersion),
									ReadinessProbe: &corev1.Probe{
										InitialDelaySeconds: 1,
										PeriodSeconds:       1,
										SuccessThreshold:    1,
									},
								},
							},
						},
					},
				},
			},
			DataPlaneOptions: &operatorv1beta1.GatewayConfigDataPlaneOptions{
				Deployment: operatorv1beta1.DataPlaneDeploymentOptions{
					DeploymentOptions: operatorv1beta1.DeploymentOptions{
						PodTemplateSpec: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  consts.DataPlaneProxyContainerName,
										Image: fmt.Sprintf("%s:%s", originalDataPlaneImageName, originalDataPlaneImageVersion),
										ReadinessProbe: &corev1.Probe{
											InitialDelaySeconds: 1,
											PeriodSeconds:       1,
											SuccessThreshold:    1,
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
	var err error
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

	t.Log("verifying that the ControlPlane receives the configuration override")
	require.Eventually(t, func() bool {
		controlplanes, err := gatewayutils.ListControlPlanesForGateway(GetCtx(), GetClients().MgrClient, gateway)
		if err != nil {
			return false
		}
		if len(controlplanes) != 1 {
			return false
		}

		container := k8sutils.GetPodContainerByName(&controlplanes[0].Spec.Deployment.PodTemplateSpec.Spec, consts.ControlPlaneControllerContainerName)
		if container == nil {
			return false
		}
		return container.Image == fmt.Sprintf("%s:%s", originalControlPlaneImageName, originalControlPlaneImageVersion)
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
		container := k8sutils.GetPodContainerByName(&dataplanes[0].Spec.Deployment.PodTemplateSpec.Spec, consts.DataPlaneProxyContainerName)
		if container == nil {
			return false
		}
		return container.Image == fmt.Sprintf("%s:%s", originalDataPlaneImageName, originalDataPlaneImageVersion)
	}, testutils.GatewayReadyTimeLimit, time.Second)

	t.Log("verifying initial pod images for Gateway")
	require.Eventually(t, func() bool {
		upToDate, err := verifyContainerImageForGateway(gateway, originalControlPlaneImage, originalDataPlaneImage)
		return err == nil && upToDate
	}, time.Minute, time.Second)

	t.Run("upgrade the ControlPlane", func(t *testing.T) {
		t.Skip("skipping upgrade test because we do not have a newer image than 3.1.2 to upgrade to, unskip when 3.1.3 or 3.2.0 is out")

		newControlPlaneImageVersion := "3.2.0"
		newControlPlaneImage := fmt.Sprintf("%s:%s", originalControlPlaneImageName, newControlPlaneImageVersion)

		t.Log("upgrading the ControlPlane version for the Gateway")
		require.Eventually(t, func() bool {
			return changeControlPlaneImage(gatewayConfig, originalControlPlaneImageName, newControlPlaneImageVersion) == nil
		}, time.Second*10, time.Second)

		t.Log("verifying that the ControlPlane receives the configuration override")
		require.Eventually(t, func() bool {
			controlplanes, err := gatewayutils.ListControlPlanesForGateway(GetCtx(), GetClients().MgrClient, gateway)
			if err != nil {
				return false
			}
			if len(controlplanes) != 1 {
				return false
			}
			container := k8sutils.GetPodContainerByName(&controlplanes[0].Spec.Deployment.PodTemplateSpec.Spec, consts.ControlPlaneControllerContainerName)
			if container == nil {
				return false
			}
			return container.Image == fmt.Sprintf("%s:%s", originalControlPlaneImageName, newControlPlaneImageVersion)
		}, testutils.ControlPlaneSchedulingTimeLimit, time.Second)

		t.Log("verifying upgraded ControlPlane Pod images for Gateway")
		require.Eventually(t, func() bool {
			upToDate, err := verifyContainerImageForGateway(gateway, newControlPlaneImage, originalDataPlaneImage)
			return err == nil && upToDate
		}, time.Minute, time.Second)
	})

	t.Log("upgrading the DataPlane version for the Gateway")
	require.Eventually(t, func() bool {
		return changeDataPlaneImage(gatewayConfig, originalDataPlaneImageName, newDataPlaneImageVersion) == nil
	}, time.Second*10, time.Second)

	t.Log("verifying that the DataPlane receives the configuration override")
	require.Eventually(t, func() bool {
		dataplanes, err := gatewayutils.ListDataPlanesForGateway(GetCtx(), GetClients().MgrClient, gateway)
		if err != nil {
			return false
		}
		if len(dataplanes) != 1 {
			return false
		}
		container := k8sutils.GetPodContainerByName(&dataplanes[0].Spec.Deployment.PodTemplateSpec.Spec, consts.DataPlaneProxyContainerName)
		if container == nil {
			return false
		}
		return container.Image == fmt.Sprintf("%s:%s", originalDataPlaneImageName, newDataPlaneImageVersion)
	}, testutils.GatewayReadyTimeLimit, time.Second)

	t.Log("verifying upgraded DataPlane Pod images for Gateway")
	require.Eventually(t, func() bool {
		upToDate, err := verifyContainerImageForGateway(gateway, originalControlPlaneImage, newDataPlaneImage)
		return err == nil && upToDate
	}, time.Minute, time.Second)

	t.Run("downgrade the ControlPlane", func(t *testing.T) {
		t.Skip("skipping downagrade because we only have 1 supported image now (3.1.2)")

		t.Log("downgrading the ControlPlane version for the Gateway")
		require.Eventually(t, func() bool {
			return changeControlPlaneImage(gatewayConfig, originalControlPlaneImageName, originalControlPlaneImageVersion) == nil
		}, time.Second*10, time.Second)

		t.Log("verifying that the ControlPlane receives the configuration override")
		require.Eventually(t, func() bool {
			controlplanes, err := gatewayutils.ListControlPlanesForGateway(GetCtx(), GetClients().MgrClient, gateway)
			if err != nil {
				return false
			}
			if len(controlplanes) != 1 {
				return false
			}
			container := k8sutils.GetPodContainerByName(&controlplanes[0].Spec.Deployment.PodTemplateSpec.Spec, consts.ControlPlaneControllerContainerName)
			if container == nil {
				return false
			}
			return container.Image == fmt.Sprintf("%s:%s", originalControlPlaneImageName, originalControlPlaneImageVersion)
		}, testutils.ControlPlaneSchedulingTimeLimit, time.Second)

		t.Log("verifying downgraded ControlPlane Pod images for Gateway")
		require.Eventually(t, func() bool {
			upToDate, err := verifyContainerImageForGateway(gateway, originalControlPlaneImage, newDataPlaneImage)
			return err == nil && upToDate
		}, time.Minute, time.Second)
	})

	t.Log("downgrading the DataPlane version for the Gateway")
	require.Eventually(t, func() bool {
		return changeDataPlaneImage(gatewayConfig, originalDataPlaneImageName, originalDataPlaneImageVersion) == nil
	}, time.Second*10, time.Second)

	t.Log("verifying that the DataPlane receives the configuration override")
	require.Eventually(t, func() bool {
		dataplanes, err := gatewayutils.ListDataPlanesForGateway(GetCtx(), GetClients().MgrClient, gateway)
		if err != nil {
			return false
		}
		if len(dataplanes) != 1 {
			return false
		}
		container := k8sutils.GetPodContainerByName(&dataplanes[0].Spec.Deployment.PodTemplateSpec.Spec, consts.DataPlaneProxyContainerName)
		if container == nil {
			return false
		}
		return container.Image == fmt.Sprintf("%s:%s", originalDataPlaneImageName, originalDataPlaneImageVersion)
	}, testutils.GatewayReadyTimeLimit, time.Second)

	t.Log("verifying downgraded DataPlane Pod images for Gateway")
	require.Eventually(t, func() bool {
		upToDate, err := verifyContainerImageForGateway(gateway, originalControlPlaneImage, originalDataPlaneImage)
		return err == nil && upToDate
	}, time.Minute, time.Second)
}

// verifyContainerImageForGateway indicates whether or not the underlying
// Pods' containers are configured with the images provided.
func verifyContainerImageForGateway(gateway *gwtypes.Gateway, controlPlaneImage, dataPlaneImage string) (bool, error) {
	controlPlanes, err := gatewayutils.ListControlPlanesForGateway(GetCtx(), GetClients().MgrClient, gateway)
	if err != nil {
		return false, err
	}

	dataPlanes, err := gatewayutils.ListDataPlanesForGateway(GetCtx(), GetClients().MgrClient, gateway)
	if err != nil {
		return false, err
	}

	if len(controlPlanes) != 1 {
		return false, fmt.Errorf("waiting for only 1 ControlPlane")
	}

	if len(dataPlanes) != 1 {
		return false, fmt.Errorf("waiting for only 1 DataPlane")
	}

	deployments, err := k8sutils.ListDeploymentsForOwner(
		GetCtx(),
		GetClients().MgrClient,
		controlPlanes[0].Namespace,
		controlPlanes[0].UID,
		client.MatchingLabels{
			consts.GatewayOperatorManagedByLabel: consts.ControlPlaneManagedLabelValue,
		},
	)
	if err != nil {
		return false, err
	}

	if len(deployments) != 1 {
		return false, fmt.Errorf("waiting for only 1 ControlPlane Deployment")
	}

	for _, deployment := range deployments {
		if len(deployment.Spec.Template.Spec.Containers) < 1 {
			return false, fmt.Errorf("waiting for ControlPlane Deployment to have at least 1 container")
		}
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if container.Image != controlPlaneImage {
				return false, nil
			}
		}
	}

	deployments, err = k8sutils.ListDeploymentsForOwner(
		GetCtx(),
		GetClients().MgrClient,
		dataPlanes[0].Namespace,
		dataPlanes[0].UID,
		client.MatchingLabels{
			consts.GatewayOperatorManagedByLabel: consts.DataPlaneManagedLabelValue,
		},
	)
	if err != nil {
		return false, err
	}

	if len(deployments) != 1 {
		return false, fmt.Errorf("waiting for only 1 DataPlane Deployment")
	}

	for _, deployment := range deployments {
		if len(deployment.Spec.Template.Spec.Containers) < 1 {
			return false, fmt.Errorf("waiting for DataPlane Deployment to have at least 1 container")
		}
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if container.Image != dataPlaneImage {
				return false, nil
			}
		}
	}

	return true, nil
}

// changeControlPlaneImage is a helper function to update the image
// for ControlPlanes in a given GatewayConfiguration.
func changeControlPlaneImage(
	gcfg *operatorv1beta1.GatewayConfiguration,
	controlPlaneImageName,
	controlPlaneImageVersion string,
) error {
	// refresh the object
	gcfg, err := GetClients().OperatorClient.ApisV1beta1().GatewayConfigurations(gcfg.Namespace).Get(GetCtx(), gcfg.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	container := k8sutils.GetPodContainerByName(&gcfg.Spec.ControlPlaneOptions.Deployment.PodTemplateSpec.Spec, consts.ControlPlaneControllerContainerName)
	if container == nil {
		return errors.New("container is nil in GatewayConfiguration ControlPlane options")
	}
	container.Image = fmt.Sprintf("%s:%s", controlPlaneImageName, controlPlaneImageVersion)

	_, err = GetClients().OperatorClient.ApisV1beta1().GatewayConfigurations(gcfg.Namespace).Update(GetCtx(), gcfg, metav1.UpdateOptions{})
	return err
}

// changeDataPlaneImage is a helper function to update the image
// for DataPlane in a given GatewayConfiguration.
func changeDataPlaneImage(
	gcfg *operatorv1beta1.GatewayConfiguration,
	dataPlaneImageName,
	dataPlaneImageVersion string,
) error {
	// refresh the object
	gcfg, err := GetClients().OperatorClient.ApisV1beta1().GatewayConfigurations(gcfg.Namespace).Get(GetCtx(), gcfg.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	container := k8sutils.GetPodContainerByName(&gcfg.Spec.DataPlaneOptions.Deployment.PodTemplateSpec.Spec, consts.DataPlaneProxyContainerName)
	if container == nil {
		return errors.New("container is nil in GatewayConfiguration DataPlane options")
	}
	container.Image = fmt.Sprintf("%s:%s", dataPlaneImageName, dataPlaneImageVersion)

	_, err = GetClients().OperatorClient.ApisV1beta1().GatewayConfigurations(gcfg.Namespace).Update(GetCtx(), gcfg, metav1.UpdateOptions{})
	return err
}
