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
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/blang/semver/v4"
	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
	"github.com/kong/gateway-operator/internal/logging"
	"github.com/kong/gateway-operator/internal/rbac"
)

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

//+kubebuilder:rbac:groups=gateway-operator.konghq.com,resources=controlplanes,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=gateway-operator.konghq.com,resources=controlplanes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gateway-operator.konghq.com,resources=controlplanes/finalizers,verbs=update

// Reconcile moves the current state of an object to the intended state.
func (r *ControlPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.V(logging.DebugLevel).Info("reconciling control-plane resource", "namespace", req.Namespace, "name", req.Name)
	controlPlane := new(operatorv1alpha1.ControlPlane)
	if err := r.Client.Get(ctx, req.NamespacedName, controlPlane); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	log.V(logging.DebugLevel).Info("found control-plane object", "namespace", req.Namespace, "name", req.Name)

	// TODO: switch to using ownership
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

	serviceAccountExists := false
	serviceAccount := new(corev1.ServiceAccount)
	if err := r.Client.Get(ctx, req.NamespacedName, serviceAccount); err == nil {
		serviceAccountExists = true
	} else {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	if !serviceAccountExists {
		serviceAccount = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: req.Namespace,
				Name:      req.Name,
			},
		}
		if err := r.Client.Create(ctx, serviceAccount); err != nil {
			return ctrl.Result{}, err
		}
	}

	strVersion := "2.0.0" // FIXME - determine default version
	if controlPlane.Spec.Version != nil {
		strVersion = *controlPlane.Spec.Version
	}
	version, err := semver.Parse(strVersion)
	if err != nil {
		return ctrl.Result{}, err
	}
	roles, clusterRoles := rbac.GetRBACRolesForControlPlaneVersion(version)

	for _, role := range roles {
		role.SetNamespace(req.Namespace)
		if err := r.Client.Create(ctx, &role); err != nil {
			if !errors.IsAlreadyExists(err) {
				return ctrl.Result{}, err
			}
		}
	}

	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kong-leader-election",
			Namespace: req.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "kong-leader-election",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccount.Name,
				Namespace: serviceAccount.Namespace,
			},
		},
	}
	if err := r.Client.Create(ctx, rb); err != nil {
		if !errors.IsAlreadyExists(err) {
			return ctrl.Result{}, err
		}
	}

	for _, role := range clusterRoles {
		if err := r.Client.Create(ctx, &role); err != nil {
			if !errors.IsAlreadyExists(err) {
				return ctrl.Result{}, err
			}
		}
	}

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kong-ingress",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "kong-ingress",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccount.Name,
				Namespace: serviceAccount.Namespace,
			},
		},
	}
	if err := r.Client.Create(ctx, crb); err != nil {
		if !errors.IsAlreadyExists(err) {
			return ctrl.Result{}, err
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

	controlPlane.Spec.Env = append(controlPlane.Spec.Env, corev1.EnvVar{Name: "CONTROLLER_KONG_ADMIN_URL", Value: fmt.Sprintf("https://dataplane-sample.%s.svc:8444", req.Namespace)}) // FIXME
	controlPlane.Spec.Env = append(controlPlane.Spec.Env, corev1.EnvVar{Name: "CONTROLLER_KONG_ADMIN_TLS_SKIP_VERIFY", Value: "true"})
	controlPlane.Spec.Env = append(controlPlane.Spec.Env, corev1.EnvVar{Name: "CONTROLLER_PUBLISH_SERVICE", Value: "kong/kong-gateway-sample"})
	controlPlane.Spec.Env = append(controlPlane.Spec.Env, corev1.EnvVar{Name: "CONTROLLER_FEATURE_GATES", Value: "Gateway=true"})
	controlPlane.Spec.Env = append(controlPlane.Spec.Env, corev1.EnvVar{Name: "POD_NAME", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "metadata.name"}}})
	controlPlane.Spec.Env = append(controlPlane.Spec.Env, corev1.EnvVar{Name: "POD_NAMESPACE", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "metadata.namespace"}}})

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
				ServiceAccountName: serviceAccount.Name,
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
		log.Info("deployment for control-plane already exists, updating", "namespace", req.Namespace, "name", req.Name)
		if err := r.Client.Update(ctx, controlPlaneDeployment); err != nil {
			if errors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}
	} else {
		log.Info("no deployment exists for control-plane, creating", "namespace", req.Namespace, "name", req.Name)
		if err := r.Client.Create(ctx, controlPlaneDeployment); err != nil {
			return ctrl.Result{}, err
		}
	}

	serviceExists := false
	controlPlaneService := new(corev1.Service)
	if err := r.Client.Get(ctx, req.NamespacedName, controlPlaneService); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	} else {
		serviceExists = true
		log.Info("service found, no need to create")
	}

	if !serviceExists {
		log.Info("service not found, creating")
		svcPorts := []corev1.ServicePort{}
		for _, p := range controlPlaneDeployment.Spec.Template.Spec.Containers[0].Ports {
			svcPorts = append(svcPorts, corev1.ServicePort{
				Name:     p.Name,
				Protocol: p.Protocol,
				Port:     p.ContainerPort,
			})
		}

		controlPlaneService.ObjectMeta.Namespace = req.Namespace
		controlPlaneService.ObjectMeta.Name = req.Name
		controlPlaneService.Spec.Type = corev1.ServiceTypeLoadBalancer
		controlPlaneService.Spec.Selector = controlPlaneDeployment.Spec.Selector.MatchLabels
		controlPlaneService.Spec.Ports = svcPorts

		if err := r.Client.Create(ctx, controlPlaneService); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
