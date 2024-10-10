package dataplane

import (
	"context"
	"reflect"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
	operatorv1beta1 "github.com/kong/gateway-operator/api/v1beta1"
	"github.com/kong/gateway-operator/controller/pkg/ctxinjector"
	"github.com/kong/gateway-operator/controller/pkg/log"
	operatorerrors "github.com/kong/gateway-operator/internal/errors"
	"github.com/kong/gateway-operator/internal/utils/index"
	"github.com/kong/gateway-operator/pkg/consts"
)

// -----------------------------------------------------------------------------
// DataKonnectExtensionReconciler
// -----------------------------------------------------------------------------

// DataPlaneKonnectExtensionReconciler reconciles a DataPlaneKonnectExtension object.
type DataPlaneKonnectExtensionReconciler struct {
	client.Client
	ContextInjector ctxinjector.CtxInjector
	// DevelopmentMode indicates if the controller should run in development mode,
	// which causes it to e.g. perform less validations.
	DevelopmentMode bool
}

// SetupWithManager sets up the controller with the Manager.
func (r *DataPlaneKonnectExtensionReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.DataPlaneKonnectExtension{}).
		Watches(&operatorv1beta1.DataPlane{}, handler.EnqueueRequestsFromMapFunc(r.listDataPlaneExtensionsReferenced)).
		Complete(r)
}

// listDataPlaneExtensionsReferenced returns a list of all the DataPlaneKonnectExtensions referenced by the DataPlane object.
// Maximum one reference is expected.
func (r *DataPlaneKonnectExtensionReconciler) listDataPlaneExtensionsReferenced(ctx context.Context, obj client.Object) []reconcile.Request {
	logger := ctrllog.FromContext(ctx)
	dataPlane, ok := obj.(*operatorv1beta1.DataPlane)
	if !ok {
		logger.Error(
			operatorerrors.ErrUnexpectedObject,
			"failed to run map funcs",
			"expected", "DataPlane", "found", reflect.TypeOf(obj),
		)
		return nil
	}

	if len(dataPlane.Spec.Extensions) == 0 {
		return nil
	}

	recs := []reconcile.Request{}

	for _, ext := range dataPlane.Spec.Extensions {
		namespace := dataPlane.Namespace
		if ext.Group != operatorv1alpha1.SchemeGroupVersion.Group ||
			ext.Kind != operatorv1alpha1.DataPlaneKonnectExtensionKind {
			continue
		}
		if ext.Namespace != nil && *ext.Namespace != namespace {
			continue
		}
		recs = append(recs, reconcile.Request{
			NamespacedName: client.ObjectKey{
				Namespace: namespace,
				Name:      ext.Name,
			},
		})
	}
	return recs
}

// Reconcile reconciles a DataPlaneKonnectExtension object.
func (r *DataPlaneKonnectExtensionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = r.ContextInjector.InjectKeyValues(ctx)
	var konnectExtension operatorv1alpha1.DataPlaneKonnectExtension
	if err := r.Client.Get(ctx, req.NamespacedName, &konnectExtension); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger := log.GetLogger(ctx, operatorv1alpha1.DataPlaneKonnectExtensionKind, r.DevelopmentMode)
	var dataPlaneList operatorv1beta1.DataPlaneList
	if err := r.List(ctx, &dataPlaneList, client.MatchingFields{
		index.DataPlaneKonnectExtensionIndex: konnectExtension.Namespace + "/" + konnectExtension.Name,
	}); err != nil {
		return ctrl.Result{}, err
	}

	var updated bool
	switch len(dataPlaneList.Items) {
	case 0:
		updated = controllerutil.RemoveFinalizer(&konnectExtension, consts.DataPlaneExtensionFinalizer)
	default:
		updated = controllerutil.AddFinalizer(&konnectExtension, consts.DataPlaneExtensionFinalizer)
	}
	if updated {
		if err := r.Client.Update(ctx, &konnectExtension); err != nil {
			if k8serrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}

		log.Info(logger, "DataPlaneKonnectExtension finalizer updated", konnectExtension)
	}

	return ctrl.Result{}, nil
}
