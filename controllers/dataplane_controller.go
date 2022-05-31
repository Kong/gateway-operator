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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
)

// -----------------------------------------------------------------------------
// DataPlaneReconciler
// -----------------------------------------------------------------------------

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

// TODO: need to label Deployments and Services created by this controller
// TODO: need to trigger reconciliation whenever labeled Deployments or Services are changed or deleted
// TODO: revisit places to emit events
// TODO: handle the case where a deployment gets deleted and remove from status

// -----------------------------------------------------------------------------
// DataPlaneReconciler - Reconciliation
// -----------------------------------------------------------------------------

//+kubebuilder:rbac:groups=gateway-operator.konghq.com,resources=dataplanes,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=gateway-operator.konghq.com,resources=dataplanes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gateway-operator.konghq.com,resources=dataplanes/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=create;get;list;watch;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups=core,resources=services,verbs=create;get;list;watch;update;patch
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get

// Reconcile moves the current state of an object to the intended state.
func (r *DataPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	debug(log, "reconciling DataPlane resource", req)
	dataplane := new(operatorv1alpha1.DataPlane)
	if err := r.Client.Get(ctx, req.NamespacedName, dataplane); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	debug(log, "validating DataPlane resource conditions", dataplane)
	changed, err := r.ensureDataPlaneIsMarkedScheduled(ctx, dataplane)
	if err != nil {
		return ctrl.Result{}, err
	}
	if changed {
		debug(log, "DataPlane resource now marked as scheduled", dataplane)
		return ctrl.Result{}, nil // no need to requeue, status update will requeue
	}

	debug(log, "validing DataPlane configuration", dataplane)
	if len(dataplane.Spec.Env) == 0 && len(dataplane.Spec.EnvFrom) == 0 {
		debug(log, "no ENV config found for DataPlane resource, setting defaults", dataplane)
		setDataPlaneDefaults(dataplane) // FIXME: this probably shouldn't be done on the SPEC, perhaps we can remove this when we support Gateway?
		if err := r.Client.Update(ctx, dataplane); err != nil {
			if errors.IsConflict(err) {
				debug(log, "conflict found when updating DataPlane resource, retrying", dataplane)
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil // no need to requeue, the update will trigger.
	}

	debug(log, "looking for existing Deployments for DataPlane resource", dataplane)
	dataplaneDeployment, err := r.getDeploymentForDataPlane(ctx, log, dataplane)
	if err != nil {
		if errors.IsNotFound(err) {
			dataplaneDeployment = nil
		} else {
			return ctrl.Result{}, err
		}
	}

	if dataplaneDeployment == nil {
		debug(log, "no Deployment found for DataPlane, creating", dataplane)
		if err := r.createDeploymentForDataPlane(ctx, dataplane); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	debug(log, "checking readiness of DataPlane Deployments", dataplane)
	if dataplaneDeployment.Status.Replicas == 0 || dataplaneDeployment.Status.AvailableReplicas < dataplaneDeployment.Status.Replicas {
		debug(log, "Deployment for DataPlane not yet ready, waiting", dataplane)
		return ctrl.Result{Requeue: true}, nil
	}

	// TODO: associate a service for the deployment with this dataplane via status

	debug(log, "exposing DataPlane Deployment via Service", dataplane)
	// TODO: needs to update the existing service too, for cases like if the deployment is recreated
	created, dataplaneService, err := r.ensureServiceForDataPlane(ctx, dataplane)
	if err != nil {
		return ctrl.Result{}, err
	}
	if created {
		return ctrl.Result{Requeue: true}, nil // TODO: change once service create triggers reconciliation
	}

	if dataplaneService.Spec.ClusterIP == "" {
		debug(log, "waiting for DataPlane Service to be provisioned a ClusterIP", dataplaneService)
		return ctrl.Result{Requeue: true}, nil
	}

	debug(log, "reconciliation complete for DataPlane resource", dataplane)
	return ctrl.Result{}, r.ensureDataPlaneIsMarkedProvisioned(ctx, dataplane)
}

func (r *DataPlaneReconciler) ensureServiceForDataPlane(
	ctx context.Context,
	dataplane *operatorv1alpha1.DataPlane,
) (bool, *corev1.Service, error) {
	nsn := types.NamespacedName{
		Namespace: dataplane.Namespace,
		Name:      "dataplane-" + dataplane.Name,
	}

	service := &corev1.Service{}
	if err := r.Client.Get(ctx, nsn, service); err == nil { // TODO: use generated name
		return false, service, nil
	} else if !errors.IsNotFound(err) {
		return false, nil, err
	}

	service = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       nsn.Namespace,
			Name:            nsn.Name,
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
	labelObjForDataplane(service)

	return true, service, r.Client.Create(ctx, service)
}

// -----------------------------------------------------------------------------
// DataPlaneReconciler - Reconciliation Helper Methods
// -----------------------------------------------------------------------------

// getDeploymentForDataPlane attempts to retrieve an existing Deployment object
// for the provided DataPlane object if one exists.
func (r *DataPlaneReconciler) getDeploymentForDataPlane(
	ctx context.Context,
	log logr.Logger,
	dataplane *operatorv1alpha1.DataPlane,
) (*appsv1.Deployment, error) {
	debug(log, "listing deployments for dataplane", dataplane)
	deployments, err := ListDeploymentsForDataPlane(ctx, r.Client, dataplane)
	if err != nil {
		return nil, err
	}

	count := len(deployments)
	if count > 1 { // FIXME: temporary until there's handling for multiple Deployments
		return nil, fmt.Errorf("found %d deployments for dataplane, expected 1 or less", count)
	}

	if count == 0 {
		debug(log, "found no deployments for dataplane", dataplane)
		return nil, errors.NewNotFound(schema.GroupResource{Group: "apps", Resource: "deployments"}, "unknown")
	}

	return &deployments[0], nil
}

func (r *DataPlaneReconciler) createDeploymentForDataPlane(ctx context.Context, dataplane *operatorv1alpha1.DataPlane) error {
	deployment := generateNewDeploymentForDataPlane(dataplane)
	setDataPlaneAsDeploymentOwner(dataplane, deployment)
	labelObjForDataplane(deployment)
	return r.Client.Create(ctx, deployment)
}

func (r *DataPlaneReconciler) ensureDataPlaneIsMarkedScheduled(ctx context.Context, dataplane *operatorv1alpha1.DataPlane) (bool, error) {
	isScheduled := false
	for _, condition := range dataplane.Status.Conditions {
		if condition.Type == string(DataPlaneConditionTypeProvisioned) {
			isScheduled = true
		}
	}

	if !isScheduled {
		dataplane.Status.Conditions = append(dataplane.Status.Conditions, metav1.Condition{
			Type:               string(DataPlaneConditionTypeProvisioned),
			Reason:             DataPlaneConditionReasonPodsNotReady,
			Status:             metav1.ConditionFalse,
			Message:            "dataplane resource is scheduled for provisioning",
			ObservedGeneration: dataplane.Generation,
			LastTransitionTime: metav1.Now(),
		})
		return true, r.Client.Status().Update(ctx, dataplane)
	}

	return false, nil
}

func (r *DataPlaneReconciler) ensureDataPlaneIsMarkedProvisioned(ctx context.Context, dataplane *operatorv1alpha1.DataPlane) error {
	updatedConditions := make([]metav1.Condition, 0)
	for _, condition := range dataplane.Status.Conditions {
		if condition.Type == string(DataPlaneConditionTypeProvisioned) {
			condition.Status = metav1.ConditionTrue
			condition.Reason = DataPlaneConditionReasonPodsReady
			condition.Message = "pods for all Deployments and/or Daemonsets are ready"
			condition.ObservedGeneration = dataplane.Generation
			condition.LastTransitionTime = metav1.Now()
		}
		updatedConditions = append(updatedConditions, condition)
	}

	dataplane.Status.Conditions = updatedConditions
	return r.Status().Update(ctx, dataplane)
}
