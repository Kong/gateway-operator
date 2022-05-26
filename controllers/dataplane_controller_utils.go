package controllers

import (
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kong/gateway-operator/api/v1alpha1"
	"github.com/kong/gateway-operator/internal/consts"
	"github.com/kong/gateway-operator/internal/logging"
)

func setDataPlaneDefaults(dataplane *v1alpha1.DataPlane) {
	// FIXME: these defaults are kind of esoteric, is there a better way to express and document them? do we actually need all of them?
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_ADMIN_ACCESS_LOG", Value: "/dev/stdout"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_ADMIN_ERROR_LOG", Value: "/dev/stderr"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_ADMIN_GUI_ACCESS_LOG", Value: "/dev/stdout"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_ADMIN_GUI_ERROR_LOG", Value: "/dev/stderr"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_ADMIN_LISTEN", Value: "0.0.0.0:8444 ssl"})
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
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_PROXY_LISTEN", Value: "0.0.0.0:8000, 0.0.0.0:8443 http2 ssl"})
	dataplane.Spec.Env = append(dataplane.Spec.Env, corev1.EnvVar{Name: "KONG_STATUS_LISTEN", Value: "0.0.0.0:8100"})
}

func setDataPlaneAsDeploymentOwner(dataplane *v1alpha1.DataPlane, deployment *appsv1.Deployment) {
	foundOwnerRef := false
	for _, ownerRef := range deployment.ObjectMeta.OwnerReferences {
		if ownerRef.UID == dataplane.UID {
			foundOwnerRef = true
		}
	}
	if !foundOwnerRef {
		deployment.ObjectMeta.OwnerReferences = append(deployment.ObjectMeta.OwnerReferences, metav1.OwnerReference{
			APIVersion: fmt.Sprintf("%s/%s", dataplane.GroupVersionKind().Group, dataplane.GroupVersionKind().Version),
			Kind:       dataplane.GroupVersionKind().Kind,
			Name:       dataplane.Name,
			UID:        dataplane.UID,
		})
	}
}

func generateNewDeploymentForDataPlane(dataplane *v1alpha1.DataPlane) *appsv1.Deployment {
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

func debug(log logr.Logger, msg string, rawOBJ interface{}) {
	if obj, ok := rawOBJ.(client.Object); ok {
		log.V(logging.DebugLevel).Info(msg, "namespace", obj.GetNamespace(), "name", obj.GetName())
	} else if req, ok := rawOBJ.(reconcile.Request); ok {
		log.V(logging.DebugLevel).Info(msg, "namespace", req.Namespace, "name", req.Name)
	} else {
		log.V(logging.DebugLevel).Info(fmt.Sprintf("unexpected type processed for debug logging: %T, this is a bug!", rawOBJ))
	}
}
