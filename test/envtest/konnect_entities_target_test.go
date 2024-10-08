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
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kong/gateway-operator/controller/konnect"
	"github.com/kong/gateway-operator/controller/konnect/ops"
	"github.com/kong/gateway-operator/modules/manager/scheme"
	"github.com/kong/gateway-operator/test/helpers/deploy"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

func TestKongTarget(t *testing.T) {
	t.Parallel()
	ctx, cancel := Context(t, context.Background())
	defer cancel()
	cfg, ns := Setup(t, ctx, scheme.Get())

	t.Log("Setting up the manager with reconcilers")
	mgr, logs := NewManager(t, ctx, cfg, scheme.Get())
	factory := ops.NewMockSDKFactory(t)
	sdk := factory.SDK
	StartReconcilers(ctx, t, mgr, logs,
		konnect.NewKonnectEntityReconciler(factory, false, mgr.GetClient(),
			konnect.WithKonnectEntitySyncPeriod[configurationv1alpha1.KongTarget](konnectInfiniteSyncTime),
		),
	)

	t.Log("Setting up clients")
	cl, err := client.NewWithWatch(mgr.GetConfig(), client.Options{
		Scheme: scheme.Get(),
	})
	require.NoError(t, err)
	clientNamespaced := client.NewNamespacedClient(mgr.GetClient(), ns.Name)

	t.Log("Creating KonnectAPIAuthConfiguration and KonnectGatewayControlPlane")
	apiAuth := deploy.KonnectAPIAuthConfigurationWithProgrammed(t, ctx, clientNamespaced)
	cp := deploy.KonnectGatewayControlPlaneWithID(t, ctx, clientNamespaced, apiAuth)

	t.Run("adding, patching and deleting KongTarget", func(t *testing.T) {
		const (
			upstreamID   = "kup-12345"
			targetID     = "target-12345"
			targetHost   = "example.com"
			targetWeight = 100
		)

		t.Log("Creating a KongUpstream and setting it to programmed")
		upstream := deploy.KongUpstreamAttachedToCP(t, ctx, clientNamespaced, cp)
		updateKongUpstreamStatusWithProgrammed(t, ctx, clientNamespaced, upstream, upstreamID, cp.GetKonnectID())

		t.Log("Setting up a watch for KongTarget events")
		w := setupWatch[configurationv1alpha1.KongTargetList](t, ctx, cl, client.InNamespace(ns.Name))

		t.Log("Setting up SDK expectations on Target creation")
		sdk.TargetsSDK.EXPECT().CreateTargetWithUpstream(
			mock.Anything,
			mock.MatchedBy(func(req sdkkonnectops.CreateTargetWithUpstreamRequest) bool {
				return *req.TargetWithoutParents.Target == targetHost && *req.TargetWithoutParents.Weight == int64(targetWeight)
			}),
		).Return(&sdkkonnectops.CreateTargetWithUpstreamResponse{
			Target: &sdkkonnectcomp.Target{
				ID: lo.ToPtr(targetID),
			},
		}, nil)

		t.Log("Creating a KongTarget")
		createdTarget := deploy.KongTargetAttachedToUpstream(t, ctx, clientNamespaced, upstream,
			func(obj client.Object) {
				kt := obj.(*configurationv1alpha1.KongTarget)
				kt.Spec.KongTargetAPISpec.Target = targetHost
				kt.Spec.KongTargetAPISpec.Weight = targetWeight
			},
		)
		t.Log("Checking SDK KongTarget operations")
		require.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, factory.SDK.TargetsSDK.AssertExpectations(t))
		}, waitTime, tickTime)

		t.Log("Waiting for Target to be programmed and get Konnect ID")
		watchFor(t, ctx, w, watch.Modified, func(kt *configurationv1alpha1.KongTarget) bool {
			return kt.GetKonnectID() == targetID && lo.ContainsBy(kt.Status.Conditions,
				func(c metav1.Condition) bool {
					return c.Type == "Programmed" && c.Status == metav1.ConditionTrue
				})
		}, "KongTarget didn't get Programmed status condition or didn't get the correct (target-12345) Konnect ID assigned")

		t.Log("Setting up SDK expectations on Target update")
		sdk.TargetsSDK.EXPECT().UpsertTargetWithUpstream(
			mock.Anything,
			mock.MatchedBy(func(req sdkkonnectops.UpsertTargetWithUpstreamRequest) bool {
				return req.TargetID == targetID && *req.TargetWithoutParents.Weight == int64(200)
			}),
		).Return(&sdkkonnectops.UpsertTargetWithUpstreamResponse{}, nil)

		t.Log("Patching KongTarget")
		targetToPatch := createdTarget.DeepCopy()
		targetToPatch.Spec.Weight = 200
		require.NoError(t, clientNamespaced.Patch(ctx, targetToPatch, client.MergeFrom(createdTarget)))

		t.Log("Waiting for Target to be updated in the SDK")
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, factory.SDK.TargetsSDK.AssertExpectations(t))
		}, waitTime, tickTime)

		t.Log("Setting up SDK expectations on Target deletion")
		sdk.TargetsSDK.EXPECT().DeleteTargetWithUpstream(
			mock.Anything,
			mock.MatchedBy(func(req sdkkonnectops.DeleteTargetWithUpstreamRequest) bool {
				return req.TargetID == targetID
			}),
		).Return(&sdkkonnectops.DeleteTargetWithUpstreamResponse{}, nil)

		t.Log("Deleting KongTarget")
		require.NoError(t, clientNamespaced.Delete(ctx, createdTarget))

		t.Log("Waiting for KongTarget to disappear")
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			err := clientNamespaced.Get(ctx, client.ObjectKeyFromObject(createdTarget), createdTarget)
			assert.True(c, err != nil && k8serrors.IsNotFound(err))
		}, waitTime, tickTime)

		t.Log("Waiting for Target to be deleted in the SDK")
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, factory.SDK.TargetsSDK.AssertExpectations(t))
		}, waitTime, tickTime)

	})
}
