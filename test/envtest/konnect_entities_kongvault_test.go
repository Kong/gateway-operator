package envtest

import (
	"context"
	"testing"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectops "github.com/Kong/sdk-konnect-go/models/operations"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kong/gateway-operator/controller/konnect"
	"github.com/kong/gateway-operator/controller/konnect/conditions"
	konnectops "github.com/kong/gateway-operator/controller/konnect/ops"
	"github.com/kong/gateway-operator/modules/manager/scheme"
	"github.com/kong/gateway-operator/test/helpers/deploy"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

func TestKongVault(t *testing.T) {
	t.Parallel()
	ctx, cancel := Context(t, context.Background())
	defer cancel()
	cfg, ns := Setup(t, ctx, scheme.Get())

	t.Log("Setting up the manager with reconcilers")
	mgr, logs := NewManager(t, ctx, cfg, scheme.Get())
	factory := konnectops.NewMockSDKFactory(t)
	sdk := factory.SDK
	reconcilers := []Reconciler{
		konnect.NewKonnectEntityReconciler(factory, false, mgr.GetClient(),
			konnect.WithKonnectEntitySyncPeriod[configurationv1alpha1.KongVault](konnectInfiniteSyncTime),
		),
	}
	StartReconcilers(ctx, t, mgr, logs, reconcilers...)

	t.Log("Setting up clients")
	cl, err := client.NewWithWatch(mgr.GetConfig(), client.Options{
		Scheme: scheme.Get(),
	})
	require.NoError(t, err)
	clientNamespaced := client.NewNamespacedClient(mgr.GetClient(), ns.Name)

	t.Log("Creating KonnectAPIAuthConfiguration and KonnectGatewayControlPlane")
	apiAuth := deploy.KonnectAPIAuthConfigurationWithProgrammed(t, ctx, clientNamespaced)
	cp := deploy.KonnectGatewayControlPlaneWithID(t, ctx, clientNamespaced, apiAuth)

	t.Log("Setting up a watch for KongVault events")
	vaultWatch := setupWatch[configurationv1alpha1.KongVaultList](t, ctx, cl)

	t.Run("Should create, update and delete vault successfully", func(t *testing.T) {
		const (
			vaultBackend     = "env"
			vaultPrefix      = "env-vault"
			vaultRawConfig   = `{"prefix":"env_vault"}`
			vaultDespription = "test-env-vault"

			vaultID = "vault-12345"
		)

		t.Log("Setting up mock SDK for vault creation")
		sdk.VaultSDK.EXPECT().CreateVault(mock.Anything, cp.GetKonnectStatus().GetKonnectID(), mock.MatchedBy(func(input sdkkonnectcomp.VaultInput) bool {
			return input.Name == vaultBackend && input.Prefix == vaultPrefix
		})).Return(&sdkkonnectops.CreateVaultResponse{
			Vault: &sdkkonnectcomp.Vault{
				ID: lo.ToPtr(vaultID),
			},
		}, nil)

		vault := deploy.KongVaultAttachedToCP(t, ctx, cl, vaultBackend, vaultPrefix, []byte(vaultRawConfig), cp)

		t.Log("Waiting for KongVault to be programmed")
		watchFor(t, ctx, vaultWatch, watch.Modified, func(v *configurationv1alpha1.KongVault) bool {
			if v.GetName() != vault.GetName() {
				return false
			}

			return lo.ContainsBy(v.Status.Conditions, func(condition metav1.Condition) bool {
				return condition.Type == conditions.KonnectEntityProgrammedConditionType &&
					condition.Status == metav1.ConditionTrue
			})
		}, "KongVault's Programmed condition should be true eventually")

		t.Log("Waiting for KongVault to be created in the SDK")
		require.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, factory.SDK.VaultSDK.AssertExpectations(t))
		}, waitTime, tickTime)

		t.Log("Setting up mock SDK for vault update")
		sdk.VaultSDK.EXPECT().UpsertVault(mock.Anything, mock.MatchedBy(func(r sdkkonnectops.UpsertVaultRequest) bool {
			return r.VaultID == vaultID &&
				r.Vault.Name == vaultBackend &&
				r.Vault.Description != nil && *r.Vault.Description == vaultDespription
		})).Return(&sdkkonnectops.UpsertVaultResponse{}, nil)

		t.Log("Patching KongVault")
		vaultToPatch := vault.DeepCopy()
		vaultToPatch.Spec.Description = vaultDespription
		require.NoError(t, clientNamespaced.Patch(ctx, vaultToPatch, client.MergeFrom(vault)))

		t.Log("Waiting for KongVault to be updated in the SDK")
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, factory.SDK.ConsumersSDK.AssertExpectations(t))
		}, waitTime, tickTime)

		t.Log("Setting up mock SDK for vault deletion")
		sdk.VaultSDK.EXPECT().DeleteVault(mock.Anything, cp.GetKonnectStatus().GetKonnectID(), vaultID).
			Return(&sdkkonnectops.DeleteVaultResponse{}, nil)

		t.Log("Deleting KongVault")
		require.NoError(t, cl.Delete(ctx, vault))

		t.Log("Waiting for vault to be deleted in the SDK")
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, factory.SDK.VaultSDK.AssertExpectations(t))
		}, waitTime, tickTime)
	})
}
