package konnect

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	fakectrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kong/gateway-operator/modules/manager/scheme"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

func TestNewKonnectEntityReconciler(t *testing.T) {
	testNewKonnectEntityReconciler(t, konnectv1alpha1.KonnectControlPlane{})
	testNewKonnectEntityReconciler(t, configurationv1alpha1.KongService{})

	// TODO(pmalek): add support for KongRoute
	// https://github.com/Kong/gateway-operator/issues/435
	// testNewKonnectEntityReconciler(t, configurationv1alpha1.KongRoute{})

	// TODO(pmalek): add support for KongConsumer
	// https://github.com/Kong/gateway-operator/issues/436
	// testNewKonnectEntityReconciler(t, configurationv1.KongConsumer{})

	// TODO: GetConditions() and SetConditions() is missing from KongConsumerGroup.
	// testNewKonnectEntityReconciler(t, configurationv1beta1.KongConsumerGroup{})
}

func testNewKonnectEntityReconciler[
	T SupportedKonnectEntityType,
	TEnt EntityType[T],
](
	t *testing.T,
	ent T,
) {
	t.Helper()

	sdkFactory := NewSDKFactory()

	t.Run(ent.GetTypeName(), func(t *testing.T) {
		cl := fakectrlruntimeclient.NewFakeClient()
		mgr, err := ctrl.NewManager(&rest.Config{}, ctrl.Options{
			Scheme: scheme.Get(),
		})
		require.NoError(t, err)
		reconciler := NewKonnectEntityReconciler[T, TEnt](sdkFactory, false, cl)
		require.NoError(t, reconciler.SetupWithManager(mgr))
	})
}