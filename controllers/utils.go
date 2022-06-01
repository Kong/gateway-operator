package controllers

import (
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kong/gateway-operator/internal/manager/logging"
)

// -----------------------------------------------------------------------------
// Private Functions - Logging
// -----------------------------------------------------------------------------

func debug(log logr.Logger, msg string, rawOBJ interface{}, keysAndValues ...interface{}) { //nolint:unparam //FIXME
	if obj, ok := rawOBJ.(client.Object); ok {
		kvs := append([]interface{}{"namespace", obj.GetNamespace(), "name", obj.GetName()}, keysAndValues...)
		log.V(logging.DebugLevel).Info(msg, kvs...)
	} else if req, ok := rawOBJ.(reconcile.Request); ok {
		kvs := append([]interface{}{"namespace", req.Namespace, "name", req.Name}, keysAndValues...)
		log.V(logging.DebugLevel).Info(msg, kvs...)
	} else {
		log.V(logging.DebugLevel).Info(fmt.Sprintf("unexpected type processed for debug logging: %T, this is a bug!", rawOBJ))
	}
}
