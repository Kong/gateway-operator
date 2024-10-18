package envtest

import (
	"context"
	"strings"
	"testing"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectops "github.com/Kong/sdk-konnect-go/models/operations"
	sdkkonnecterrs "github.com/Kong/sdk-konnect-go/models/sdkerrors"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kong/gateway-operator/controller/konnect"
	sdkmocks "github.com/kong/gateway-operator/controller/konnect/ops/sdk/mocks"
	"github.com/kong/gateway-operator/modules/manager"
	"github.com/kong/gateway-operator/modules/manager/scheme"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
	"github.com/kong/gateway-operator/test/helpers/deploy"

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	"github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

func TestKongConsumerCredential_HMAC(t *testing.T) {
	t.Parallel()
	ctx, cancel := Context(t, context.Background())
	defer cancel()

	// Setup up the envtest environment.
	cfg, ns := Setup(t, ctx, scheme.Get())

	mgr, logs := NewManager(t, ctx, cfg, scheme.Get())

	cl, err := client.NewWithWatch(mgr.GetConfig(), client.Options{
		Scheme: scheme.Get(),
	})
	require.NoError(t, err)
	clientNamespaced := client.NewNamespacedClient(mgr.GetClient(), ns.Name)

	apiAuth := deploy.KonnectAPIAuthConfigurationWithProgrammed(t, ctx, clientNamespaced)
	cp := deploy.KonnectGatewayControlPlaneWithID(t, ctx, clientNamespaced, apiAuth)

	consumerID := uuid.NewString()
	consumer := deploy.KongConsumerWithProgrammed(t, ctx, clientNamespaced, &configurationv1.KongConsumer{
		Username: "username1",
		Spec: configurationv1.KongConsumerSpec{
			ControlPlaneRef: &configurationv1alpha1.ControlPlaneRef{
				Type: configurationv1alpha1.ControlPlaneRefKonnectNamespacedRef,
				KonnectNamespacedRef: &configurationv1alpha1.KonnectNamespacedRef{
					Name: cp.Name,
				},
			},
		},
	})
	consumer.Status.Konnect = &v1alpha1.KonnectEntityStatusWithControlPlaneRef{
		ControlPlaneID: cp.GetKonnectStatus().GetKonnectID(),
		KonnectEntityStatus: v1alpha1.KonnectEntityStatus{
			ID:        consumerID,
			ServerURL: cp.GetKonnectStatus().GetServerURL(),
			OrgID:     cp.GetKonnectStatus().GetOrgID(),
		},
	}
	require.NoError(t, clientNamespaced.Status().Update(ctx, consumer))

	kongCredentialHMAC := deploy.KongCredentialHMAC(t, ctx, clientNamespaced, consumer.Name)
	hmacID := uuid.NewString()
	tags := []string{
		"k8s-generation:1",
		"k8s-group:configuration.konghq.com",
		"k8s-kind:KongCredentialHMAC",
		"k8s-name:" + kongCredentialHMAC.Name,
		"k8s-namespace:" + ns.Name,
		"k8s-uid:" + string(kongCredentialHMAC.GetUID()),
		"k8s-version:v1alpha1",
	}

	factory := sdkmocks.NewMockSDKFactory(t)
	sdk := factory.SDK.KongCredentialsHMACSDK
	sdk.EXPECT().
		CreateHmacAuthWithConsumer(
			mock.Anything,
			sdkkonnectops.CreateHmacAuthWithConsumerRequest{
				ControlPlaneID:              cp.GetKonnectStatus().GetKonnectID(),
				ConsumerIDForNestedEntities: consumerID,
				HMACAuthWithoutParents: sdkkonnectcomp.HMACAuthWithoutParents{
					Username: lo.ToPtr("username"),
					Tags:     tags,
				},
			},
		).
		Return(
			&sdkkonnectops.CreateHmacAuthWithConsumerResponse{
				HMACAuth: &sdkkonnectcomp.HMACAuth{
					ID: lo.ToPtr(hmacID),
				},
			},
			nil,
		)
	sdk.EXPECT().
		UpsertHmacAuthWithConsumer(mock.Anything, mock.Anything, mock.Anything).Maybe().
		Return(
			&sdkkonnectops.UpsertHmacAuthWithConsumerResponse{
				HMACAuth: &sdkkonnectcomp.HMACAuth{
					ID: lo.ToPtr(hmacID),
				},
			},
			nil,
		)

	require.NoError(t, manager.SetupCacheIndicesForKonnectTypes(ctx, mgr, false))
	reconcilers := []Reconciler{
		konnect.NewKonnectEntityReconciler(factory, false, mgr.GetClient(),
			konnect.WithKonnectEntitySyncPeriod[configurationv1alpha1.KongCredentialHMAC](konnectInfiniteSyncTime),
		),
	}

	StartReconcilers(ctx, t, mgr, logs, reconcilers...)

	assert.EventuallyWithT(t,
		assertCollectObjectExistsAndHasKonnectID(t, ctx, clientNamespaced, kongCredentialHMAC, hmacID),
		waitTime, tickTime,
		"KongCredentialHMAC wasn't created",
	)

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.True(c, sdk.AssertExpectations(t))
	}, waitTime, tickTime)

	sdk.EXPECT().
		DeleteHmacAuthWithConsumer(
			mock.Anything,
			sdkkonnectops.DeleteHmacAuthWithConsumerRequest{
				ControlPlaneID:              cp.GetKonnectStatus().GetKonnectID(),
				ConsumerIDForNestedEntities: consumerID,
				HMACAuthID:                  hmacID,
			},
		).
		Return(
			&sdkkonnectops.DeleteHmacAuthWithConsumerResponse{
				StatusCode: 200,
			},
			nil,
		)
	require.NoError(t, clientNamespaced.Delete(ctx, kongCredentialHMAC))

	assert.EventuallyWithT(t,
		func(c *assert.CollectT) {
			assert.True(c, k8serrors.IsNotFound(
				clientNamespaced.Get(ctx, client.ObjectKeyFromObject(kongCredentialHMAC), kongCredentialHMAC),
			))
		}, waitTime, tickTime,
		"KongCredentialHMAC wasn't deleted but it should have been",
	)

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.True(c, sdk.AssertExpectations(t))
	}, waitTime, tickTime)

	t.Run("conflict on creation should be handled successfully", func(t *testing.T) {
		t.Log("Setting up SDK expectations on creation with conflict")
		sdk.EXPECT().
			CreateHmacAuthWithConsumer(
				mock.Anything,
				mock.MatchedBy(func(r sdkkonnectops.CreateHmacAuthWithConsumerRequest) bool {
					return r.ControlPlaneID == cp.GetKonnectID() &&
						r.ConsumerIDForNestedEntities == consumerID &&
						r.HMACAuthWithoutParents.Tags != nil &&
						slices.ContainsFunc(
							r.HMACAuthWithoutParents.Tags,
							func(t string) bool {
								return strings.HasPrefix(t, "k8s-uid:")
							},
						)
				},
				),
			).
			Return(
				nil,
				&sdkkonnecterrs.SDKError{
					StatusCode: 400,
					Body:       ErrBodyDataConstraintError,
				},
			)

		sdk.EXPECT().
			ListHmacAuth(
				mock.Anything,
				mock.MatchedBy(func(r sdkkonnectops.ListHmacAuthRequest) bool {
					return r.ControlPlaneID == cp.GetKonnectID() &&
						r.Tags != nil && strings.HasPrefix(*r.Tags, "k8s-uid")
				}),
			).
			Return(&sdkkonnectops.ListHmacAuthResponse{
				Object: &sdkkonnectops.ListHmacAuthResponseBody{
					Data: []sdkkonnectcomp.HMACAuth{
						{
							ID: lo.ToPtr(hmacID),
						},
					},
				},
			}, nil)

		w := setupWatch[configurationv1alpha1.KongCredentialHMACList](t, ctx, cl, client.InNamespace(ns.Name))
		created := deploy.KongCredentialHMAC(t, ctx, clientNamespaced, consumer.Name)

		t.Log("Waiting for KongCredentialHMAC to be programmed")
		watchFor(t, ctx, w, watch.Modified, func(k *configurationv1alpha1.KongCredentialHMAC) bool {
			return k.GetName() == created.GetName() &&
				k8sutils.IsProgrammed(k)
		}, "KongCredentialHMAC's Programmed condition should be true eventually")

		t.Log("Checking SDK KongCredentialHMAC operations")
		require.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.True(c, sdk.AssertExpectations(t))
		}, waitTime, tickTime)
	})
}
