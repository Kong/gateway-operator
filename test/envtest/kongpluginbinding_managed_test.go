package envtest

import (
	"context"
	"testing"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectops "github.com/Kong/sdk-konnect-go/models/operations"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kong/gateway-operator/controller/konnect"
	"github.com/kong/gateway-operator/controller/konnect/ops"
	"github.com/kong/gateway-operator/modules/manager"
	"github.com/kong/gateway-operator/modules/manager/scheme"
	"github.com/kong/gateway-operator/pkg/consts"

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

func TestKongPluginBindingManaged(t *testing.T) {
	t.Parallel()
	ctx, cancel := Context(t, context.Background())
	defer cancel()

	// Setup up the envtest environment.
	cfg, ns := Setup(t, ctx, scheme.Get())

	mgr, logs := NewManager(t, ctx, cfg, scheme.Get())

	clientWithWatch, err := client.NewWithWatch(mgr.GetConfig(), client.Options{
		Scheme: scheme.Get(),
	})
	require.NoError(t, err)
	clientNamespaced := client.NewNamespacedClient(mgr.GetClient(), ns.Name)

	apiAuth := deployKonnectAPIAuthConfigurationWithProgrammed(t, ctx, clientNamespaced)
	cp := deployKonnectGatewayControlPlaneWithID(t, ctx, clientNamespaced, apiAuth)

	factory := ops.NewMockSDKFactory(t)
	sdk := factory.SDK

	require.NoError(t, manager.SetupCacheIndicesForKonnectTypes(ctx, mgr, false))
	reconcilers := []Reconciler{
		konnect.NewKongPluginReconciler(false, mgr.GetClient()),
		konnect.NewKonnectEntityReconciler(factory, false, mgr.GetClient(),
			konnect.WithKonnectEntitySyncPeriod[configurationv1alpha1.KongPluginBinding](konnectInfiniteSyncTime),
		),
	}

	StartReconcilers(ctx, t, mgr, logs, reconcilers...)

	rateLimitingkongPlugin := &configurationv1.KongPlugin{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "rate-limiting-kp-",
		},
		PluginName: "rate-limiting",
		Config: apiextensionsv1.JSON{
			Raw: []byte(`{"minute": 5, "policy": "local"}`),
		},
	}
	require.NoError(t, clientNamespaced.Create(ctx, rateLimitingkongPlugin))
	t.Logf("deployed %s KongPlugin (%s) resource", client.ObjectKeyFromObject(rateLimitingkongPlugin), rateLimitingkongPlugin.PluginName)

	t.Run("binding to KongService", func(t *testing.T) {
		serviceID := uuid.NewString()

		pluginID := uuid.NewString()
		sdk.PluginSDK.EXPECT().
			CreatePlugin(mock.Anything, cp.GetKonnectStatus().GetKonnectID(), mock.Anything).
			Return(
				&sdkkonnectops.CreatePluginResponse{
					Plugin: &sdkkonnectcomp.Plugin{
						ID: lo.ToPtr(pluginID),
					},
				},
				nil,
			)

		wKongPluginBinding := setupWatch[configurationv1alpha1.KongPluginBindingList](t, ctx, clientWithWatch, client.InNamespace(ns.Name))
		wKongPlugin := setupWatch[configurationv1.KongPluginList](t, ctx, clientWithWatch, client.InNamespace(ns.Name))

		kongService := deployKongServiceAttachedToCP(t, ctx, clientNamespaced, cp,
			WithAnnotation(consts.PluginsAnnotationKey, rateLimitingkongPlugin.Name),
		)
		t.Cleanup(func() {
			require.NoError(t, clientNamespaced.Delete(ctx, kongService))
		})
		updateKongServiceStatusWithProgrammed(t, ctx, clientNamespaced, kongService, serviceID, cp.GetKonnectStatus().GetKonnectID())

		t.Logf("waiting for KongPluginBinding to be created")
		kongPluginBinding := watchFor(t, ctx, wKongPluginBinding, watch.Added,
			func(kpb *configurationv1alpha1.KongPluginBinding) bool {
				return kpb.Spec.Targets.ServiceReference != nil &&
					kpb.Spec.Targets.ServiceReference.Name == kongService.Name &&
					kpb.Spec.PluginReference.Name == rateLimitingkongPlugin.Name
			},
			"KongPluginBinding wasn't created",
		)
		t.Logf(
			"checking that managed KongPlugin %s gets plugin-in-use finalizer added",
			client.ObjectKeyFromObject(rateLimitingkongPlugin),
		)
		_ = watchFor(t, ctx, wKongPlugin, watch.Modified,
			func(kp *configurationv1.KongPlugin) bool {
				return kp.Name == rateLimitingkongPlugin.Name &&
					controllerutil.ContainsFinalizer(kp, consts.PluginInUseFinalizer)
			},
			"KongPlugin wasn't updated to get plugin-in-use finalizer added",
		)

		t.Logf("delete managed kongPluginBinding %s, then check it gets recreated", client.ObjectKeyFromObject(kongPluginBinding))
		require.NoError(t, clientNamespaced.Delete(ctx, kongPluginBinding))
		kongPluginBinding = watchFor(t, ctx, wKongPluginBinding, watch.Added,
			func(kpb *configurationv1alpha1.KongPluginBinding) bool {
				return kpb.Spec.Targets.ServiceReference != nil &&
					kpb.Spec.Targets.ServiceReference.Name == kongService.Name &&
					kpb.Spec.PluginReference.Name == rateLimitingkongPlugin.Name
			},
			"KongPluginBinding wasn't recreated",
		)

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, sdk.PluginSDK.AssertExpectations(t))
		}, waitTime, tickTime)

		t.Logf(
			"remove annotation from KongService %s and check that managed KongPluginBinding %s gets deleted, "+
				"then check that gateway.konghq.com/plugin-in-use finalizer gets removed from %s KongPlugin",
			client.ObjectKeyFromObject(kongService),
			client.ObjectKeyFromObject(kongPluginBinding),
			client.ObjectKeyFromObject(rateLimitingkongPlugin),
		)

		sdk.PluginSDK.EXPECT().
			DeletePlugin(mock.Anything, cp.GetKonnectStatus().GetKonnectID(), mock.Anything).
			Return(
				&sdkkonnectops.DeletePluginResponse{
					StatusCode: 200,
				},
				nil,
			)

		t.Logf(
			"checking that managed KongPlugin %s gets plugin-in-use finalizer removed",
			client.ObjectKeyFromObject(rateLimitingkongPlugin),
		)
		wKongPlugin = setupWatch[configurationv1.KongPluginList](t, ctx, clientWithWatch, client.InNamespace(ns.Name))
		delete(kongService.Annotations, consts.PluginsAnnotationKey)
		require.NoError(t, clientNamespaced.Update(ctx, kongService))
		_ = watchFor(t, ctx, wKongPlugin, watch.Modified,
			func(kp *configurationv1.KongPlugin) bool {
				return kp.Name == rateLimitingkongPlugin.Name &&
					!controllerutil.ContainsFinalizer(kp, consts.PluginInUseFinalizer)
			},
			"KongPlugin wasn't updated to get plugin-in-use finalizer removed",
		)

		t.Logf(
			"checking that managed KongPluginBinding %s gets deleted",
			client.ObjectKeyFromObject(kongPluginBinding),
		)
		assert.EventuallyWithT(t,
			func(c *assert.CollectT) {
				assert.True(c, k8serrors.IsNotFound(
					clientNamespaced.Get(ctx, client.ObjectKeyFromObject(kongPluginBinding), kongPluginBinding),
				))
			}, waitTime, tickTime,
			"KongPluginBinding wasn't deleted after removing annotation from KongService",
		)

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, sdk.PluginSDK.AssertExpectations(t))
		}, waitTime, tickTime)
	})

	t.Run("binding to KongRoute", func(t *testing.T) {
		serviceID := uuid.NewString()
		routeID := uuid.NewString()

		wKongPluginBinding := setupWatch[configurationv1alpha1.KongPluginBindingList](t, ctx, clientWithWatch, client.InNamespace(ns.Name))
		wKongPlugin := setupWatch[configurationv1.KongPluginList](t, ctx, clientWithWatch, client.InNamespace(ns.Name))
		kongService := deployKongServiceAttachedToCP(t, ctx, clientNamespaced, cp)
		t.Cleanup(func() {
			require.NoError(t, clientNamespaced.Delete(ctx, kongService))
		})
		updateKongServiceStatusWithProgrammed(t, ctx, clientNamespaced, kongService, serviceID, cp.GetKonnectStatus().GetKonnectID())
		kongRoute := deployKongRouteAttachedToService(t, ctx, clientNamespaced, kongService,
			WithAnnotation(consts.PluginsAnnotationKey, rateLimitingkongPlugin.Name),
		)
		t.Cleanup(func() {
			require.NoError(t, clientNamespaced.Delete(ctx, kongRoute))
		})
		updateKongRouteStatusWithProgrammed(t, ctx, clientNamespaced, kongRoute, routeID, cp.GetKonnectStatus().GetKonnectID(), serviceID)

		t.Logf("waiting for KongPluginBinding to be created")
		kongPluginBinding := watchFor(t, ctx, wKongPluginBinding, watch.Added,
			func(kpb *configurationv1alpha1.KongPluginBinding) bool {
				return kpb.Spec.Targets.RouteReference != nil &&
					kpb.Spec.Targets.RouteReference.Name == kongRoute.Name &&
					kpb.Spec.PluginReference.Name == rateLimitingkongPlugin.Name
			},
			"KongPluginBinding wasn't created",
		)
		t.Logf(
			"checking that managed KongPlugin %s gets plugin-in-use finalizer added",
			client.ObjectKeyFromObject(rateLimitingkongPlugin),
		)
		_ = watchFor(t, ctx, wKongPlugin, watch.Modified,
			func(kp *configurationv1.KongPlugin) bool {
				return kp.Name == rateLimitingkongPlugin.Name &&
					controllerutil.ContainsFinalizer(kp, consts.PluginInUseFinalizer)
			},
			"KongPlugin wasn't updated to get plugin-in-use finalizer added",
		)

		t.Logf("delete managed kongPluginBinding %s, then check it gets recreated", client.ObjectKeyFromObject(kongPluginBinding))
		require.NoError(t, clientNamespaced.Delete(ctx, kongPluginBinding))
		kongPluginBinding = watchFor(t, ctx, wKongPluginBinding, watch.Added,
			func(kpb *configurationv1alpha1.KongPluginBinding) bool {
				return kpb.Spec.Targets.RouteReference != nil &&
					kpb.Spec.Targets.RouteReference.Name == kongRoute.Name &&
					kpb.Spec.PluginReference.Name == rateLimitingkongPlugin.Name
			},
			"KongPluginBinding wasn't recreated",
		)

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, sdk.PluginSDK.AssertExpectations(t))
		}, waitTime, tickTime)

		t.Logf(
			"remove annotation from KongRoute %s and check that managed KongPluginBinding %s gets deleted, "+
				"then check that gateway.konghq.com/plugin-in-use finalizer gets removed from %s KongPlugin",
			client.ObjectKeyFromObject(kongRoute),
			client.ObjectKeyFromObject(kongPluginBinding),
			client.ObjectKeyFromObject(rateLimitingkongPlugin),
		)

		sdk.PluginSDK.EXPECT().
			DeletePlugin(mock.Anything, cp.GetKonnectStatus().GetKonnectID(), mock.Anything).
			Return(
				&sdkkonnectops.DeletePluginResponse{
					StatusCode: 200,
				},
				nil,
			)

		t.Logf(
			"checking that managed KongPlugin %s gets plugin-in-use finalizer removed",
			client.ObjectKeyFromObject(rateLimitingkongPlugin),
		)
		wKongPlugin = setupWatch[configurationv1.KongPluginList](t, ctx, clientWithWatch, client.InNamespace(ns.Name))
		delete(kongRoute.Annotations, consts.PluginsAnnotationKey)
		require.NoError(t, clientNamespaced.Update(ctx, kongRoute))
		_ = watchFor(t, ctx, wKongPlugin, watch.Modified,
			func(kp *configurationv1.KongPlugin) bool {
				return kp.Name == rateLimitingkongPlugin.Name &&
					!controllerutil.ContainsFinalizer(kp, consts.PluginInUseFinalizer)
			},
			"KongPlugin wasn't updated to get plugin-in-use finalizer removed",
		)

		t.Logf(
			"checking that managed KongPluginBinding %s gets deleted",
			client.ObjectKeyFromObject(kongPluginBinding),
		)
		assert.EventuallyWithT(t,
			func(c *assert.CollectT) {
				assert.True(c, k8serrors.IsNotFound(
					clientNamespaced.Get(ctx, client.ObjectKeyFromObject(kongPluginBinding), kongPluginBinding),
				))
			}, waitTime, tickTime,
			"KongPluginBinding wasn't deleted after removing annotation from KongRoute",
		)

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, sdk.PluginSDK.AssertExpectations(t))
		}, waitTime, tickTime)
	})

	t.Run("binding to KongService and KongRoute", func(t *testing.T) {
		serviceID := uuid.NewString()
		routeID := uuid.NewString()

		wKongPluginBinding := setupWatch[configurationv1alpha1.KongPluginBindingList](t, ctx, clientWithWatch, client.InNamespace(ns.Name))
		wKongPlugin := setupWatch[configurationv1.KongPluginList](t, ctx, clientWithWatch, client.InNamespace(ns.Name))
		kongService := deployKongServiceAttachedToCP(t, ctx, clientNamespaced, cp,
			WithAnnotation(consts.PluginsAnnotationKey, rateLimitingkongPlugin.Name),
		)
		t.Cleanup(func() {
			require.NoError(t, clientNamespaced.Delete(ctx, kongService))
		})
		updateKongServiceStatusWithProgrammed(t, ctx, clientNamespaced, kongService, serviceID, cp.GetKonnectStatus().GetKonnectID())
		kongRoute := deployKongRouteAttachedToService(t, ctx, clientNamespaced, kongService,
			WithAnnotation(consts.PluginsAnnotationKey, rateLimitingkongPlugin.Name),
		)
		t.Cleanup(func() {
			require.NoError(t, clientNamespaced.Delete(ctx, kongRoute))
		})
		updateKongRouteStatusWithProgrammed(t, ctx, clientNamespaced, kongRoute, routeID, cp.GetKonnectStatus().GetKonnectID(), serviceID)

		t.Logf("waiting for KongPluginBinding to be created")
		kongPluginBinding := watchFor(t, ctx, wKongPluginBinding, watch.Added,
			func(kpb *configurationv1alpha1.KongPluginBinding) bool {
				return kpb.Spec.Targets.RouteReference != nil &&
					kpb.Spec.Targets.RouteReference.Name == kongRoute.Name &&
					kpb.Spec.Targets.ServiceReference != nil &&
					kpb.Spec.Targets.ServiceReference.Name == kongService.Name &&
					kpb.Spec.PluginReference.Name == rateLimitingkongPlugin.Name
			},
			"KongPluginBinding wasn't created",
		)
		t.Logf(
			"checking that managed KongPlugin %s gets plugin-in-use finalizer added",
			client.ObjectKeyFromObject(rateLimitingkongPlugin),
		)
		_ = watchFor(t, ctx, wKongPlugin, watch.Modified,
			func(kp *configurationv1.KongPlugin) bool {
				return kp.Name == rateLimitingkongPlugin.Name &&
					controllerutil.ContainsFinalizer(kp, consts.PluginInUseFinalizer)
			},
			"KongPlugin wasn't updated to get plugin-in-use finalizer added",
		)

		t.Logf("delete managed kongPluginBinding %s, then check it gets recreated", client.ObjectKeyFromObject(kongPluginBinding))
		require.NoError(t, clientNamespaced.Delete(ctx, kongPluginBinding))
		_ = watchFor(t, ctx, wKongPluginBinding, watch.Added,
			func(kpb *configurationv1alpha1.KongPluginBinding) bool {
				return kpb.Spec.Targets.RouteReference != nil &&
					kpb.Spec.Targets.RouteReference.Name == kongRoute.Name &&
					kpb.Spec.Targets.ServiceReference != nil &&
					kpb.Spec.Targets.ServiceReference.Name == kongService.Name &&
					kpb.Spec.PluginReference.Name == rateLimitingkongPlugin.Name
			},
			"KongPluginBinding wasn't recreated",
		)

		t.Logf(
			"remove annotation from KongRoute %s and check that there exists (is created) "+
				"a managed KongPluginBinding with only KongService in its targets",
			client.ObjectKeyFromObject(kongRoute),
		)

		delete(kongRoute.Annotations, consts.PluginsAnnotationKey)
		require.NoError(t, clientNamespaced.Update(ctx, kongRoute))
		kongPluginBinding = watchFor(t, ctx, wKongPluginBinding, watch.Added,
			func(kpb *configurationv1alpha1.KongPluginBinding) bool {
				return kpb.Spec.Targets.RouteReference == nil &&
					kpb.Spec.Targets.ServiceReference != nil &&
					kpb.Spec.Targets.ServiceReference.Name == kongService.Name &&
					kpb.Spec.PluginReference.Name == rateLimitingkongPlugin.Name
			},
			"KongPluginBinding wasn't created",
		)

		t.Logf(
			"remove annotation from KongService %s and check a managed KongPluginBinding (%s) gets deleted",
			client.ObjectKeyFromObject(kongService),
			client.ObjectKeyFromObject(kongPluginBinding),
		)
		sdk.PluginSDK.EXPECT().
			DeletePlugin(mock.Anything, cp.GetKonnectStatus().GetKonnectID(), mock.Anything).
			Return(
				&sdkkonnectops.DeletePluginResponse{
					StatusCode: 200,
				},
				nil,
			)
		delete(kongService.Annotations, consts.PluginsAnnotationKey)
		require.NoError(t, clientNamespaced.Update(ctx, kongService))

		assert.EventuallyWithT(t,
			func(c *assert.CollectT) {
				assert.True(c, k8serrors.IsNotFound(
					clientNamespaced.Get(ctx, client.ObjectKeyFromObject(kongPluginBinding), kongPluginBinding),
				))
			}, waitTime, tickTime,
			"KongPluginBinding wasn't deleted after removing annotation from KongRoute",
		)

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, sdk.PluginSDK.AssertExpectations(t))
		}, waitTime, tickTime)
	})
}
