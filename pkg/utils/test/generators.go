package test

import (
	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	operatorv1beta1 "github.com/kong/gateway-operator/api/v1beta1"
	gwtypes "github.com/kong/gateway-operator/internal/types"
	"github.com/kong/gateway-operator/pkg/consts"
	"github.com/kong/gateway-operator/pkg/vars"
	"github.com/kong/gateway-operator/test/helpers"
)

// GenerateGatewayClass generates the default GatewayClass to be used in tests
func GenerateGatewayClass() *gatewayv1.GatewayClass {
	gatewayClass := &gatewayv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: uuid.NewString(),
		},
		Spec: gatewayv1.GatewayClassSpec{
			ControllerName: gatewayv1.GatewayController(vars.ControllerName()),
		},
	}
	return gatewayClass
}

// GenerateGateway generates a Gateway to be used in tests
func GenerateGateway(gatewayNSN types.NamespacedName, gatewayClass *gatewayv1.GatewayClass, opts ...func(gateway *gatewayv1.Gateway)) *gwtypes.Gateway {
	gateway := &gwtypes.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: gatewayNSN.Namespace,
			Name:      gatewayNSN.Name,
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

	for _, opt := range opts {
		opt(gateway)
	}

	return gateway
}

// GenerateGatewayConfiguration generates a GatewayConfiguration to be used in tests
func GenerateGatewayConfiguration(gatewayConfigurationNSN types.NamespacedName) *operatorv1beta1.GatewayConfiguration {
	return &operatorv1beta1.GatewayConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: gatewayConfigurationNSN.Namespace,
			Name:      gatewayConfigurationNSN.Name,
		},
		Spec: operatorv1beta1.GatewayConfigurationSpec{
			ControlPlaneOptions: &operatorv1beta1.ControlPlaneOptions{
				Deployment: operatorv1beta1.ControlPlaneDeploymentOptions{
					PodTemplateSpec: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  consts.ControlPlaneControllerContainerName,
									Image: consts.DefaultControlPlaneImage,
									ReadinessProbe: &corev1.Probe{
										FailureThreshold:    3,
										InitialDelaySeconds: 0,
										PeriodSeconds:       1,
										SuccessThreshold:    1,
										TimeoutSeconds:      1,
										ProbeHandler: corev1.ProbeHandler{
											HTTPGet: &corev1.HTTPGetAction{
												Path:   "/readyz",
												Port:   intstr.FromInt(10254),
												Scheme: corev1.URISchemeHTTP,
											},
										},
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
										Image: helpers.GetDefaultDataPlaneImage(),
										ReadinessProbe: &corev1.Probe{
											FailureThreshold:    3,
											InitialDelaySeconds: 0,
											PeriodSeconds:       1,
											SuccessThreshold:    1,
											TimeoutSeconds:      1,
											ProbeHandler: corev1.ProbeHandler{
												HTTPGet: &corev1.HTTPGetAction{
													Path:   "/status",
													Port:   intstr.FromInt(consts.DataPlaneMetricsPort),
													Scheme: corev1.URISchemeHTTP,
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
}
