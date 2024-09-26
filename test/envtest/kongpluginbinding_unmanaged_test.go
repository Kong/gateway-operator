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
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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

func TestKongPluginBindingUnmanaged(t *testing.T) {
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

	t.Run("binding to KongService", func(t *testing.T) {
		proxyCacheKongPlugin := deployProxyCachePlugin(t, ctx, clientNamespaced)

		serviceID := uuid.NewString()
		pluginID := uuid.NewString()

		createCall := sdk.PluginSDK.EXPECT().
			CreatePlugin(mock.Anything, cp.GetKonnectStatus().GetKonnectID(), mock.Anything).
			Return(
				&sdkkonnectops.CreatePluginResponse{
					Plugin: &sdkkonnectcomp.Plugin{
						ID: lo.ToPtr(pluginID),
					},
				},
				nil,
			)
		defer createCall.Unset()

		kongService := deployKongServiceAttachedToCP(t, ctx, clientNamespaced, cp)
		t.Cleanup(func() {
			require.NoError(t, client.IgnoreNotFound(clientNamespaced.Delete(ctx, kongService)))
		})
		updateKongServiceStatusWithProgrammed(t, ctx, clientNamespaced, kongService, serviceID, cp.GetKonnectStatus().GetKonnectID())

		wKongPlugin := setupWatch[configurationv1.KongPluginList](t, ctx, clientWithWatch, client.InNamespace(ns.Name))
		kpb := deployKongPluginBinding(t, ctx, clientNamespaced,
			&configurationv1alpha1.KongPluginBinding{
				Spec: configurationv1alpha1.KongPluginBindingSpec{
					ControlPlaneRef: &configurationv1alpha1.ControlPlaneRef{
						Type: configurationv1alpha1.ControlPlaneRefKonnectNamespacedRef,
						KonnectNamespacedRef: &configurationv1alpha1.KonnectNamespacedRef{
							Name: cp.Name,
						},
					},
					PluginReference: configurationv1alpha1.PluginRef{
						Name: proxyCacheKongPlugin.Name,
					},
					Targets: configurationv1alpha1.KongPluginBindingTargets{
						ServiceReference: &configurationv1alpha1.TargetRefWithGroupKind{
							Group: configurationv1alpha1.GroupVersion.Group,
							Kind:  "KongService",
							Name:  kongService.Name,
						},
					},
				},
			},
		)
		t.Logf(
			"wait for the controller to pick the new unmanaged KongPluginBinding %s and put a %s finalizer on the referenced plugin %s",
			client.ObjectKeyFromObject(kpb),
			consts.PluginInUseFinalizer,
			client.ObjectKeyFromObject(proxyCacheKongPlugin),
		)
		_ = watchFor(t, ctx, wKongPlugin, watch.Modified,
			func(kp *configurationv1.KongPlugin) bool {
				return kp.Name == proxyCacheKongPlugin.Name &&
					controllerutil.ContainsFinalizer(kp, consts.PluginInUseFinalizer)
			},
			"KongPlugin wasn't updated to get the plugin-in-use finalizer",
		)
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, sdk.PluginSDK.AssertExpectations(t))
		}, waitTime, tickTime)

		sdk.PluginSDK.EXPECT().
			DeletePlugin(mock.Anything, cp.GetKonnectStatus().GetKonnectID(), mock.Anything).
			Return(
				&sdkkonnectops.DeletePluginResponse{
					StatusCode: 200,
				},
				nil,
			)

		t.Logf("delete the KongPlugin %s, then check it does not get collected", client.ObjectKeyFromObject(proxyCacheKongPlugin))
		require.NoError(t, clientNamespaced.Delete(ctx, proxyCacheKongPlugin))
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.False(c, k8serrors.IsNotFound(
				clientNamespaced.Get(ctx, client.ObjectKeyFromObject(proxyCacheKongPlugin), proxyCacheKongPlugin),
			))
			assert.True(c, proxyCacheKongPlugin.DeletionTimestamp != nil)
			assert.True(c, controllerutil.ContainsFinalizer(proxyCacheKongPlugin, consts.PluginInUseFinalizer))
		}, waitTime, tickTime)

		t.Logf("delete the unmanaged KongPluginBinding %s, then check the proxy-cache KongPlugin %s gets collected",
			client.ObjectKeyFromObject(kpb),
			client.ObjectKeyFromObject(proxyCacheKongPlugin),
		)
		require.NoError(t, clientNamespaced.Delete(ctx, kpb))
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, k8serrors.IsNotFound(
				clientNamespaced.Get(ctx, client.ObjectKeyFromObject(proxyCacheKongPlugin), proxyCacheKongPlugin),
			))
		}, waitTime, tickTime, "KongPlugin did not got deleted but shouldn't have")

		t.Logf(
			"delete the KongService %s and check it gets collected, as the KongPluginBinding finalizer should have been removed",
			client.ObjectKeyFromObject(kongService),
		)
		require.NoError(t, clientNamespaced.Delete(ctx, kongService))
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, k8serrors.IsNotFound(
				clientNamespaced.Get(ctx, client.ObjectKeyFromObject(kongService), kongService),
			))
		}, waitTime, tickTime)

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, sdk.PluginSDK.AssertExpectations(t))
		}, waitTime, tickTime)
	})
	t.Run("binding to KongRoute", func(t *testing.T) {
		proxyCacheKongPlugin := deployProxyCachePlugin(t, ctx, clientNamespaced)

		serviceID := uuid.NewString()
		routeID := uuid.NewString()
		pluginID := uuid.NewString()

		createCall := sdk.PluginSDK.EXPECT().
			CreatePlugin(mock.Anything, cp.GetKonnectStatus().GetKonnectID(), mock.Anything).
			Return(
				&sdkkonnectops.CreatePluginResponse{
					Plugin: &sdkkonnectcomp.Plugin{
						ID: lo.ToPtr(pluginID),
					},
				},
				nil,
			)
		defer createCall.Unset()

		kongService := deployKongServiceAttachedToCP(t, ctx, clientNamespaced, cp)
		t.Cleanup(func() {
			require.NoError(t, client.IgnoreNotFound(clientNamespaced.Delete(ctx, kongService)))
		})
		updateKongServiceStatusWithProgrammed(t, ctx, clientNamespaced, kongService, serviceID, cp.GetKonnectStatus().GetKonnectID())
		kongRoute := deployKongRouteAttachedToService(t, ctx, clientNamespaced, kongService)
		t.Cleanup(func() {
			require.NoError(t, client.IgnoreNotFound(clientNamespaced.Delete(ctx, kongRoute)))
		})
		updateKongRouteStatusWithProgrammed(t, ctx, clientNamespaced, kongRoute, routeID, cp.GetKonnectStatus().GetKonnectID(), serviceID)

		wKongPlugin := setupWatch[configurationv1.KongPluginList](t, ctx, clientWithWatch, client.InNamespace(ns.Name))
		kpb := deployKongPluginBinding(t, ctx, clientNamespaced,
			&configurationv1alpha1.KongPluginBinding{
				Spec: configurationv1alpha1.KongPluginBindingSpec{
					ControlPlaneRef: &configurationv1alpha1.ControlPlaneRef{
						Type: configurationv1alpha1.ControlPlaneRefKonnectNamespacedRef,
						KonnectNamespacedRef: &configurationv1alpha1.KonnectNamespacedRef{
							Name: cp.Name,
						},
					},
					PluginReference: configurationv1alpha1.PluginRef{
						Name: proxyCacheKongPlugin.Name,
					},
					Targets: configurationv1alpha1.KongPluginBindingTargets{
						RouteReference: &configurationv1alpha1.TargetRefWithGroupKind{
							Group: configurationv1alpha1.GroupVersion.Group,
							Kind:  "KongRoute",
							Name:  kongRoute.Name,
						},
					},
				},
			},
		)
		t.Logf(
			"wait for the controller to pick the new unmanaged KongPluginBinding %s and put a %s finalizer on the referenced plugin %s",
			client.ObjectKeyFromObject(kpb),
			consts.PluginInUseFinalizer,
			client.ObjectKeyFromObject(proxyCacheKongPlugin),
		)
		_ = watchFor(t, ctx, wKongPlugin, watch.Modified,
			func(kp *configurationv1.KongPlugin) bool {
				return kp.Name == proxyCacheKongPlugin.Name &&
					controllerutil.ContainsFinalizer(kp, consts.PluginInUseFinalizer)
			},
			"KongPlugin wasn't updated to get the plugin-in-use finalizer",
		)
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, sdk.PluginSDK.AssertExpectations(t))
		}, waitTime, tickTime)

		sdk.PluginSDK.EXPECT().
			DeletePlugin(mock.Anything, cp.GetKonnectStatus().GetKonnectID(), mock.Anything).
			Return(
				&sdkkonnectops.DeletePluginResponse{
					StatusCode: 200,
				},
				nil,
			)

		t.Logf("delete the KongPlugin %s, then check it does not get collected", client.ObjectKeyFromObject(proxyCacheKongPlugin))
		require.NoError(t, clientNamespaced.Delete(ctx, proxyCacheKongPlugin))
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.False(c, k8serrors.IsNotFound(
				clientNamespaced.Get(ctx, client.ObjectKeyFromObject(proxyCacheKongPlugin), proxyCacheKongPlugin),
			))
			assert.True(c, proxyCacheKongPlugin.DeletionTimestamp != nil)
			assert.True(c, controllerutil.ContainsFinalizer(proxyCacheKongPlugin, consts.PluginInUseFinalizer))
		}, waitTime, tickTime)

		t.Logf("delete the unmanaged KongPluginBinding %s, then check the proxy-cache KongPlugin %s gets collected",
			client.ObjectKeyFromObject(kpb),
			client.ObjectKeyFromObject(proxyCacheKongPlugin),
		)
		require.NoError(t, clientNamespaced.Delete(ctx, kpb))
		_ = watchFor(t, ctx, wKongPlugin, watch.Deleted,
			func(kp *configurationv1.KongPlugin) bool {
				return kp.Name == proxyCacheKongPlugin.Name
			},
			"KongPlugin did not got deleted but shouldn't have",
		)

		t.Logf(
			"delete the KongRoute %s and check it gets collected, as the KongPluginBinding finalizer should have been removed",
			client.ObjectKeyFromObject(kongRoute),
		)
		require.NoError(t, clientNamespaced.Delete(ctx, kongRoute))
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, k8serrors.IsNotFound(
				clientNamespaced.Get(ctx, client.ObjectKeyFromObject(kongRoute), kongRoute),
			))
		}, waitTime, tickTime)

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, sdk.PluginSDK.AssertExpectations(t))
		}, waitTime, tickTime)
	})

	t.Run("binding to KongService and KongRoute", func(t *testing.T) {
		proxyCacheKongPlugin := deployProxyCachePlugin(t, ctx, clientNamespaced)

		serviceID := uuid.NewString()
		routeID := uuid.NewString()
		pluginID := uuid.NewString()

		kongService := deployKongServiceAttachedToCP(t, ctx, clientNamespaced, cp)
		t.Cleanup(func() {
			require.NoError(t, client.IgnoreNotFound(clientNamespaced.Delete(ctx, kongService)))
		})
		updateKongServiceStatusWithProgrammed(t, ctx, clientNamespaced, kongService, serviceID, cp.GetKonnectStatus().GetKonnectID())
		kongRoute := deployKongRouteAttachedToService(t, ctx, clientNamespaced, kongService)
		t.Cleanup(func() {
			require.NoError(t, client.IgnoreNotFound(clientNamespaced.Delete(ctx, kongRoute)))
		})
		updateKongRouteStatusWithProgrammed(t, ctx, clientNamespaced, kongRoute, routeID, cp.GetKonnectStatus().GetKonnectID(), serviceID)

		wKongPlugin := setupWatch[configurationv1.KongPluginList](t, ctx, clientWithWatch, client.InNamespace(ns.Name))
		sdk.PluginSDK.EXPECT().
			CreatePlugin(
				mock.Anything,
				cp.GetKonnectStatus().GetKonnectID(),
				mock.MatchedBy(func(pi sdkkonnectcomp.PluginInput) bool {
					return pi.Route != nil && pi.Route.ID != nil && *pi.Route.ID == routeID &&
						pi.Service != nil && pi.Service.ID != nil && *pi.Service.ID == serviceID
				})).
			Return(
				&sdkkonnectops.CreatePluginResponse{
					Plugin: &sdkkonnectcomp.Plugin{
						ID: lo.ToPtr(pluginID),
					},
				},
				nil,
			)
		kpb := deployKongPluginBinding(t, ctx, clientNamespaced,
			&configurationv1alpha1.KongPluginBinding{
				Spec: configurationv1alpha1.KongPluginBindingSpec{
					ControlPlaneRef: &configurationv1alpha1.ControlPlaneRef{
						Type: configurationv1alpha1.ControlPlaneRefKonnectNamespacedRef,
						KonnectNamespacedRef: &configurationv1alpha1.KonnectNamespacedRef{
							Name: cp.Name,
						},
					},
					PluginReference: configurationv1alpha1.PluginRef{
						Name: proxyCacheKongPlugin.Name,
					},
					Targets: configurationv1alpha1.KongPluginBindingTargets{
						RouteReference: &configurationv1alpha1.TargetRefWithGroupKind{
							Group: configurationv1alpha1.GroupVersion.Group,
							Kind:  "KongRoute",
							Name:  kongRoute.Name,
						},
						ServiceReference: &configurationv1alpha1.TargetRefWithGroupKind{
							Group: configurationv1alpha1.GroupVersion.Group,
							Kind:  "KongService",
							Name:  kongService.Name,
						},
					},
				},
			},
		)
		t.Logf(
			"wait for the controller to pick the new unmanaged KongPluginBinding %s and put a %s finalizer on the referenced plugin %s",
			client.ObjectKeyFromObject(kpb),
			consts.PluginInUseFinalizer,
			client.ObjectKeyFromObject(proxyCacheKongPlugin),
		)
		_ = watchFor(t, ctx, wKongPlugin, watch.Modified,
			func(kp *configurationv1.KongPlugin) bool {
				return kp.Name == proxyCacheKongPlugin.Name &&
					controllerutil.ContainsFinalizer(kp, consts.PluginInUseFinalizer)
			},
			"KongPlugin wasn't updated to get the plugin-in-use finalizer",
		)
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, sdk.PluginSDK.AssertExpectations(t))
		}, waitTime, tickTime)

		sdk.PluginSDK.EXPECT().
			DeletePlugin(mock.Anything, cp.GetKonnectStatus().GetKonnectID(), mock.Anything).
			Return(
				&sdkkonnectops.DeletePluginResponse{
					StatusCode: 200,
				},
				nil,
			)

		t.Logf("delete the KongPlugin %s, then check it does not get collected", client.ObjectKeyFromObject(proxyCacheKongPlugin))
		require.NoError(t, clientNamespaced.Delete(ctx, proxyCacheKongPlugin))
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.False(c, k8serrors.IsNotFound(
				clientNamespaced.Get(ctx, client.ObjectKeyFromObject(proxyCacheKongPlugin), proxyCacheKongPlugin),
			))
			assert.True(c, proxyCacheKongPlugin.DeletionTimestamp != nil)
			assert.True(c, controllerutil.ContainsFinalizer(proxyCacheKongPlugin, consts.PluginInUseFinalizer))
		}, waitTime, tickTime)

		t.Logf("delete the unmanaged KongPluginBinding %s, then check the proxy-cache KongPlugin %s gets collected",
			client.ObjectKeyFromObject(kpb),
			client.ObjectKeyFromObject(proxyCacheKongPlugin),
		)
		require.NoError(t, clientNamespaced.Delete(ctx, kpb))
		_ = watchFor(t, ctx, wKongPlugin, watch.Deleted,
			func(kp *configurationv1.KongPlugin) bool {
				return kp.Name == proxyCacheKongPlugin.Name
			},
			"KongPlugin did not got deleted but shouldn't have",
		)

		t.Logf(
			"delete the KongRoute %s and check it gets collected, as the KongPluginBinding finalizer should have been removed",
			client.ObjectKeyFromObject(kongRoute),
		)
		require.NoError(t, clientNamespaced.Delete(ctx, kongRoute))
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, k8serrors.IsNotFound(
				clientNamespaced.Get(ctx, client.ObjectKeyFromObject(kongRoute), kongRoute),
			))
		}, waitTime, tickTime)

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, sdk.PluginSDK.AssertExpectations(t))
		}, waitTime, tickTime)
	})
}
