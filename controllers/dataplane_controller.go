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

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

// TODO: need to label Deployments created by this controller
// TODO: need to trigger reconciliation whenever labeled Deployments are changed or deleted
// TODO: revisit places to emit events

// -----------------------------------------------------------------------------
// DataPlaneReconciler - Reconciliation
// -----------------------------------------------------------------------------

//+kubebuilder:rbac:groups=gateway-operator.konghq.com,resources=dataplanes,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=gateway-operator.konghq.com,resources=dataplanes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gateway-operator.konghq.com,resources=dataplanes/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=create;get;list;watch;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get

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
	dataplaneDeployment, err := r.getDeploymentForDataPlane(ctx, dataplane)
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
	if dataplaneDeployment.Status.AvailableReplicas < dataplaneDeployment.Status.Replicas {
		debug(log, "Deployment for DataPlane not yet ready, waiting", dataplane)
		return ctrl.Result{Requeue: true}, nil
	}

	debug(log, "exposing DataPlane Deployment via Service", dataplane)
	// TODO: create a service to expose the DataPlane also (and add an owner reference to this dataplane).
	//       wait for a cluster IP.

	debug(log, "reconciliation complete for DataPlane resource", dataplane)
	return ctrl.Result{}, r.ensureDataPlaneIsMarkedProvisioned(ctx, dataplane)
}

// -----------------------------------------------------------------------------
// DataPlaneReconciler - Reconciliation Helper Methods
// -----------------------------------------------------------------------------

// getDeploymentForDataPlane attempts to retrieve an existing Deployment object
// for the provided DataPlane object if one exists.
func (r *DataPlaneReconciler) getDeploymentForDataPlane(ctx context.Context, dataPlane *operatorv1alpha1.DataPlane) (*appsv1.Deployment, error) {
	dataPlaneDeployment := new(appsv1.Deployment)
	if err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: dataPlane.Namespace,
		Name:      dataPlane.Name,
	}, dataPlaneDeployment); err != nil {
		return nil, err
	}
	return dataPlaneDeployment, nil
}

func (r *DataPlaneReconciler) createDeploymentForDataPlane(ctx context.Context, dataplane *operatorv1alpha1.DataPlane) error {
	deployment := generateNewDeploymentForDataPlane(dataplane)
	setDataPlaneAsDeploymentOwner(dataplane, deployment)
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
