package envtest

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectops "github.com/Kong/sdk-konnect-go/models/operations"
	sdkkonnecterrs "github.com/Kong/sdk-konnect-go/models/sdkerrors"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kong/gateway-operator/controller/konnect"
	sdkmocks "github.com/kong/gateway-operator/controller/konnect/ops/sdk/mocks"
	"github.com/kong/gateway-operator/modules/manager/scheme"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
	"github.com/kong/gateway-operator/test/helpers/deploy"

	commonv1alpha1 "github.com/kong/kubernetes-configuration/api/common/v1alpha1"
	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	configurationv1beta1 "github.com/kong/kubernetes-configuration/api/configuration/v1beta1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

func TestKongConsumer(t *testing.T) {
	t.Parallel()
	ctx, cancel := Context(t, context.Background())
	defer cancel()
	cfg, ns := Setup(t, ctx, scheme.Get())

	t.Log("Setting up the manager with reconcilers")
	mgr, logs := NewManager(t, ctx, cfg, scheme.Get(), WithKonnectCacheIndices(ctx))
	factory := sdkmocks.NewMockSDKFactory(t)
	sdk := factory.SDK
	reconcilers := []Reconciler{
		konnect.NewKonnectEntityReconciler(factory, false, mgr.GetClient(),
			konnect.WithKonnectEntitySyncPeriod[configurationv1.KongConsumer](konnectInfiniteSyncTime),
		),
		konnect.NewKonnectEntityReconciler(factory, false, mgr.GetClient(),
			konnect.WithKonnectEntitySyncPeriod[configurationv1beta1.KongConsumerGroup](konnectInfiniteSyncTime),
		),
		konnect.NewKongCredentialSecretReconciler(false, mgr.GetClient(), mgr.GetScheme()),
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

	t.Log("Setting up a watch for KongConsumer events")
	cWatch := setupWatch[configurationv1.KongConsumerList](t, ctx, cl, client.InNamespace(ns.Name))

	t.Run("should create, update and delete Consumer without ConsumerGroups successfully", func(t *testing.T) {
		const (
			consumerID      = "consumer-id"
			username        = "user-1"
			updatedUsername = "user-1-updated"
		)
		t.Log("Setting up SDK expectations on KongConsumer creation")
		sdk.ConsumersSDK.EXPECT().
			CreateConsumer(mock.Anything, cp.GetKonnectStatus().GetKonnectID(),
				mock.MatchedBy(func(input sdkkonnectcomp.ConsumerInput) bool {
					return input.Username != nil && *input.Username == username
				}),
			).Return(&sdkkonnectops.CreateConsumerResponse{
			Consumer: &sdkkonnectcomp.Consumer{
				ID: lo.ToPtr(consumerID),
			},
		}, nil)

		t.Log("Setting up SDK expectation on possibly updating KongConsumer ( due to asynchronous nature of updates between KongConsumer and KongConsumerGroup)")
		sdk.ConsumersSDK.EXPECT().
			UpsertConsumer(mock.Anything, mock.MatchedBy(func(r sdkkonnectops.UpsertConsumerRequest) bool {
				return r.ConsumerID == consumerID
			})).
			Return(&sdkkonnectops.UpsertConsumerResponse{}, nil).
			Maybe()

		t.Log("Setting up SDK expectation on KongConsumerGroups listing")
		sdk.ConsumerGroupSDK.EXPECT().
			ListConsumerGroupsForConsumer(mock.Anything, sdkkonnectops.ListConsumerGroupsForConsumerRequest{
				ConsumerID:     consumerID,
				ControlPlaneID: cp.GetKonnectStatus().GetKonnectID(),
			}).Return(&sdkkonnectops.ListConsumerGroupsForConsumerResponse{}, nil)

		t.Log("Creating KongConsumer")
		createdConsumer := deploy.KongConsumer(t, ctx, clientNamespaced, username,
			deploy.WithKonnectNamespacedRefControlPlaneRef(cp),
		)

		t.Log("Waiting for KongConsumer to be programmed")
		watchFor(t, ctx, cWatch, watch.Modified, func(c *configurationv1.KongConsumer) bool {
			if c.GetName() != createdConsumer.GetName() {
				return false
			}
			return lo.ContainsBy(c.Status.Conditions, func(condition metav1.Condition) bool {
				return condition.Type == konnectv1alpha1.KonnectEntityProgrammedConditionType &&
					condition.Status == metav1.ConditionTrue
			})
		}, "KongConsumer's Programmed condition should be true eventually")

		t.Log("Waiting for KongConsumer to be created in the SDK")
		require.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, factory.SDK.ConsumersSDK.AssertExpectations(t))
		}, waitTime, tickTime)

		t.Log("Setting up SDK expectations on KongConsumer update")
		sdk.ConsumersSDK.EXPECT().
			UpsertConsumer(mock.Anything, mock.MatchedBy(func(r sdkkonnectops.UpsertConsumerRequest) bool {
				match := r.ConsumerID == consumerID &&
					r.Consumer.Username != nil && *r.Consumer.Username == updatedUsername
				return match
			})).
			Return(&sdkkonnectops.UpsertConsumerResponse{}, nil)

		t.Log("Patching KongConsumer")
		consumerToPatch := createdConsumer.DeepCopy()
		consumerToPatch.Username = updatedUsername
		require.NoError(t, clientNamespaced.Patch(ctx, consumerToPatch, client.MergeFrom(createdConsumer)))

		t.Log("Waiting for KongConsumer to be updated in the SDK")
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, factory.SDK.ConsumersSDK.AssertExpectations(t))
		}, waitTime, tickTime)

		t.Log("Setting up SDK expectations on KongConsumer deletion")
		sdk.ConsumersSDK.EXPECT().
			DeleteConsumer(mock.Anything, cp.GetKonnectStatus().GetKonnectID(), consumerID).
			Return(&sdkkonnectops.DeleteConsumerResponse{}, nil)

		t.Log("Deleting KongConsumer")
		require.NoError(t, cl.Delete(ctx, createdConsumer))

		require.EventuallyWithT(t,
			func(c *assert.CollectT) {
				assert.True(c, k8serrors.IsNotFound(
					clientNamespaced.Get(ctx, client.ObjectKeyFromObject(createdConsumer), createdConsumer),
				))
			}, waitTime, tickTime,
		)

		t.Log("Waiting for KongConsumer to be deleted in the SDK")
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, factory.SDK.ConsumersSDK.AssertExpectations(t))
		}, waitTime, tickTime)
	})

	t.Log("Setting up a watch for KongConsumerGroup events")
	cgWatch := setupWatch[configurationv1beta1.KongConsumerGroupList](t, ctx, cl, client.InNamespace(ns.Name))

	t.Run("should create, update and delete Consumer with ConsumerGroups successfully", func(t *testing.T) {
		const (
			consumerID = "consumer-id"
			username   = "user-2"

			cgID              = "consumer-group-id"
			consumerGroupName = "consumer-group-1"
		)
		t.Log("Setting up SDK expectations on KongConsumer creation with ConsumerGroup")
		sdk.ConsumersSDK.EXPECT().
			CreateConsumer(mock.Anything, cp.GetKonnectStatus().GetKonnectID(),
				mock.MatchedBy(func(input sdkkonnectcomp.ConsumerInput) bool {
					return input.Username != nil && *input.Username == username
				}),
			).Return(&sdkkonnectops.CreateConsumerResponse{
			Consumer: &sdkkonnectcomp.Consumer{
				ID: lo.ToPtr(consumerID),
			},
		}, nil)

		t.Log("Setting up SDK expectation on possibly updating KongConsumer (due to asynchronous nature of updates between KongConsumer and KongConsumerGroup)")
		sdk.ConsumersSDK.EXPECT().
			UpsertConsumer(mock.Anything, mock.MatchedBy(func(r sdkkonnectops.UpsertConsumerRequest) bool {
				return r.ConsumerID == consumerID
			})).
			Return(&sdkkonnectops.UpsertConsumerResponse{}, nil).
			Maybe()

		sdk.ConsumerGroupSDK.EXPECT().
			CreateConsumerGroup(mock.Anything, cp.GetKonnectStatus().GetKonnectID(),
				mock.MatchedBy(func(input sdkkonnectcomp.ConsumerGroupInput) bool {
					return input.Name == consumerGroupName
				}),
			).Return(&sdkkonnectops.CreateConsumerGroupResponse{
			ConsumerGroup: &sdkkonnectcomp.ConsumerGroup{
				ID: lo.ToPtr(cgID),
			},
		}, nil)

		t.Log("Setting up SDK expectation on KongConsumerGroups listing")
		emptyListCall := sdk.ConsumerGroupSDK.EXPECT().
			ListConsumerGroupsForConsumer(mock.Anything, sdkkonnectops.ListConsumerGroupsForConsumerRequest{
				ConsumerID:     consumerID,
				ControlPlaneID: cp.GetKonnectStatus().GetKonnectID(),
			}).Return(&sdkkonnectops.ListConsumerGroupsForConsumerResponse{
			// Returning no ConsumerGroups associated with the Consumer in Konnect to trigger addition.
		}, nil)

		t.Log("Setting up SDK expectation on adding Consumer to ConsumerGroup")
		sdk.ConsumerGroupSDK.EXPECT().
			AddConsumerToGroup(mock.Anything, sdkkonnectops.AddConsumerToGroupRequest{
				ConsumerGroupID: cgID,
				ControlPlaneID:  cp.GetKonnectStatus().GetKonnectID(),
				RequestBody: &sdkkonnectops.AddConsumerToGroupRequestBody{
					ConsumerID: lo.ToPtr(consumerID),
				},
			}).Return(&sdkkonnectops.AddConsumerToGroupResponse{}, nil)

		t.Log("Creating KongConsumerGroup")
		createdConsumerGroup := deploy.KongConsumerGroupAttachedToCP(t, ctx, clientNamespaced,
			deploy.WithKonnectNamespacedRefControlPlaneRef(cp),
			func(obj client.Object) {
				cg := obj.(*configurationv1beta1.KongConsumerGroup)
				cg.Spec.Name = consumerGroupName
			},
		)

		t.Log("Creating KongConsumer and patching it with ConsumerGroup")
		createdConsumer := deploy.KongConsumer(t, ctx, clientNamespaced, username,
			deploy.WithKonnectNamespacedRefControlPlaneRef(cp),
		)
		consumer := createdConsumer.DeepCopy()
		consumer.ConsumerGroups = []string{createdConsumerGroup.GetName()}
		require.NoError(t, clientNamespaced.Patch(ctx, consumer, client.MergeFrom(createdConsumer)))

		t.Log("Waiting for KongConsumer to be programmed")
		watchFor(t, ctx, cWatch, watch.Modified, func(c *configurationv1.KongConsumer) bool {
			if c.GetName() != createdConsumer.GetName() {
				return false
			}
			return lo.ContainsBy(c.Status.Conditions, func(condition metav1.Condition) bool {
				return condition.Type == konnectv1alpha1.KonnectEntityProgrammedConditionType &&
					condition.Status == metav1.ConditionTrue
			})
		}, "KongConsumer's Programmed condition should be true eventually")

		t.Log("Waiting for KongConsumerGroup to be programmed")
		watchFor(t, ctx, cgWatch, watch.Modified, func(c *configurationv1beta1.KongConsumerGroup) bool {
			if c.GetName() != createdConsumerGroup.GetName() {
				return false
			}
			return lo.ContainsBy(c.Status.Conditions, func(condition metav1.Condition) bool {
				return condition.Type == konnectv1alpha1.KonnectEntityProgrammedConditionType &&
					condition.Status == metav1.ConditionTrue
			})
		}, "KongConsumerGroup's Programmed condition should be true eventually")

		t.Log("Waiting for SDK expectations to be met")
		require.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, factory.SDK.ConsumersSDK.AssertExpectations(t))
		}, waitTime, tickTime)

		t.Log("Setting up SDK expectations on KongConsumer update with ConsumerGroup")
		sdk.ConsumersSDK.EXPECT().
			UpsertConsumer(mock.Anything, mock.MatchedBy(func(r sdkkonnectops.UpsertConsumerRequest) bool {
				return r.ConsumerID == consumerID &&
					r.Consumer.Username != nil && *r.Consumer.Username == "user-2-updated"
			})).
			Return(&sdkkonnectops.UpsertConsumerResponse{}, nil)

		emptyListCall.Unset() // Unset the previous expectation to allow the new one to be set.
		sdk.ConsumerGroupSDK.EXPECT().
			ListConsumerGroupsForConsumer(mock.Anything, sdkkonnectops.ListConsumerGroupsForConsumerRequest{
				ConsumerID:     consumerID,
				ControlPlaneID: cp.GetKonnectStatus().GetKonnectID(),
			}).Return(&sdkkonnectops.ListConsumerGroupsForConsumerResponse{
			Object: &sdkkonnectops.ListConsumerGroupsForConsumerResponseBody{
				Data: []sdkkonnectcomp.ConsumerGroup{
					{
						// Returning an ID that we haven't defined to be associated with the Consumer.
						// Should trigger removal.
						ID: lo.ToPtr("not-defined-in-crd"),
					},
					{
						// Returning the ID of the ConsumerGroup we have defined to be associated with the Consumer.
						// Should not trigger any action as it's already associated.
						ID: lo.ToPtr(cgID),
					},
				},
			},
		}, nil)

		sdk.ConsumerGroupSDK.EXPECT().
			RemoveConsumerFromGroup(mock.Anything, sdkkonnectops.RemoveConsumerFromGroupRequest{
				ConsumerGroupID: "not-defined-in-crd",
				ControlPlaneID:  cp.GetKonnectStatus().GetKonnectID(),
				ConsumerID:      consumerID,
			}).Return(&sdkkonnectops.RemoveConsumerFromGroupResponse{}, nil)

		t.Log("Patching KongConsumer to trigger reconciliation")
		consumerToPatch := createdConsumer.DeepCopy()
		consumerToPatch.Username = "user-2-updated"
		require.NoError(t, clientNamespaced.Patch(ctx, consumerToPatch, client.MergeFrom(createdConsumer)))

		t.Log("Waiting for SDK expectations to be met")
		require.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, factory.SDK.ConsumerGroupSDK.AssertExpectations(t))
		}, waitTime, tickTime)
	})

	t.Run("should handle conflict in creation correctly", func(t *testing.T) {
		const (
			consumerID = "consumer-id-conflict"
			username   = "user-3"
		)
		t.Log("Setup mock SDK for creating consumer and listing consumers by UID")
		cpID := cp.GetKonnectStatus().GetKonnectID()
		sdk.ConsumersSDK.EXPECT().CreateConsumer(
			mock.Anything,
			cpID,
			mock.MatchedBy(func(input sdkkonnectcomp.ConsumerInput) bool {
				return input.Username != nil && *input.Username == username
			}),
		).Return(&sdkkonnectops.CreateConsumerResponse{}, &sdkkonnecterrs.ConflictError{})

		sdk.ConsumersSDK.EXPECT().ListConsumer(
			mock.Anything,
			mock.MatchedBy(func(req sdkkonnectops.ListConsumerRequest) bool {
				return req.ControlPlaneID == cpID &&
					req.Tags != nil && strings.HasPrefix(*req.Tags, "k8s-uid")
			}),
		).Return(&sdkkonnectops.ListConsumerResponse{
			Object: &sdkkonnectops.ListConsumerResponseBody{
				Data: []sdkkonnectcomp.Consumer{
					{
						ID: lo.ToPtr(consumerID),
					},
				},
			},
		}, nil)

		t.Log("Creating a KongConsumer")
		deploy.KongConsumer(t, ctx, clientNamespaced, username,
			deploy.WithKonnectNamespacedRefControlPlaneRef(cp),
		)

		t.Log("Watching for KongConsumers to verify the created KongConsumer programmed")
		watchFor(t, ctx, cWatch, watch.Modified, func(c *configurationv1.KongConsumer) bool {
			return c.GetKonnectID() == consumerID && k8sutils.IsProgrammed(c)
		}, "KongConsumer should be programmed and have ID in status after handling conflict")

		t.Log("Ensuring that the SDK's create and list methods are called")
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, factory.SDK.ConsumersSDK.AssertExpectations(t))
		}, waitTime, tickTime)
	})

	t.Run("should handle konnectID control plane reference", func(t *testing.T) {
		const (
			consumerID = "consumer-with-cp-konnect-id"
			username   = "user-with-cp-konnect-id"
		)
		t.Log("Setting up SDK expectations on KongConsumer creation")
		sdk.ConsumersSDK.EXPECT().
			CreateConsumer(mock.Anything, cp.GetKonnectStatus().GetKonnectID(),
				mock.MatchedBy(func(input sdkkonnectcomp.ConsumerInput) bool {
					return input.Username != nil && *input.Username == username
				}),
			).Return(&sdkkonnectops.CreateConsumerResponse{
			Consumer: &sdkkonnectcomp.Consumer{
				ID: lo.ToPtr(consumerID),
			},
		}, nil)

		t.Log("Setting up SDK expectation on possibly updating KongConsumer ( due to asynchronous nature of updates between KongConsumer and KongConsumerGroup)")
		sdk.ConsumersSDK.EXPECT().
			UpsertConsumer(mock.Anything, mock.MatchedBy(func(r sdkkonnectops.UpsertConsumerRequest) bool {
				return r.ConsumerID == consumerID
			})).
			Return(&sdkkonnectops.UpsertConsumerResponse{}, nil).
			Maybe()

		t.Log("Setting up SDK expectation on KongConsumerGroups listing")
		sdk.ConsumerGroupSDK.EXPECT().
			ListConsumerGroupsForConsumer(mock.Anything, sdkkonnectops.ListConsumerGroupsForConsumerRequest{
				ConsumerID:     consumerID,
				ControlPlaneID: cp.GetKonnectStatus().GetKonnectID(),
			}).Return(&sdkkonnectops.ListConsumerGroupsForConsumerResponse{}, nil)

		t.Log("Creating KongConsumer with ControlPlaneRef type=konnectID")
		createdConsumer := deploy.KongConsumer(t, ctx, clientNamespaced, username,
			deploy.WithKonnectNamespacedRefControlPlaneRef(cp),
			deploy.WithKonnectIDControlPlaneRef(cp),
		)

		t.Log("Waiting for KongConsumer to be programmed")
		watchFor(t, ctx, cWatch, watch.Modified, func(c *configurationv1.KongConsumer) bool {
			if c.GetName() != createdConsumer.GetName() {
				return false
			}
			if c.GetControlPlaneRef().Type != commonv1alpha1.ControlPlaneRefKonnectID {
				return false
			}
			return lo.ContainsBy(c.Status.Conditions, func(condition metav1.Condition) bool {
				return condition.Type == konnectv1alpha1.KonnectEntityProgrammedConditionType &&
					condition.Status == metav1.ConditionTrue
			})
		}, "KongConsumer's Programmed condition should be true eventually")

		t.Log("Waiting for KongConsumer to be created in the SDK")
		require.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, factory.SDK.ConsumersSDK.AssertExpectations(t))
		}, waitTime, tickTime)
	})

	t.Run("removing referenced CP sets the status conditions properly", func(t *testing.T) {
		const (
			id   = "abc-12345"
			name = "name-1"
		)

		t.Log("Creating KonnectAPIAuthConfiguration and KonnectGatewayControlPlane")
		apiAuth := deploy.KonnectAPIAuthConfigurationWithProgrammed(t, ctx, clientNamespaced)
		cp := deploy.KonnectGatewayControlPlaneWithID(t, ctx, clientNamespaced, apiAuth)

		t.Log("Setting up a watch for KongConsumer events")
		w := setupWatch[configurationv1.KongConsumerList](t, ctx, cl, client.InNamespace(ns.Name))

		t.Log("Setting up SDK expectations on KongConsumer creation")
		sdk.ConsumersSDK.EXPECT().
			CreateConsumer(
				mock.Anything,
				cp.GetKonnectID(),
				mock.MatchedBy(func(req sdkkonnectcomp.ConsumerInput) bool {
					return req.Username != nil && *req.Username == name
				}),
			).
			Return(
				&sdkkonnectops.CreateConsumerResponse{
					Consumer: &sdkkonnectcomp.Consumer{
						ID:       lo.ToPtr(id),
						Username: lo.ToPtr(name),
					},
				},
				nil,
			)

		t.Log("Setting up SDK expectation on KongConsumerGroups listing")
		sdk.ConsumerGroupSDK.EXPECT().
			ListConsumerGroupsForConsumer(mock.Anything, sdkkonnectops.ListConsumerGroupsForConsumerRequest{
				ConsumerID:     id,
				ControlPlaneID: cp.GetKonnectStatus().GetKonnectID(),
			}).Return(&sdkkonnectops.ListConsumerGroupsForConsumerResponse{}, nil)

		created := deploy.KongConsumer(t, ctx, clientNamespaced, name,
			deploy.WithKonnectIDControlPlaneRef(cp),
			func(obj client.Object) {
				cert := obj.(*configurationv1.KongConsumer)
				cert.Username = name
			},
		)
		t.Log("Checking SDK Consumer operations")
		require.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, factory.SDK.ConsumersSDK.AssertExpectations(t))
		}, waitTime, tickTime)

		t.Log("Waiting for object to be programmed and get Konnect ID")
		watchFor(t, ctx, w, watch.Modified, conditionProgrammedIsSetToTrue(created, id),
			fmt.Sprintf("Consumer didn't get Programmed status condition or didn't get the correct %s Konnect ID assigned", id))

		t.Log("Deleting KonnectGatewayControlPlane")
		require.NoError(t, clientNamespaced.Delete(ctx, cp))

		t.Log("Waiting for KongConsumer to be get Programmed and ControlPlaneRefValid conditions with status=False")
		watchFor(t, ctx, w, watch.Modified,
			conditionsAreSetWhenReferencedControlPlaneIsMissing(created),
			"KongConsumer didn't get Programmed and/or ControlPlaneRefValid status condition set to False",
		)
	})
}

