package watch

import (
	"context"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	operatorerrors "github.com/kong/gateway-operator/internal/errors"
	"github.com/kong/gateway-operator/pkg/vars"
)

// -----------------------------------------------------------------------------
// GatewayClass - Watch Predicates
// -----------------------------------------------------------------------------

// GatewayClassMatchesController is a controller runtime watch predicate
// function which can be used to determine whether a given GatewayClass
// belongs to and is served by the current controller.
func GatewayClassMatchesController(obj client.Object) bool {
	gatewayClass, ok := obj.(*gatewayv1.GatewayClass)
	if !ok {
		log.FromContext(context.Background()).Error(
			operatorerrors.ErrUnexpectedObject,
			"failed to run predicate function",
			"expected", "GatewayClass", "found", reflect.TypeOf(obj),
		)
		return false
	}

	return string(gatewayClass.Spec.ControllerName) == vars.ControllerName()
}
