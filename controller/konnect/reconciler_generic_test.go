package konnect

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	fakectrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kong/gateway-operator/controller/konnect/constraints"
	"github.com/kong/gateway-operator/controller/konnect/ops"
	"github.com/kong/gateway-operator/modules/manager/scheme"

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	configurationv1beta1 "github.com/kong/kubernetes-configuration/api/configuration/v1beta1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

func TestNewKonnectEntityReconciler(t *testing.T) {
	testNewKonnectEntityReconciler(t, konnectv1alpha1.KonnectGatewayControlPlane{})
	testNewKonnectEntityReconciler(t, configurationv1alpha1.KongService{})
	testNewKonnectEntityReconciler(t, configurationv1.KongConsumer{})
	testNewKonnectEntityReconciler(t, configurationv1alpha1.KongRoute{})
	testNewKonnectEntityReconciler(t, configurationv1beta1.KongConsumerGroup{})
	// TODO: Reconcilers setting index require a real k8s API server:
	// https://github.com/kubernetes-sigs/controller-runtime/issues/657
	// Maybe we should import envtest.
	// testNewKonnectEntityReconciler(t, configurationv1alpha1.KongPluginBinding{})
}

func testNewKonnectEntityReconciler[
	T constraints.SupportedKonnectEntityType,
	TEnt constraints.EntityType[T],
](
	t *testing.T,
	ent T,
) {
	t.Helper()

	// TODO: use a mock Konnect SDK factory here and use envtest to trigger real reconciliations and Konnect requests
	sdkFactory := &ops.MockSDKFactory{}

	t.Run(ent.GetTypeName(), func(t *testing.T) {
		cl := fakectrlruntimeclient.NewFakeClient()
		mgr, err := ctrl.NewManager(&rest.Config{}, ctrl.Options{
			Scheme: scheme.Get(),
		})
		require.NoError(t, err)

		reconciler := NewKonnectEntityReconciler[T, TEnt](sdkFactory, false, cl)
		require.NoError(t, reconciler.SetupWithManager(context.Background(), mgr))
	})
}