func TestKongConsumerSecretCredentials(t *testing.T) {
	t.Parallel()
	ctx, cancel := Context(t, context.Background())
	defer cancel()
	cfg, ns := Setup(t, ctx, scheme.Get())

	t.Log("Setting up the manager with reconcilers")
	mgr, logs := NewManager(t, ctx, cfg, scheme.Get(), WithKonnectCacheIndices(ctx))
	factory := sdkmocks.NewMockSDKFactory(t)
	sdk := factory.SDK
	reconcilers := []Reconciler{
		konnect.NewKonnectEntityReconciler(factory, false, mgr.GetClient(),
			konnect.WithKonnectEntitySyncPeriod[configurationv1.KongConsumer](konnectInfiniteSyncTime),
		),
		konnect.NewKonnectEntityReconciler(factory, false, mgr.GetClient(),
			konnect.WithKonnectEntitySyncPeriod[configurationv1alpha1.KongCredentialBasicAuth](konnectInfiniteSyncTime),
		),
		konnect.NewKonnectEntityReconciler(factory, false, mgr.GetClient(),
			konnect.WithKonnectEntitySyncPeriod[configurationv1alpha1.KongCredentialAPIKey](konnectInfiniteSyncTime),
		),
		konnect.NewKongCredentialSecretReconciler(false, mgr.GetClient(), mgr.GetScheme()),
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

	t.Log("Setting up a watch for KongConsumer events")
	cWatch := setupWatch[configurationv1.KongConsumerList](t, ctx, cl, client.InNamespace(ns.Name))

	t.Run("BasicAuth", func(t *testing.T) {
		consumerID := fmt.Sprintf("consumer-%d", rand.Int31n(1000))
		username := fmt.Sprintf("user-secret-credentials-%d", rand.Int31n(1000))
		cp := deploy.KonnectGatewayControlPlaneWithID(t, ctx, clientNamespaced, apiAuth)

		t.Log("Setting up SDK expectations on KongConsumer creation")
		sdk.ConsumersSDK.EXPECT().
			CreateConsumer(mock.Anything, cp.GetKonnectStatus().GetKonnectID(),
				mock.MatchedBy(func(input sdkkonnectcomp.ConsumerInput) bool {
					return input.Username != nil && *input.Username == username
				}),
			).Return(&sdkkonnectops.CreateConsumerResponse{
			Consumer: &sdkkonnectcomp.Consumer{
				ID: lo.ToPtr(consumerID),
			},
		}, nil)

		t.Log("Setting up SDK expectation on possibly updating KongConsumer ( due to asynchronous nature of updates between KongConsumer and KongConsumerGroup)")
		sdk.ConsumersSDK.EXPECT().
			UpsertConsumer(mock.Anything, mock.MatchedBy(func(r sdkkonnectops.UpsertConsumerRequest) bool {
				return r.ConsumerID == consumerID
			})).
			Return(&sdkkonnectops.UpsertConsumerResponse{}, nil).
			Maybe()

		t.Log("Setting up SDK expectation on KongConsumerGroups listing")
		sdk.ConsumerGroupSDK.EXPECT().
			ListConsumerGroupsForConsumer(mock.Anything, sdkkonnectops.ListConsumerGroupsForConsumerRequest{
				ConsumerID:     consumerID,
				ControlPlaneID: cp.GetKonnectStatus().GetKonnectID(),
			}).Return(&sdkkonnectops.ListConsumerGroupsForConsumerResponse{}, nil)

		s := deploy.Secret(t, ctx, clientNamespaced,
			map[string][]byte{
				"username": []byte(username),
				"password": []byte("password"),
			},
			deploy.WithLabel("konghq.com/credential", konnect.KongCredentialTypeBasicAuth),
		)

		t.Log("Setting up SDK expectation on (managed) BasicAuth credentials creation")
		sdk.KongCredentialsBasicAuthSDK.EXPECT().
			CreateBasicAuthWithConsumer(
				mock.Anything,
				mock.MatchedBy(
					func(r sdkkonnectops.CreateBasicAuthWithConsumerRequest) bool {
						return r.ControlPlaneID == cp.GetKonnectID() &&
							r.BasicAuthWithoutParents.Username == username &&
							r.BasicAuthWithoutParents.Password == "password"
					},
				),
			).
			Return(
				&sdkkonnectops.CreateBasicAuthWithConsumerResponse{
					BasicAuth: &sdkkonnectcomp.BasicAuth{
						ID: lo.ToPtr("basic-auth-id"),
					},
				},
				nil,
			)

		t.Log("Creating KongConsumer with ControlPlaneRef type=konnectID")
		createdConsumer := deploy.KongConsumer(t, ctx, clientNamespaced, username,
			deploy.WithKonnectIDControlPlaneRef(cp),
			func(obj client.Object) {
				consumer := obj.(*configurationv1.KongConsumer)
				consumer.Credentials = []string{
					s.Name,
				}
			},
		)

		t.Log("Waiting for KongConsumer to be programmed")
		watchFor(t, ctx, cWatch, watch.Modified, func(c *configurationv1.KongConsumer) bool {
			if c.GetName() != createdConsumer.GetName() {
				return false
			}
			if c.GetControlPlaneRef().Type != commonv1alpha1.ControlPlaneRefKonnectID {
				return false
			}
			return lo.ContainsBy(c.Status.Conditions, func(condition metav1.Condition) bool {
				return condition.Type == konnectv1alpha1.KonnectEntityProgrammedConditionType &&
					condition.Status == metav1.ConditionTrue
			})
		}, "KongConsumer's Programmed condition should be true eventually")

		t.Log("Waiting for KongConsumer to be created in the SDK")
		require.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, factory.SDK.ConsumersSDK.AssertExpectations(t))
		}, waitTime, tickTime)
		t.Log("Waiting for KongCredentialBasicAuth to be created in the SDK")
		require.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, factory.SDK.KongCredentialsBasicAuthSDK.AssertExpectations(t))
		}, waitTime, tickTime)
	})

	t.Run("APIKey", func(t *testing.T) {
		consumerID := fmt.Sprintf("consumer-%d", rand.Int31n(1000))
		username := fmt.Sprintf("user-secret-credentials-%d", rand.Int31n(1000))
		cp := deploy.KonnectGatewayControlPlaneWithID(t, ctx, clientNamespaced, apiAuth)

		t.Log("Setting up SDK expectations on KongConsumer creation")
		sdk.ConsumersSDK.EXPECT().
			CreateConsumer(mock.Anything, cp.GetKonnectStatus().GetKonnectID(),
				mock.MatchedBy(func(input sdkkonnectcomp.ConsumerInput) bool {
					return input.Username != nil && *input.Username == username
				}),
			).Return(&sdkkonnectops.CreateConsumerResponse{
			Consumer: &sdkkonnectcomp.Consumer{
				ID: lo.ToPtr(consumerID),
			},
		}, nil)

		t.Log("Setting up SDK expectation on possibly updating KongConsumer ( due to asynchronous nature of updates between KongConsumer and KongConsumerGroup)")
		sdk.ConsumersSDK.EXPECT().
			UpsertConsumer(mock.Anything, mock.MatchedBy(func(r sdkkonnectops.UpsertConsumerRequest) bool {
				return r.ConsumerID == consumerID
			})).
			Return(&sdkkonnectops.UpsertConsumerResponse{}, nil).
			Maybe()

		t.Log("Setting up SDK expectation on KongConsumerGroups listing")
		sdk.ConsumerGroupSDK.EXPECT().
			ListConsumerGroupsForConsumer(mock.Anything, sdkkonnectops.ListConsumerGroupsForConsumerRequest{
				ConsumerID:     consumerID,
				ControlPlaneID: cp.GetKonnectStatus().GetKonnectID(),
			}).Return(&sdkkonnectops.ListConsumerGroupsForConsumerResponse{}, nil)

		s := deploy.Secret(t, ctx, clientNamespaced,
			map[string][]byte{
				"key": []byte("api-key"),
			},
			deploy.WithLabel("konghq.com/credential", konnect.KongCredentialTypeAPIKey),
		)

		t.Log("Setting up SDK expectation on (managed) APIKey credentials creation")
		sdk.KongCredentialsAPIKeySDK.EXPECT().
			CreateKeyAuthWithConsumer(
				mock.Anything,
				mock.MatchedBy(
					func(r sdkkonnectops.CreateKeyAuthWithConsumerRequest) bool {
						return r.ControlPlaneID == cp.GetKonnectID() &&
							r.KeyAuthWithoutParents.Key == "api-key"
					},
				),
			).
			Return(
				&sdkkonnectops.CreateKeyAuthWithConsumerResponse{
					KeyAuth: &sdkkonnectcomp.KeyAuth{
						ID: lo.ToPtr("key-auth-id"),
					},
				},
				nil,
			)

		t.Log("Creating KongConsumer with ControlPlaneRef type=konnectID")
		createdConsumer := deploy.KongConsumer(t, ctx, clientNamespaced, username,
			deploy.WithKonnectIDControlPlaneRef(cp),
			func(obj client.Object) {
				consumer := obj.(*configurationv1.KongConsumer)
				consumer.Credentials = []string{
					s.Name,
				}
			},
		)

		t.Log("Waiting for KongConsumer to be programmed")
		watchFor(t, ctx, cWatch, watch.Modified, func(c *configurationv1.KongConsumer) bool {
			if c.GetName() != createdConsumer.GetName() {
				return false
			}
			if c.GetControlPlaneRef().Type != configurationv1alpha1.ControlPlaneRefKonnectID {
				return false
			}
			return lo.ContainsBy(c.Status.Conditions, func(condition metav1.Condition) bool {
				return condition.Type == konnectv1alpha1.KonnectEntityProgrammedConditionType &&
					condition.Status == metav1.ConditionTrue
			})
		}, "KongConsumer's Programmed condition should be true eventually")

		t.Log("Waiting for KongConsumer to be created in the SDK")
		require.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, factory.SDK.ConsumersSDK.AssertExpectations(t))
		}, waitTime, tickTime)
		t.Log("Waiting for KongCredentialAPIKey to be created in the SDK")
		require.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, factory.SDK.KongCredentialsAPIKeySDK.AssertExpectations(t))
		}, waitTime, tickTime)
	})
}
