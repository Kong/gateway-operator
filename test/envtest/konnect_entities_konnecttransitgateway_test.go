package envtest

import (
	"testing"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectops "github.com/Kong/sdk-konnect-go/models/operations"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiwatch "k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kong/gateway-operator/controller/konnect"
	sdkmocks "github.com/kong/gateway-operator/controller/konnect/ops/sdk/mocks"
	"github.com/kong/gateway-operator/modules/manager/logging"
	"github.com/kong/gateway-operator/modules/manager/scheme"
	"github.com/kong/gateway-operator/test/helpers/deploy"
	"github.com/kong/gateway-operator/test/helpers/eventually"

	commonv1alpha1 "github.com/kong/kubernetes-configuration/api/common/v1alpha1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

func TestKonnectCloudGatewayTransitGateway(t *testing.T) {
	t.Parallel()
	ctx, cancel := Context(t, t.Context())
	defer cancel()
	cfg, ns := Setup(t, ctx, scheme.Get())

	t.Log("Setting up the manager with reconcilers")
	mgr, logs := NewManager(t, ctx, cfg, scheme.Get())
	factory := sdkmocks.NewMockSDKFactory(t)
	sdk := factory.SDK
	StartReconcilers(ctx, t, mgr, logs,
		konnect.NewKonnectEntityReconciler(factory, logging.DevelopmentMode, mgr.GetClient(),
			konnect.WithKonnectEntitySyncPeriod[konnectv1alpha1.KonnectCloudGatewayTransitGateway](konnectInfiniteSyncTime),
		),
	)

	t.Log("Setting up clients")
	cl, err := client.NewWithWatch(mgr.GetConfig(), client.Options{
		Scheme: scheme.Get(),
	})
	require.NoError(t, err)
	clientNamespaced := client.NewNamespacedClient(mgr.GetClient(), ns.Name)

	t.Run("Creating and Deleting Konnect transit gateway", func(t *testing.T) {
		var (
			id        = "ktg-" + uuid.New().String()
			networkID = "network-" + uuid.New().String()

			transitGatewayName = "test-aws-transit-gateway-" + uuid.New().String()
		)

		t.Log("Setting up a watch for KonnectCloudGatewayTransitGateway events")
		w := setupWatch[konnectv1alpha1.KonnectCloudGatewayTransitGatewayList](t, ctx, cl, client.InNamespace(ns.Name))
		t.Log("Setting up SDK expectations on creation")
		sdk.TransitGatewaysSDK.EXPECT().CreateTransitGateway(
			mock.Anything,
			networkID,
			mock.MatchedBy(func(req sdkkonnectcomp.CreateTransitGatewayRequest) bool {
				return req.Type == sdkkonnectcomp.CreateTransitGatewayRequestTypeAWSTransitGateway &&
					req.AWSTransitGateway.Name == transitGatewayName &&
					req.AWSTransitGateway.TransitGatewayAttachmentConfig.Kind == sdkkonnectcomp.AWSTransitGatewayAttachmentTypeAwsTransitGatewayAttachment
			}),
		).Return(
			&sdkkonnectops.CreateTransitGatewayResponse{
				TransitGatewayResponse: &sdkkonnectcomp.TransitGatewayResponse{
					Type: sdkkonnectcomp.TransitGatewayResponseTypeAwsTransitGatewayResponse,
					AwsTransitGatewayResponse: &sdkkonnectcomp.AwsTransitGatewayResponse{
						Name:  transitGatewayName,
						ID:    id,
						State: sdkkonnectcomp.TransitGatewayStateCreated,
					},
				},
			},
			nil,
		)

		t.Log("Creating KonnectAPIAuthConfiguration and KonnectGatewayControlPlane")
		apiAuth := deploy.KonnectAPIAuthConfigurationWithProgrammed(t, ctx, clientNamespaced)

		t.Log("Creating a KonnectCloudGatewayNetwork and a KonnectCloudGatewayTransitGateway attaching to the network")
		n := deploy.KonnectCloudGatewayNetworkWithProgrammed(t, ctx, clientNamespaced, apiAuth,
			func(obj client.Object) {
				n := obj.(*konnectv1alpha1.KonnectCloudGatewayNetwork)
				n.Status.State = string(sdkkonnectcomp.NetworkStateReady)
			},
			deploy.WithKonnectID(networkID),
		)

		tg := &konnectv1alpha1.KonnectCloudGatewayTransitGateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: transitGatewayName,
			},
			Spec: konnectv1alpha1.KonnectCloudGatewayTransitGatewaySpec{
				NetworkRef: commonv1alpha1.ObjectRef{
					Type: commonv1alpha1.ObjectRefTypeNamespacedRef,
					NamespacedRef: &commonv1alpha1.NamespacedRef{
						Name: n.Name,
					},
				},
				KonnectTransitGatewayAPISpec: konnectv1alpha1.KonnectTransitGatewayAPISpec{
					Type: konnectv1alpha1.TransitGatewayTypeAWSTransitGateway,
					AWSTransitGateway: &konnectv1alpha1.AWSTransitGateway{
						Name:       transitGatewayName,
						CIDRBlocks: []string{"10.10.10.0/24"},
						AttachmentConfig: konnectv1alpha1.AwsTransitGatewayAttachmentConfig{
							TransitGatewayID: "tgw-012345abcdef",
							RAMShareArn:      "ram_share_arn",
						},
					},
				},
			},
		}
		require.NoError(t, clientNamespaced.Create(ctx, tg))

		t.Log("Waiting for KonnectCloudGatewayTransitGateway to be Programmed and get a Konnect ID")
		watchFor(t, ctx, w, apiwatch.Modified, func(tg *konnectv1alpha1.KonnectCloudGatewayTransitGateway) bool {
			return tg.GetKonnectID() == id && conditionsContainProgrammed(tg.GetConditions(), metav1.ConditionTrue)
		}, "Did not see KonnectCloudGatewayTransitGateway get Programmed and Konnect ID set.")

		t.Log("Setting up SDK expctations on updating/get")
		sdk.TransitGatewaysSDK.EXPECT().GetTransitGateway(mock.Anything, networkID, id).Return(
			&sdkkonnectops.GetTransitGatewayResponse{
				TransitGatewayResponse: &sdkkonnectcomp.TransitGatewayResponse{
					Type: sdkkonnectcomp.TransitGatewayResponseTypeAwsTransitGatewayResponse,
					AwsTransitGatewayResponse: &sdkkonnectcomp.AwsTransitGatewayResponse{
						Name:  transitGatewayName,
						ID:    id,
						State: sdkkonnectcomp.TransitGatewayStateReady,
					},
				},
			}, nil,
		)

		t.Log("Updating KonnectCloudGatewayTransitGateway")
		require.NoError(t, clientNamespaced.Get(ctx, client.ObjectKeyFromObject(tg), tg))
		oldTg := tg.DeepCopy()
		tg.Spec.AWSTransitGateway.AttachmentConfig.RAMShareArn = "ram_share_arn_"
		require.NoError(t, clientNamespaced.Patch(ctx, tg, client.MergeFrom(oldTg)))
		watchFor(t, ctx, w, apiwatch.Modified, func(tg *konnectv1alpha1.KonnectCloudGatewayTransitGateway) bool {
			return tg.GetKonnectID() == id && tg.Status.State == sdkkonnectcomp.TransitGatewayStateReady
		}, "Did not see KonnectCloudGatewayTransitGateway get status.state updated")

		t.Log("Setting up SDK expectations on deletion")
		sdk.TransitGatewaysSDK.EXPECT().DeleteTransitGateway(mock.Anything, networkID, id).Return(
			&sdkkonnectops.DeleteTransitGatewayResponse{}, nil,
		)
		t.Log("Deleting")
		require.NoError(t, clientNamespaced.Delete(ctx, tg))
		eventually.WaitForObjectToNotExist(t, ctx, cl, tg, waitTime, tickTime)

		t.Log("Waiting for object to be deleted in the SDK")
		eventuallyAssertSDKExpectations(t, factory.SDK.TransitGatewaysSDK, waitTime, tickTime)
	})
}
