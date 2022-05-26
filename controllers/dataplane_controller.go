/*
Copyright 2022 Kong Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
	"github.com/kong/gateway-operator/internal/logging"
)

// DataPlaneReconciler reconciles a DataPlane object
type DataPlaneReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager sets up the controller with the Manager.
func (r *DataPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.DataPlane{}).
		Complete(r)
}

//+kubebuilder:rbac:groups=gateway-operator.konghq.com,resources=dataplanes,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=gateway-operator.konghq.com,resources=dataplanes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gateway-operator.konghq.com,resources=dataplanes/finalizers,verbs=update

// Reconcile moves the current state of an object to the intended state.
func (r *DataPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.V(logging.DebugLevel).Info("reconciling data-plane resource", "namespace", req.Namespace, "name", req.Name)
	dataPlane := new(operatorv1alpha1.DataPlane)
	if err := r.Client.Get(ctx, req.NamespacedName, dataPlane); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	log.V(logging.DebugLevel).Info("found data-plane object", "namespace", req.Namespace, "name", req.Name)

	dataPlaneDeployment := new(appsv1.Deployment)
	deploymentExists := false
	if err := r.Client.Get(ctx, req.NamespacedName, dataPlaneDeployment); err == nil {
		deploymentExists = true
	} else {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	dataPlane.Spec.Env = append(dataPlane.Spec.Env, corev1.EnvVar{Name: "KONG_ADMIN_ACCESS_LOG", Value: "/dev/stdout"})
	dataPlane.Spec.Env = append(dataPlane.Spec.Env, corev1.EnvVar{Name: "KONG_ADMIN_ERROR_LOG", Value: "/dev/stderr"})
	dataPlane.Spec.Env = append(dataPlane.Spec.Env, corev1.EnvVar{Name: "KONG_ADMIN_GUI_ACCESS_LOG", Value: "/dev/stdout"})
	dataPlane.Spec.Env = append(dataPlane.Spec.Env, corev1.EnvVar{Name: "KONG_ADMIN_GUI_ERROR_LOG", Value: "/dev/stderr"})
	dataPlane.Spec.Env = append(dataPlane.Spec.Env, corev1.EnvVar{Name: "KONG_ADMIN_LISTEN", Value: "0.0.0.0:8444 ssl"})
	dataPlane.Spec.Env = append(dataPlane.Spec.Env, corev1.EnvVar{Name: "KONG_CLUSTER_LISTEN", Value: "off"})
	dataPlane.Spec.Env = append(dataPlane.Spec.Env, corev1.EnvVar{Name: "KONG_DATABASE", Value: "off"})
	dataPlane.Spec.Env = append(dataPlane.Spec.Env, corev1.EnvVar{Name: "KONG_LUA_PACKAGE_PATH", Value: "/opt/?.lua;/opt/?/init.lua;;"})
	dataPlane.Spec.Env = append(dataPlane.Spec.Env, corev1.EnvVar{Name: "KONG_NGINX_WORKER_PROCESSES", Value: "2"})
	dataPlane.Spec.Env = append(dataPlane.Spec.Env, corev1.EnvVar{Name: "KONG_PLUGINS", Value: "bundled"})
	dataPlane.Spec.Env = append(dataPlane.Spec.Env, corev1.EnvVar{Name: "KONG_PORTAL_API_ACCESS_LOG", Value: "/dev/stdout"})
	dataPlane.Spec.Env = append(dataPlane.Spec.Env, corev1.EnvVar{Name: "KONG_PORTAL_API_ERROR_LOG", Value: "/dev/stderr"})
	dataPlane.Spec.Env = append(dataPlane.Spec.Env, corev1.EnvVar{Name: "KONG_PORT_MAPS", Value: "80:8000, 443:8443"})
	dataPlane.Spec.Env = append(dataPlane.Spec.Env, corev1.EnvVar{Name: "KONG_PROXY_ACCESS_LOG", Value: "/dev/stdout"})
	dataPlane.Spec.Env = append(dataPlane.Spec.Env, corev1.EnvVar{Name: "KONG_PROXY_ERROR_LOG", Value: "/dev/stderr"})
	dataPlane.Spec.Env = append(dataPlane.Spec.Env, corev1.EnvVar{Name: "KONG_PROXY_LISTEN", Value: "0.0.0.0:8000, 0.0.0.0:8443 http2 ssl"})
	dataPlane.Spec.Env = append(dataPlane.Spec.Env, corev1.EnvVar{Name: "KONG_STATUS_LISTEN", Value: "0.0.0.0:8100"})

	// set owner reference to cascade DataPlane deletions to Deployment
	dataPlaneDeployment.ObjectMeta.OwnerReferences = append(dataPlane.ObjectMeta.OwnerReferences, metav1.OwnerReference{
		APIVersion: fmt.Sprintf("%s/%s", dataPlane.GroupVersionKind().Group, dataPlane.GroupVersionKind().Version),
		Kind:       dataPlane.GroupVersionKind().Kind,
		Name:       dataPlane.Name,
		UID:        dataPlane.UID,
	})

	dataPlaneDeployment.ObjectMeta.Namespace = req.Namespace
	dataPlaneDeployment.ObjectMeta.Name = req.Name
	dataPlaneDeployment.ObjectMeta.Labels = map[string]string{"app": req.Name}
	dataPlaneDeployment.Spec = appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": req.Name,
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": req.Name,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:    "proxy",
					Env:     dataPlane.Spec.Env,
					EnvFrom: dataPlane.Spec.EnvFrom,
					Image:   "kong:2.8",
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
	}

	if deploymentExists {
		log.Info("deployment for data-plane already exists, updating", "namespace", req.Namespace, "name", req.Name)
		if err := r.Client.Update(ctx, dataPlaneDeployment); err != nil {
			if errors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}
	} else {
		log.Info("no deployment exists for data-plane, creating", "namespace", req.Namespace, "name", req.Name)
		if err := r.Client.Create(ctx, dataPlaneDeployment); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil

}
