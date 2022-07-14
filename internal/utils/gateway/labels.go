package gateway

import (
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/kong/gateway-operator/internal/consts"
	k8sutils "github.com/kong/gateway-operator/internal/utils/kubernetes"
)

// -----------------------------------------------------------------------------
// Gateway Utils - Labels
// -----------------------------------------------------------------------------

// LabelObjectAsGatewayManaged ensures that labels are set on the provided
// object to signal that it's owned by a Gateway resource and that it's
// lifecycle is managed by this operator.
func LabelObjectAsGatewayManaged(obj client.Object) {
	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[consts.GatewayOperatorControlledLabel] = consts.GatewayManagedLabelValue
	obj.SetLabels(labels)
}

// NewGatewayManagedListSelectorOption returns a ListOption that limits the
// objects returned by List operation to owned by a Gateway resource
// and managed by this operator.
func NewGatewayManagedListSelectorOption(gateway *gatewayv1alpha2.Gateway) (*client.ListOptions, error) {
	return k8sutils.NewListSelectorOption(
		gateway.Namespace,
		consts.GatewayOperatorControlledLabel,
		selection.Equals,
		consts.GatewayManagedLabelValue,
	)
}
