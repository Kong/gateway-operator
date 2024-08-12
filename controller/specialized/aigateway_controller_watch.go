package specialized

import (
	"context"
	"errors"
	"reflect"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/kong/gateway-operator/api/v1alpha1"
	operatorerrors "github.com/kong/gateway-operator/internal/errors"
)

// -----------------------------------------------------------------------------
// AIGatewayReconciler - Watch Predicates
// -----------------------------------------------------------------------------

func (r *AIGatewayReconciler) aiGatewayHasMatchingGatewayClass(obj client.Object) bool {
	aigateway, ok := obj.(*v1alpha1.AIGateway)
	if !ok {
		ctrllog.FromContext(context.Background()).Error(
			operatorerrors.ErrUnexpectedObject,
			"failed to run predicate function",
			"expected", "Gateway", "found", reflect.TypeOf(obj),
		)
		return false
	}

	_, err := r.verifyGatewayClassSupport(context.Background(), aigateway)
	if err != nil {
		// filtering here is just an optimization, the reconciler will check the
		// class as well. If we fail here it's most likely because of some failure
		// of the Kubernetes API and it's technically better to enqueue the object
		// than to drop it for eventual consistency during cluster outages.
		return !errors.Is(err, operatorerrors.ErrUnsupportedGateway)
	}

	return true
}

// -----------------------------------------------------------------------------
// AIGatewayReconciler - Watch Mapping Funcs
// -----------------------------------------------------------------------------

func (r *AIGatewayReconciler) listAIGatewaysForGatewayClass(ctx context.Context, obj client.Object) (recs []reconcile.Request) {
	gatewayClass, ok := obj.(*gatewayv1.GatewayClass)
	if !ok {
		ctrllog.FromContext(ctx).Error(
			operatorerrors.ErrUnexpectedObject,
			"failed to run map funcs",
			"expected", "GatewayClass", "found", reflect.TypeOf(obj),
		)
		return
	}

	aigateways := new(v1alpha1.AIGatewayList)
	if err := r.Client.List(ctx, aigateways); err != nil {
		ctrllog.FromContext(ctx).Error(err, "could not list aigateways in map func")
		return
	}

	for _, aigateway := range aigateways.Items {
		if aigateway.Spec.GatewayClassName == gatewayClass.Name {
			recs = append(recs, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: aigateway.Namespace,
					Name:      aigateway.Name,
				},
			})
		}
	}

	return
}
