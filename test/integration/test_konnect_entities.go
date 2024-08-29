package integration

import (
	"testing"
	"time"

	"github.com/Kong/sdk-konnect-go/models/components"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	testutils "github.com/kong/gateway-operator/pkg/utils/test"
	"github.com/kong/gateway-operator/test"
	"github.com/kong/gateway-operator/test/helpers"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

func TestKonnectEntities(t *testing.T) {
	ns, _ := helpers.SetupTestEnv(t, GetCtx(), GetEnv())

	testID := uuid.NewString()[:8]
	t.Logf("Running Konnect entities test with ID: %s", testID)

	t.Logf("Creating KonnectAPIAuthConfiguration")
	authCfg := &konnectv1alpha1.KonnectAPIAuthConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "auth-" + testID,
			Namespace: ns.Name,
		},
		Spec: konnectv1alpha1.KonnectAPIAuthConfigurationSpec{
			Type:      konnectv1alpha1.KonnectAPIAuthTypeToken,
			Token:     test.KonnectAccessToken(),
			ServerURL: test.KonnectServerURL(),
		},
	}
	err := GetClients().MgrClient.Create(GetCtx(), authCfg)
	require.NoError(t, err)

	cpName := "cp-" + testID
	t.Logf("Creating KonnectControlPlane %s", cpName)
	cp := &konnectv1alpha1.KonnectControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cpName,
			Namespace: ns.Name,
		},
		Spec: konnectv1alpha1.KonnectControlPlaneSpec{
			CreateControlPlaneRequest: components.CreateControlPlaneRequest{
				Name:        cpName,
				ClusterType: lo.ToPtr(components.ClusterTypeClusterTypeControlPlane),
				Labels:      map[string]string{"test_id": testID},
			},
			KonnectConfiguration: konnectv1alpha1.KonnectConfiguration{
				APIAuthConfigurationRef: konnectv1alpha1.KonnectAPIAuthConfigurationRef{
					Name: authCfg.Name,
				},
			},
		},
	}
	err = GetClients().MgrClient.Create(GetCtx(), cp)
	require.NoError(t, err)
	t.Cleanup(deleteObjectAndWaitForDeletionFn(t, cp))

	t.Logf("Waiting for Konnect ID to be assigned to ControlPlane %s/%s", cp.Namespace, cp.Name)
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		err := GetClients().MgrClient.Get(GetCtx(), types.NamespacedName{Name: cp.Name, Namespace: cp.Namespace}, cp)
		require.NoError(t, err)
		assert.NotEmpty(t, cp.Status.KonnectEntityStatus.GetKonnectID())
		assert.NotEmpty(t, cp.Status.KonnectEntityStatus.GetOrgID())
		assert.NotEmpty(t, cp.Status.KonnectEntityStatus.GetServerURL())
	}, testutils.ObjectUpdateTimeout, time.Second)

	t.Logf("Creating KongService")
	ksName := "ks-" + testID
	ks := &configurationv1alpha1.KongService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ks-" + testID,
			Namespace: ns.Name,
		},
		Spec: configurationv1alpha1.KongServiceSpec{
			ControlPlaneRef: &configurationv1alpha1.ControlPlaneRef{
				Type:                 configurationv1alpha1.ControlPlaneRefKonnectNamespacedRef,
				KonnectNamespacedRef: &configurationv1alpha1.KonnectNamespacedRef{Name: cp.Name},
			},
			KongServiceAPISpec: configurationv1alpha1.KongServiceAPISpec{
				Name: lo.ToPtr(ksName),
				URL:  lo.ToPtr("http://example.com"),
			},
		},
	}
	err = GetClients().MgrClient.Create(GetCtx(), ks)
	require.NoError(t, err)

	t.Logf("Waiting for KongService to be updated with Konnect ID")
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		err := GetClients().MgrClient.Get(GetCtx(), types.NamespacedName{Name: ks.Name, Namespace: ks.Namespace}, ks)
		require.NoError(t, err)
		assert.NotEmpty(t, ks.Status.Konnect.KonnectEntityStatus.GetKonnectID())
		assert.NotEmpty(t, ks.Status.Konnect.KonnectEntityStatus.GetOrgID())
		assert.NotEmpty(t, ks.Status.Konnect.KonnectEntityStatus.GetServerURL())
	}, testutils.ObjectUpdateTimeout, time.Second)

	// TODO(czeslavo): test all supported entities here
}

// deleteObjectAndWaitForDeletionFn returns a function that deletes the given object and waits for it to be gone.
// It's designed to be used with t.Cleanup() to ensure the object is properly deleted (it's not stuck with finalizers, etc.).
func deleteObjectAndWaitForDeletionFn(t *testing.T, obj client.Object) func() {
	return func() {
		err := GetClients().MgrClient.Delete(GetCtx(), obj)
		require.NoError(t, err)

		require.EventuallyWithT(t, func(t *assert.CollectT) {
			err := GetClients().MgrClient.Get(GetCtx(), types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}, obj)
			assert.True(t, k8serrors.IsNotFound(err))
		}, testutils.ObjectUpdateTimeout, time.Second)
	}
}
