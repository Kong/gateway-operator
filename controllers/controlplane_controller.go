package controllers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
)

// -----------------------------------------------------------------------------
// ControlPlaneReconciler
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

// TODO: need to promote some of the debug log entries to info, and then turn off debug logs by default
// TODO: need to trigger reconciliation whenever labeled Deployments or Services are changed or deleted
// TODO: revisit places to emit events

// -----------------------------------------------------------------------------
// ControlPlaneReconciler - Reconciliation
// -----------------------------------------------------------------------------

//+kubebuilder:rbac:groups=gateway-operator.konghq.com,resources=controlplanes,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=gateway-operator.konghq.com,resources=controlplanes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gateway-operator.konghq.com,resources=controlplanes/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=create;get;list;watch;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups=core,resources=services,verbs=create;get;list;watch;update;patch
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get

// Reconcile moves the current state of an object to the intended state.
func (r *ControlPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	debug(log, "reconciling controlplane resource", req)
	controlplane := new(operatorv1alpha1.ControlPlane)
	if err := r.Client.Get(ctx, req.NamespacedName, controlplane); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	debug(log, "validating controlplane resource conditions", controlplane)
	changed, err := r.ensureControlPlaneIsMarkedScheduled(ctx, controlplane)
	if err != nil {
		return ctrl.Result{}, err
	}
	if changed {
		debug(log, "controlplane resource now marked as scheduled", controlplane)
		return ctrl.Result{}, nil // no need to requeue, status update will requeue
	}

	debug(log, "validating controlplane configuration", controlplane)
	if len(controlplane.Spec.Env) == 0 && len(controlplane.Spec.EnvFrom) == 0 {
		debug(log, "no ENV config found for controlplane resource, setting defaults", controlplane)
		// TODO setControlPlaneDefaults(controlplane) // FIXME: this probably shouldn't be done on the SPEC, perhaps we can remove this when we support Gateway?
		if err := r.Client.Update(ctx, controlplane); err != nil {
			if errors.IsConflict(err) {
				debug(log, "conflict found when updating controlplane resource, retrying", controlplane)
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil // no need to requeue, the update will trigger.
	}

	debug(log, "looking for existing deployments for controlplane resource", controlplane)
	created, controlplaneDeployment, err := r.ensureDeploymentForControlPlane(ctx, controlplane)
	if err != nil {
		return ctrl.Result{}, err
	}
	if created {
		return ctrl.Result{Requeue: true}, nil // TODO: change once deployment create triggers reconciliation
	}

	// TODO: needs to update the existing deployment to ensure latest config

	debug(log, "checking readiness of controlplane deployments", controlplane)
	if controlplaneDeployment.Status.Replicas == 0 || controlplaneDeployment.Status.AvailableReplicas < controlplaneDeployment.Status.Replicas {
		debug(log, "deployment for controlplane not yet ready, waiting", controlplane)
		return ctrl.Result{Requeue: true}, nil
	}

	debug(log, "exposing controlplane deployment via service", controlplane)
	created, controlplaneService, err := r.ensureServiceForControlPlane(ctx, controlplane)
	if err != nil {
		return ctrl.Result{}, err
	}
	if created {
		return ctrl.Result{Requeue: true}, nil // TODO: change once service create triggers reconciliation
	}

	// TODO: needs to update the existing service too to ensure latest config

	debug(log, "checking readiness of controlplane service", controlplaneService)
	if controlplaneService.Spec.ClusterIP == "" {
		return ctrl.Result{Requeue: true}, nil
	}

	debug(log, "reconciliation complete for controlplane resource", controlplane)
	return ctrl.Result{}, r.ensureControlPlaneIsMarkedProvisioned(ctx, controlplane)
}

// -----------------------------------------------------------------------------
// ControlPlaneReconciler - Status Management
// -----------------------------------------------------------------------------

func (r *ControlPlaneReconciler) ensureControlPlaneIsMarkedScheduled(
	ctx context.Context,
	controlplane *operatorv1alpha1.ControlPlane,
) (bool, error) {
	isScheduled := false
	for _, condition := range controlplane.Status.Conditions {
		if condition.Type == string(ControlPlaneConditionTypeProvisioned) {
			isScheduled = true
		}
	}

	if !isScheduled {
		controlplane.Status.Conditions = append(controlplane.Status.Conditions, metav1.Condition{
			Type:               string(ControlPlaneConditionTypeProvisioned),
			Reason:             ControlPlaneConditionReasonPodsNotReady,
			Status:             metav1.ConditionFalse,
			Message:            "controlplane resource is scheduled for provisioning",
			ObservedGeneration: controlplane.Generation,
			LastTransitionTime: metav1.Now(),
		})
		return true, r.Client.Status().Update(ctx, controlplane)
	}

	return false, nil
}

func (r *ControlPlaneReconciler) ensureControlPlaneIsMarkedProvisioned(
	ctx context.Context,
	controlplane *operatorv1alpha1.ControlPlane,
) error {
	updatedConditions := make([]metav1.Condition, 0)
	for _, condition := range controlplane.Status.Conditions {
		if condition.Type == string(ControlPlaneConditionTypeProvisioned) {
			condition.Status = metav1.ConditionTrue
			condition.Reason = ControlPlaneConditionReasonPodsReady
			condition.Message = "pods for all Deployments and/or Daemonsets are ready"
			condition.ObservedGeneration = controlplane.Generation
			condition.LastTransitionTime = metav1.Now()
		}
		updatedConditions = append(updatedConditions, condition)
	}

	controlplane.Status.Conditions = updatedConditions
	return r.Status().Update(ctx, controlplane)
}

// -----------------------------------------------------------------------------
// ControlPlaneReconciler - Owned Resource Management
// -----------------------------------------------------------------------------

func (r *ControlPlaneReconciler) ensureDeploymentForControlPlane(
	ctx context.Context,
	controlplane *operatorv1alpha1.ControlPlane,
) (bool, *appsv1.Deployment, error) {
	/*
		deployments, err := ListDeploymentsForControlPlane(ctx, r.Client, controlplane)
		if err != nil {
			return false, nil, err
		}

		count := len(deployments)
		if count > 1 { // TODO: add handling for multiple deployments
			return false, nil, fmt.Errorf("found %d deployments for controlplane currently unsupported: expected 1 or less", count)
		}

		if count == 1 {
			return false, &deployments[0], nil
		}

		deployment := generateNewDeploymentForControlPlane(controlplane)
		setControlPlaneAsDeploymentOwner(controlplane, deployment)
		labelObjForDataplane(deployment)
		return true, deployment, r.Client.Create(ctx, deployment)
	*/
	return false, nil, fmt.Errorf("unimplemented")
}

func (r *ControlPlaneReconciler) ensureServiceForControlPlane(
	ctx context.Context,
	controlplane *operatorv1alpha1.ControlPlane,
) (bool, *corev1.Service, error) {
	/*
		services, err := ListServicesForControlPlane(ctx, r.Client, controlplane)
		if err != nil {
			return false, nil, err
		}

		count := len(services)
		if count > 1 { // TODO: add handling for multiple services
			return false, nil, fmt.Errorf("found %d services for controlplane currently unsupported: expected 1 or less", count)
		}

		if count == 1 {
			return false, &services[0], nil
		}

		service := generateNewServiceForDataplane(controlplane)
		labelObjForDataplane(service)

		return true, service, r.Client.Create(ctx, service)
	*/
	return false, nil, fmt.Errorf("unimplemented")
}
