package controllers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
	"github.com/kong/gateway-operator/internal/consts"
)

// -----------------------------------------------------------------------------
// Public Consts - Labels
// -----------------------------------------------------------------------------

const (
	// GatewayOperatorControlledLabel is the label that is used for objects which
	// were created by this operator.
	GatewayOperatorControlledLabel = "konghq.com/gateway-operator"

	// DataPlaneManagedLabel indicates that an object's lifecycle is managed by
	// the dataplane controller.
	DataPlaneManagedLabel = "dataplane"
)

// -----------------------------------------------------------------------------
// Public Functions - Owner References
// -----------------------------------------------------------------------------

// ListDeploymentsForDataPlane uses labels to list Deployments which were
// created by the dataplane controller, and further filters on those which
// were created specifically for the provided dataplane object.
func ListDeploymentsForDataPlane(
	ctx context.Context,
	c client.Client,
	dataplane *operatorv1alpha1.DataPlane,
) ([]appsv1.Deployment, error) {
	requirement, err := labels.NewRequirement(
		GatewayOperatorControlledLabel,
		selection.Equals,
		[]string{DataPlaneManagedLabel},
	)
	if err != nil {
		return nil, err
	}
	selector := labels.NewSelector().Add(*requirement)

	deploymentList := &appsv1.DeploymentList{}
	if err := c.List(ctx, deploymentList, &client.ListOptions{
		Namespace:     dataplane.Namespace,
		LabelSelector: selector,
	}); err != nil {
		return nil, err
	}

	deployments := make([]appsv1.Deployment, 0)
	for _, deployment := range deploymentList.Items {
		for _, ownerRef := range deployment.ObjectMeta.OwnerReferences {
			if ownerRef.UID == dataplane.UID {
				deployments = append(deployments, deployment)
				break
			}
		}
	}

	return deployments, nil
}

// ListServicesForDataPlane uses labels to list Services which were
// created by the dataplane controller, and further filters on those which
// were created specifically for the provided dataplane object.
func ListServicesForDataPlane(
	ctx context.Context,
	c client.Client,
	dataplane *operatorv1alpha1.DataPlane,
) ([]corev1.Service, error) {
	requirement, err := labels.NewRequirement(
		GatewayOperatorControlledLabel,
		selection.Equals,
		[]string{DataPlaneManagedLabel},
	)
	if err != nil {
		return nil, err
	}
	selector := labels.NewSelector().Add(*requirement)

	serviceList := &corev1.ServiceList{}
	if err := c.List(ctx, serviceList, &client.ListOptions{
		Namespace:     dataplane.Namespace,
		LabelSelector: selector,
	}); err != nil {
		return nil, err
	}

	services := make([]corev1.Service, 0)
	for _, service := range serviceList.Items {
		for _, ownerRef := range service.ObjectMeta.OwnerReferences {
			if ownerRef.UID == dataplane.UID {
				services = append(services, service)
				break
			}
		}
	}

	return services, nil
}

// -----------------------------------------------------------------------------
// Private Functions - Generators
// -----------------------------------------------------------------------------

const (
	defaultHTTPPort  = 80
	defaultHTTPSPort = 443

	defaultKongHTTPPort   = 8000
	defaultKongHTTPSPort  = 8443
	defaultKongAdminPort  = 8444
	defaultKongStatusPort = 8100
)

func setDataPlaneDefaults(dataplane *operatorv1alpha1.DataPlane) {
	// FIXME: these defaults are kind of esoteric, is there a better way to express and document them? do we actually need all of them?
	// TODO: make this a generator that returns the object instead
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_ADMIN_ACCESS_LOG", Value: "/dev/stdout"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_ADMIN_ERROR_LOG", Value: "/dev/stderr"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_ADMIN_GUI_ACCESS_LOG", Value: "/dev/stdout"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_ADMIN_GUI_ERROR_LOG", Value: "/dev/stderr"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_ADMIN_LISTEN", Value: fmt.Sprintf("0.0.0.0:%d ssl", defaultKongAdminPort)})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_CLUSTER_LISTEN", Value: "off"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_DATABASE", Value: "off"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_LUA_PACKAGE_PATH", Value: "/opt/?.lua;/opt/?/init.lua;;"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_NGINX_WORKER_PROCESSES", Value: "2"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_PLUGINS", Value: "bundled"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_PORTAL_API_ACCESS_LOG", Value: "/dev/stdout"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_PORTAL_API_ERROR_LOG", Value: "/dev/stderr"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_PORT_MAPS", Value: "80:8000, 443:8443"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_PROXY_ACCESS_LOG", Value: "/dev/stdout"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_PROXY_ERROR_LOG", Value: "/dev/stderr"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_PROXY_LISTEN", Value: fmt.Sprintf("0.0.0.0:%d, 0.0.0.0:%d http2 ssl", defaultKongHTTPPort, defaultKongHTTPSPort)})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_STATUS_LISTEN", Value: fmt.Sprintf("0.0.0.0:%d", defaultKongStatusPort)})
}

