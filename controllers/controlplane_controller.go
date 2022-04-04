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

	operatorv1alpha1 "github.com/kong/operator/api/v1alpha1"
)

// -----------------------------------------------------------------------------
// ControlPlaneReconciler - Public
// -----------------------------------------------------------------------------

// ControlPlaneReconciler reconciles a ControlPlane object
type ControlPlaneReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager sets up the controller with the Manager.
func (r *ControlPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.ControlPlane{}).
		Complete(r)
}

// -----------------------------------------------------------------------------
// ControlPlaneReconciler - Reconcilation
// -----------------------------------------------------------------------------

//+kubebuilder:rbac:groups=operator.konghq.com,resources=controlplanes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.konghq.com,resources=controlplanes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.konghq.com,resources=controlplanes/finalizers,verbs=update

func (r *ControlPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	controlPlane := new(operatorv1alpha1.ControlPlane)
	if err := r.Client.Get(ctx, req.NamespacedName, controlPlane); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	log.Info("found controlplane instance", "namespace", req.Namespace, "name", req.Name)

	finalizers := controlPlane.GetFinalizers()
	foundControlPlaneFinalizer := false
	for _, finalizer := range finalizers {
		if finalizer == "wait-for-deployments" {
			foundControlPlaneFinalizer = true
		}
	}
	if !foundControlPlaneFinalizer {
		controlPlane.Finalizers = append(controlPlane.Finalizers, "wait-for-deployments")
		if err := r.Client.Update(ctx, controlPlane); err != nil {
			if errors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}
	}

	controlPlaneDeployment := new(appsv1.Deployment)
	if controlPlane.DeletionTimestamp != nil {
		now := metav1.Now()
		if controlPlane.DeletionTimestamp.Before(&now) {
			if err := r.Client.Get(ctx, req.NamespacedName, controlPlaneDeployment); err != nil {
				if errors.IsNotFound(err) {
					// TODO: all set, drop finalizer
					var newFinalizers []string
					for _, finalizer := range finalizers {
						if finalizer != "wait-for-deployments" {
							newFinalizers = append(newFinalizers, finalizer)
						}
					}
					controlPlane.SetFinalizers(newFinalizers)
					if err := r.Client.Update(ctx, controlPlane); err != nil {
						if errors.IsConflict(err) {
							return ctrl.Result{Requeue: true}, nil
						}
						return ctrl.Result{}, err
					}
					return ctrl.Result{}, nil
				}
				return ctrl.Result{}, err
			}
			if err := r.Client.Delete(ctx, controlPlaneDeployment); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
	}

	deploymentExists := false
	if err := r.Client.Get(ctx, req.NamespacedName, controlPlaneDeployment); err == nil {
		deploymentExists = true
	} else {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	controlPlaneDeployment.ObjectMeta.Namespace = req.Namespace
	controlPlaneDeployment.ObjectMeta.Name = req.Name
	controlPlaneDeployment.ObjectMeta.Labels = map[string]string{"app": req.Name}
	controlPlaneDeployment.Spec = appsv1.DeploymentSpec{
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
				ServiceAccountName: "kong-serviceaccount",
				Containers: []corev1.Container{{
					Name:    "ingress-controller",
					Env:     controlPlane.Spec.Env,
					EnvFrom: controlPlane.Spec.EnvFrom,
					Image:   "kong/kubernetes-ingress-controller:latest",
					Ports: []corev1.ContainerPort{
						{
							Name:          "webhook",
							ContainerPort: 8080,
							Protocol:      corev1.ProtocolTCP,
						},
						{
							Name:          "cmetrics",
							ContainerPort: 10255,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					LivenessProbe: &corev1.Probe{
						FailureThreshold:    3,
						InitialDelaySeconds: 5,
						PeriodSeconds:       10,
						SuccessThreshold:    1,
						TimeoutSeconds:      1,
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path:   "/healthz",
								Port:   intstr.FromInt(10254),
								Scheme: corev1.URISchemeHTTP,
							},
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
								Path:   "/readyz",
								Port:   intstr.FromInt(10254),
								Scheme: corev1.URISchemeHTTP,
							},
						},
					},
				}},
			},
		},
	}

	if deploymentExists {
		log.Info("deployment for controlplane already exists, updating", "namespace", req.Namespace, "name", req.Name)
		if err := r.Client.Update(ctx, controlPlaneDeployment); err != nil {
			if errors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}
	} else {
		log.Info("no deployment exists for controlplane, creating", "namespace", req.Namespace, "name", req.Name)
		if err := r.Client.Create(ctx, controlPlaneDeployment); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
