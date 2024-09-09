package dataplane

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	appsv1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1beta1 "github.com/kong/gateway-operator/api/v1beta1"
	"github.com/kong/gateway-operator/controller/pkg/ctxinjector"
	"github.com/kong/gateway-operator/controller/pkg/log"
	"github.com/kong/gateway-operator/controller/pkg/op"
	"github.com/kong/gateway-operator/pkg/consts"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
	k8sresources "github.com/kong/gateway-operator/pkg/utils/kubernetes/resources"
)

// -----------------------------------------------------------------------------
// DataPlaneReconciler
// -----------------------------------------------------------------------------

type dataPlaneValidator interface {
	Validate(*operatorv1beta1.DataPlane) error
}

// Reconciler reconciles a DataPlane object
type Reconciler struct {
	client.Client
	Scheme                   *runtime.Scheme
	eventRecorder            record.EventRecorder
	ClusterCASecretName      string
	ClusterCASecretNamespace string
	DevelopmentMode          bool
	Validator                dataPlaneValidator
	Callbacks                DataPlaneCallbacks
	ContextInjector          ctxinjector.CtxInjector
	DefaultImage             string
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	r.eventRecorder = mgr.GetEventRecorderFor("dataplane")

	return DataPlaneWatchBuilder(mgr).
		Complete(r)
}

// -----------------------------------------------------------------------------
// DataPlaneReconciler - Reconciliation
// -----------------------------------------------------------------------------