func generateNewDeploymentForDataPlane(dataplane *operatorv1alpha1.DataPlane) *appsv1.Deployment {
	var dataplaneImage string
	if dataplane.Spec.ContainerImage != nil {
		dataplaneImage = *dataplane.Spec.ContainerImage
		if dataplane.Spec.Version != nil {
			dataplaneImage = fmt.Sprintf("%s:%s", dataplaneImage, *dataplane.Spec.Version)
		}
	} else {
		dataplaneImage = consts.DefaultDataPlaneImage // FIXME: find default dynamically if possible
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: dataplane.Namespace,
			Name:      dataplane.Name, // FIXME need a unique generated name
			Labels: map[string]string{
				"app": dataplane.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": dataplane.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": dataplane.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:    "proxy",
						Env:     dataplane.Spec.Env,
						EnvFrom: dataplane.Spec.EnvFrom,
						Image:   dataplaneImage,
						Lifecycle: &corev1.Lifecycle{
							PreStop: &corev1.LifecycleHandler{
								Exec: &corev1.ExecAction{
									Command: []string{
										"/bin/sh",
										"-c",
										"kong quit",
									},
								},
							},
						},
						Ports: []corev1.ContainerPort{
							{
								Name:          "proxy",
								ContainerPort: 8000,
								Protocol:      corev1.ProtocolTCP,
							},
							{
								Name:          "proxy-ssl",
								ContainerPort: 8443,
								Protocol:      corev1.ProtocolTCP,
							},
							{
								Name:          "metrics",
								ContainerPort: 8100,
								Protocol:      corev1.ProtocolTCP,
							},
							{
								Name:          "admin-ssl",
								ContainerPort: 8444,
								Protocol:      corev1.ProtocolTCP,
							},
						},
						ReadinessProbe: &corev1.Probe{
							FailureThreshold:    3,
							InitialDelaySeconds: 5,
							PeriodSeconds:       10,
							SuccessThreshold:    1,
							TimeoutSeconds:      1,
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path:   "/status",
									Port:   intstr.FromInt(8100),
									Scheme: corev1.URISchemeHTTP,
								},
							},
						},
					}},
				},
			},
		},
	}
	return deployment
}

func generateNewServiceForDataplane(dataplane *operatorv1alpha1.DataPlane) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       dataplane.Namespace,
			Name:            "dataplane-" + dataplane.Name, // TODO: generate instead
			OwnerReferences: []metav1.OwnerReference{createOwnerRefForDataPlane(dataplane)},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeLoadBalancer, // TODO: dynamically figure this out
			Selector: map[string]string{"app": dataplane.Name},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       defaultHTTPPort, // TODO: add dynamic port determinations
					TargetPort: intstr.FromInt(defaultKongHTTPPort),
				},
				{
					Name:       "https",
					Protocol:   corev1.ProtocolTCP,
					Port:       defaultHTTPSPort, // TODO: add dynamic port determinations
					TargetPort: intstr.FromInt(defaultKongHTTPSPort),
				},
				{ // TODO: in time, create a separate ClusterIP ONLY admin Service (this is just convenient for the moment, but not secure)
					Name:     "admin",
					Protocol: corev1.ProtocolTCP,
					Port:     defaultKongAdminPort, // TODO: add dynamic port determinations
				},
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Private Functions - Owner References
// -----------------------------------------------------------------------------

func createOwnerRefForDataPlane(dataplane *operatorv1alpha1.DataPlane) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: apiVersionForDataplane(dataplane),
		Kind:       dataplane.GroupVersionKind().Kind,
		Name:       dataplane.Name,
		UID:        dataplane.UID,
	}
}

func apiVersionForDataplane(dataplane *operatorv1alpha1.DataPlane) string {
	return fmt.Sprintf("%s/%s", dataplane.GroupVersionKind().Group, dataplane.GroupVersionKind().Version)
}

func setDataPlaneAsDeploymentOwner(dataplane *operatorv1alpha1.DataPlane, deployment *appsv1.Deployment) {
	foundOwnerRef := false
	for _, ownerRef := range deployment.ObjectMeta.OwnerReferences {
		if ownerRef.UID == dataplane.UID {
			foundOwnerRef = true
		}
	}
	if !foundOwnerRef {
		deployment.ObjectMeta.OwnerReferences = append(deployment.ObjectMeta.OwnerReferences, createOwnerRefForDataPlane(dataplane))
	}
}

func labelObjForDataplane(obj client.Object) {
	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[GatewayOperatorControlledLabel] = DataPlaneManagedLabel
	obj.SetLabels(labels)
}
