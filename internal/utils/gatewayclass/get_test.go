package gatewayclass_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	operatorerrors "github.com/kong/gateway-operator/internal/errors"
	"github.com/kong/gateway-operator/internal/utils/gatewayclass"
	"github.com/kong/gateway-operator/modules/manager/scheme"
	"github.com/kong/gateway-operator/pkg/vars"
)

func TestGet(t *testing.T) {
	testCases := []struct {
		name             string
		gatewayClassName string
		objectsToAdd     []client.Object
		expectedError    error
	}{
		{
			name:             "gateway class not found",
			gatewayClassName: "non-existent-gateway-class",
			objectsToAdd:     []client.Object{},
			expectedError: fmt.Errorf(`error while fetching GatewayClass "non-existent-gateway-class": %w`,
				errors.New(`gatewayclasses.gateway.networking.k8s.io "non-existent-gateway-class" not found`)),
		},
		{
			name:             "gateway class not supported",
			gatewayClassName: "gateway-class-1",
			objectsToAdd: []client.Object{&gatewayv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-class-1",
				},
				Spec: gatewayv1.GatewayClassSpec{
					ControllerName: "some-other-controller",
				},
			}},
			expectedError: operatorerrors.NewErrUnsupportedGateway(
				`GatewayClass "gateway-class-1" with "some-other-controller" ` +
					`ControllerName does not match expected "konghq.com/gateway-operator"`,
			),
		},
		{
			name:             "empty gateway class name",
			gatewayClassName: "",
			objectsToAdd:     []client.Object{},
			expectedError: operatorerrors.NewErrUnsupportedGateway(
				"no GatewayClassName provided",
			),
		},
		{
			name:             "gateway class supported",
			gatewayClassName: "gateway-class-2",
			objectsToAdd: []client.Object{&gatewayv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gateway-class-2",
				},
				Spec: gatewayv1.GatewayClassSpec{
					ControllerName: gatewayv1.GatewayController(vars.ControllerName()),
				},
			}},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClient := fake.
				NewClientBuilder().
				WithScheme(scheme.Get()).
				WithObjects(tc.objectsToAdd...).
				Build()

			gwc, err := gatewayclass.Get(context.Background(), fakeClient, tc.gatewayClassName)
			if tc.expectedError != nil {
				require.ErrorContains(t, err, tc.expectedError.Error())
				require.IsType(t, tc.expectedError, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, gwc, "returned decorator should not be nil")
			require.NotNil(t, gwc.GatewayClass, "decorator's GatewayClass should not be nil")
		})
	}
}