// Reconcile moves the current state of an object to the intended state.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Calling it here ensures that evaluated values will be used for the duration of this function.
	ctx = r.ContextInjector.InjectKeyValues(ctx)
	logger := log.GetLogger(ctx, "dataplane", r.DevelopmentMode)

	log.Trace(logger, "reconciling DataPlane resource", req)
	dpNn := req.NamespacedName
	dataplane := new(operatorv1beta1.DataPlane)
	if err := r.Client.Get(ctx, dpNn, dataplane); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if k8sutils.InitReady(dataplane) {
		if patched, err := patchDataPlaneStatus(ctx, r.Client, logger, dataplane); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed initializing DataPlane Ready condition: %w", err)
		} else if patched {
			return ctrl.Result{}, nil
		}
	}

	if err := r.initSelectorInStatus(ctx, logger, dataplane); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed updating DataPlane with selector in Status: %w", err)
	}

	log.Trace(logger, "validating DataPlane configuration", dataplane)
	err := r.Validator.Validate(dataplane)
	if err != nil {
		log.Info(logger, "failed to validate dataplane: "+err.Error(), dataplane)
		r.eventRecorder.Event(dataplane, "Warning", "ValidationFailed", err.Error())
		markErr := r.ensureDataPlaneIsMarkedNotReady(ctx, logger, dataplane, DataPlaneConditionValidationFailed, err.Error())
		return ctrl.Result{}, markErr
	}

	log.Trace(logger, "exposing DataPlane deployment admin API via headless service", dataplane)
	res, dataplaneAdminService, err := ensureAdminServiceForDataPlane(ctx, r.Client, dataplane,
		client.MatchingLabels{
			consts.DataPlaneServiceStateLabel: consts.DataPlaneStateLabelValueLive,
		},
		k8sresources.LabelSelectorFromDataPlaneStatusSelectorServiceOpt(dataplane),
	)
	if err != nil {
		return ctrl.Result{}, err
	}
	switch res {
	case op.Created, op.Updated:
		log.Debug(logger, "DataPlane admin service modified", dataplane, "service", dataplaneAdminService.Name, "reason", res)
		return ctrl.Result{}, nil // dataplane admin service creation/update will trigger reconciliation
	case op.Noop:
	case op.Deleted: // This should not happen.
	}

	log.Trace(logger, "exposing DataPlane deployment via service", dataplane)
	additionalServiceLabels := map[string]string{
		consts.DataPlaneServiceStateLabel: consts.DataPlaneStateLabelValueLive,
	}
	serviceRes, dataplaneIngressService, err := ensureIngressServiceForDataPlane(
		ctx,
		log.GetLogger(ctx, "dataplane_ingress_service", r.DevelopmentMode),
		r.Client,
		dataplane,
		additionalServiceLabels,
		k8sresources.LabelSelectorFromDataPlaneStatusSelectorServiceOpt(dataplane),
		k8sresources.ServicePortsFromDataPlaneIngressOpt(dataplane),
	)
	if err != nil {
		return ctrl.Result{}, err
	}
	if serviceRes == op.Created || serviceRes == op.Updated {
		log.Debug(logger, "DataPlane ingress service created/updated", dataplane, "service", dataplaneIngressService.Name)
		return ctrl.Result{}, nil
	}

	dataplaneServiceChanged, err := r.ensureDataPlaneServiceStatus(ctx, logger, dataplane, dataplaneIngressService.Name)
	if err != nil {
		return ctrl.Result{}, err
	}
	if dataplaneServiceChanged {
		log.Debug(logger, "ingress service updated in the dataplane status", dataplane)
		return ctrl.Result{}, nil // dataplane status update will trigger reconciliation
	}

	log.Trace(logger, "ensuring mTLS certificate", dataplane)
	res, certSecret, err := ensureDataPlaneCertificate(ctx, r.Client, dataplane,
		types.NamespacedName{
			Namespace: r.ClusterCASecretNamespace,
			Name:      r.ClusterCASecretName,
		},
		types.NamespacedName{
			Namespace: dataplaneAdminService.Namespace,
			Name:      dataplaneAdminService.Name,
		},
	)
	if err != nil {
		return ctrl.Result{}, err
	}
	if res != op.Noop {
		log.Debug(logger, "mTLS certificate created/updated", dataplane)
		return ctrl.Result{}, nil // requeue will be triggered by the creation or update of the owned object
	}

	log.Trace(logger, "checking readiness of DataPlane service", dataplaneIngressService)
	if dataplaneIngressService.Spec.ClusterIP == "" {
		return ctrl.Result{}, nil // no need to requeue, the update will trigger.
	}

	log.Trace(logger, "ensuring DataPlane has service addesses in status", dataplaneIngressService)
	if updated, err := r.ensureDataPlaneAddressesStatus(ctx, logger, dataplane, dataplaneIngressService); err != nil {
		return ctrl.Result{}, err
	} else if updated {
		log.Debug(logger, "dataplane status.Addresses updated", dataplane)
		return ctrl.Result{}, nil // no need to requeue, the update will trigger.
	}

	deploymentLabels := client.MatchingLabels{
		consts.DataPlaneDeploymentStateLabel: consts.DataPlaneStateLabelValueLive,
	}
	deploymentOpts := []k8sresources.DeploymentOpt{
		labelSelectorFromDataPlaneStatusSelectorDeploymentOpt(dataplane),
	}
	deploymentBuilder := NewDeploymentBuilder(logger.WithName("deployment_builder"), r.Client).
		WithBeforeCallbacks(r.Callbacks.BeforeDeployment).
		WithAfterCallbacks(r.Callbacks.AfterDeployment).
		WithClusterCertificate(certSecret.Name).
		WithOpts(deploymentOpts...).
		WithDefaultImage(r.DefaultImage).
		WithAdditionalLabels(deploymentLabels)

	deployment, res, err := deploymentBuilder.BuildAndDeploy(ctx, dataplane, r.DevelopmentMode)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("could not build Deployment for DataPlane %s: %w",
			dpNn, err)
	}
	if res != op.Noop {
		return ctrl.Result{}, nil
	}

	res, _, err = ensureHPAForDataPlane(ctx, r.Client, logger, dataplane, deployment.Name)
	if err != nil {
		return ctrl.Result{}, err
	}
	if res != op.Noop {
		return ctrl.Result{}, nil
	}

	res, _, err = ensurePodDisruptionBudgetForDataPlane(ctx, r.Client, logger, dataplane)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("could not ensure PodDisruptionBudget for DataPlane %s: %w", dpNn, err)
	}
	if res != op.Noop {
		log.Debug(logger, "PodDisruptionBudget created/updated", dataplane)
		return ctrl.Result{}, nil
	}

	if res, err := ensureDataPlaneReadyStatus(ctx, r.Client, logger, dataplane, dataplane.Generation); err != nil {
		return ctrl.Result{}, err
	} else if res.Requeue {
		return res, nil
	}

	log.Debug(logger, "reconciliation complete for DataPlane resource", dataplane)
	return ctrl.Result{}, nil
}

func (r *Reconciler) initSelectorInStatus(ctx context.Context, logger logr.Logger, dataplane *operatorv1beta1.DataPlane) error {
	if dataplane.Status.Selector != "" {
		return nil
	}

	dataplane.Status.Selector = uuid.New().String()
	_, err := patchDataPlaneStatus(ctx, r.Client, logger, dataplane)
	return err
}

// labelSelectorFromDataPlaneStatusSelectorDeploymentOpt returns a DeploymentOpt
// function which will set Deployment's selector and spec template labels, based
// on provided DataPlane's Status selector field.
func labelSelectorFromDataPlaneStatusSelectorDeploymentOpt(dataplane *operatorv1beta1.DataPlane) func(s *appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if dataplane.Status.Selector != "" {
			d.Labels[consts.OperatorLabelSelector] = dataplane.Status.Selector
			d.Spec.Selector.MatchLabels[consts.OperatorLabelSelector] = dataplane.Status.Selector
			d.Spec.Template.Labels[consts.OperatorLabelSelector] = dataplane.Status.Selector
		}
	}
}
