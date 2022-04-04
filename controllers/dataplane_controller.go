/*
Copyright 2022.

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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	operatorv1alpha1 "github.com/kong/operator/api/v1alpha1"
)

// -----------------------------------------------------------------------------
// DataPlaneReconciler - Public
// -----------------------------------------------------------------------------

// DataPlaneReconciler reconciles a DataPlane object
type DataPlaneReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager sets up the controller with the Manager.
func (r *DataPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1alpha2.Gateway{}).
		Owns(&operatorv1alpha1.GatewayConfiguration{}).
		Complete(r)
}

// -----------------------------------------------------------------------------
// DataPlaneReconciler - Reconcilation
// -----------------------------------------------------------------------------

//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=gateways,verbs=get;list;watch;update
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=gateways/status,verbs=get;update
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=gatewayclasses,verbs=get;list;watch
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=gatewayclasses/status,verbs=get;update

func (r *DataPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	gw := new(gatewayv1alpha2.Gateway)
	if err := r.Client.Get(ctx, req.NamespacedName, gw); err != nil {
		if errors.IsNotFound(err) {
			log.Info("gateway object queued but no longer available, skipping", "namespace", req.Namespace, "name", req.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	finalizers := gw.GetFinalizers()
	foundControlPlaneFinalizer := false
	for _, finalizer := range finalizers {
		if finalizer == "wait-for-deployments" {
			foundControlPlaneFinalizer = true
		}
	}
	if !foundControlPlaneFinalizer {
		gw.Finalizers = append(gw.Finalizers, "wait-for-deployments")
		if err := r.Client.Update(ctx, gw); err != nil {
			if errors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}
	}

	gwcfg := new(operatorv1alpha1.GatewayConfiguration)
	if err := r.Client.Get(ctx, req.NamespacedName, gwcfg); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		log.Info("gateway config not found, using vanilla config", "namespace", req.Namespace, "name", req.Name)
	}

	gatewayDeployment := new(appsv1.Deployment)
	gatewayService := new(corev1.Service)
	if gw.DeletionTimestamp != nil {
		now := metav1.Now()
		if gw.DeletionTimestamp.Before(&now) {
			if err := r.Client.Get(ctx, req.NamespacedName, gatewayDeployment); err != nil {
				if errors.IsNotFound(err) {
					if err := r.Client.Get(ctx, req.NamespacedName, gatewayService); err != nil {
						if errors.IsNotFound(err) {
							var newFinalizers []string
							for _, finalizer := range finalizers {
								if finalizer != "wait-for-deployments" {
									newFinalizers = append(newFinalizers, finalizer)
								}
							}
							gw.SetFinalizers(newFinalizers)
							if err := r.Client.Update(ctx, gw); err != nil {
								if errors.IsConflict(err) {
									return ctrl.Result{Requeue: true}, nil
								}
								return ctrl.Result{}, err
							}
							return ctrl.Result{}, nil
						}
						return ctrl.Result{}, err
					}
					if err := r.Client.Delete(ctx, gatewayService); err != nil {
						return ctrl.Result{}, err
					}
					return ctrl.Result{Requeue: true}, nil
				}
				return ctrl.Result{}, err
			}
			if err := r.Client.Delete(ctx, gatewayDeployment); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
	}

	deploymentExists := false
	if err := r.Client.Get(ctx, req.NamespacedName, gatewayDeployment); err == nil {
		deploymentExists = true
	} else {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	serviceExists := false
	if err := r.Client.Get(ctx, req.NamespacedName, gatewayService); err == nil {
		serviceExists = true
	} else {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	gatewayDeployment.ObjectMeta.Namespace = req.Namespace
	gatewayDeployment.ObjectMeta.Name = req.Name
	gatewayDeployment.ObjectMeta.Labels = map[string]string{"app": req.Name}
	gatewayDeployment.Spec = appsv1.DeploymentSpec{
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
					Env:     gwcfg.Spec.Env,
					EnvFrom: gwcfg.Spec.EnvFrom,
					Image:   "kong:2.7",
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

	svcPorts := []corev1.ServicePort{}
	for _, p := range gatewayDeployment.Spec.Template.Spec.Containers[0].Ports {
		svcPorts = append(svcPorts, corev1.ServicePort{
			Name:     p.Name,
			Protocol: p.Protocol,
			Port:     p.ContainerPort,
		})
	}

	gatewayService.ObjectMeta.Namespace = req.Namespace
	gatewayService.ObjectMeta.Name = req.Name
	gatewayService.Spec.Type = corev1.ServiceTypeLoadBalancer
	gatewayService.Spec.Selector = gatewayDeployment.Spec.Selector.MatchLabels
	gatewayService.Spec.Ports = svcPorts

	if deploymentExists {
		log.Info("deployment for gateway already exists, updating", "namespace", req.Namespace, "name", req.Name)
		if err := r.Client.Update(ctx, gatewayDeployment); err != nil {
			if errors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}
	} else {
		log.Info("no deployment exists for gateway, creating", "namespace", req.Namespace, "name", req.Name)
		if err := r.Client.Create(ctx, gatewayDeployment); err != nil {
			return ctrl.Result{}, err
		}
	}

	if serviceExists {
		log.Info("service for gateway already exists, updating", "namespace", req.Namespace, "name", req.Name)
		if err := r.Client.Update(ctx, gatewayService); err != nil {
			if errors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}
	} else {
		log.Info("no service for gateway exists, creating", "namespace", req.Namespace, "name", req.Name)
		if err := r.Client.Create(ctx, gatewayService); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
