package specialized

import (
	"context"
	"errors"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/kong/gateway-operator/api/v1alpha1"
	"github.com/kong/gateway-operator/controller/pkg/log"
	"github.com/kong/gateway-operator/controller/pkg/watch"
	operatorerrors "github.com/kong/gateway-operator/internal/errors"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
	"github.com/kong/gateway-operator/pkg/vars"
)

// ----------------------------------------------------------------------------
// AIGatewayReconciler
// ----------------------------------------------------------------------------

// AIGatewayReconciler reconciles a AIGateway object
type AIGatewayReconciler struct {
	client.Client

	Scheme          *runtime.Scheme
	DevelopmentMode bool
}

// SetupWithManager sets up the controller with the Manager.
func (r *AIGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// watch AIGateway objects, filtering out any Gateways which are not
		// configured with a supported GatewayClass controller name.
		For(&v1alpha1.AIGateway{},
			builder.WithPredicates(predicate.NewPredicateFuncs(r.aiGatewayHasMatchingGatewayClass))).
		Watches(
			&gatewayv1.GatewayClass{},
			handler.EnqueueRequestsFromMapFunc(r.listAIGatewaysForGatewayClass),
			builder.WithPredicates(predicate.NewPredicateFuncs(watch.GatewayClassMatchesController)),
		).
		// TODO watch on Gateways, KongPlugins, e.t.c.
		//
		// See: https://github.com/Kong/gateway-operator/issues/1368
		Complete(r)
}

// Reconcile reconciles the AIGateway resource.
func (r *AIGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.GetLogger(ctx, "aigateway", r.DevelopmentMode)

	var aigateway v1alpha1.AIGateway
	if err := r.Client.Get(ctx, req.NamespacedName, &aigateway); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	log.Trace(logger, "verifying gatewayclass for aigateway", aigateway)
	// we verify the GatewayClass in the watch predicates as well, but the watch
	// predicates are known to be lossy, so they are considered only an optimization
	// and this check must be done here to ensure consistency.
	//
	// See: https://github.com/kubernetes-sigs/controller-runtime/issues/1996
	gwc, err := r.verifyGatewayClassSupport(ctx, &aigateway)
	if err != nil {
		if errors.Is(err, operatorerrors.ErrUnsupportedGateway) {
			log.Debug(logger, "resource not supported, ignoring", aigateway, "ExpectedGatewayClass", vars.ControllerName())
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if !gwc.IsAccepted() {
		log.Debug(logger, "gatewayclass for aigateway is not accepted, ignoring", aigateway)
		return ctrl.Result{}, nil
	}

	log.Trace(logger, "handling any necessary aigateway cleanup", aigateway)
	if aigateway.GetDeletionTimestamp() != nil {
		log.Debug(logger, "aigateway is being deleted, ignoring", aigateway)
		return ctrl.Result{}, nil
	}

	log.Trace(logger, "marking aigateway as accepted", aigateway)
	oldAIGateway := aigateway.DeepCopy()
	k8sutils.SetCondition(newAIGatewayAcceptedCondition(&aigateway), &aigateway)
	if k8sutils.NeedsUpdate(oldAIGateway, &aigateway) {
		if err := r.Client.Status().Patch(ctx, &aigateway, client.MergeFrom(oldAIGateway)); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to patch status for aigateway: %w", err)
		}
		log.Info(logger, "aigateway marked as accepted", aigateway)
		return ctrl.Result{}, nil // update will re-queue
	}

	log.Info(logger, "managing gateway resources for aigateway", aigateway)
	gatewayResourcesChanged, err := r.manageGateway(ctx, logger, &aigateway)
	if err != nil {
		return ctrl.Result{}, err
	}
	if gatewayResourcesChanged {
		return ctrl.Result{Requeue: true}, nil
	}

	log.Info(logger, "configuring plugin and route resources for aigateway", aigateway)
	pluginResourcesChanged, err := r.configurePlugins(ctx, logger, &aigateway)
	if err != nil {
		return ctrl.Result{}, err
	}
	if pluginResourcesChanged {
		return ctrl.Result{Requeue: true}, err
	}

	// TODO: manage status updates
	//
	// See: https://github.com/Kong/gateway-operator/issues/1368

	log.Info(logger, "reconciliation complete for aigateway resource", aigateway)
	return ctrl.Result{}, nil
}
