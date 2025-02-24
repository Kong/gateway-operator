package envtest

import (
	"fmt"
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
	sdkmocks "github.com/kong/gateway-operator/controller/konnect/ops/sdk/mocks"
	"github.com/kong/gateway-operator/modules/manager/scheme"
	"github.com/kong/gateway-operator/test/helpers/deploy"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

func TestKongDataPlaneClientCertificate(t *testing.T) {
	t.Parallel()
	ctx, cancel := Context(t, t.Context())
	defer cancel()
	cfg, ns := Setup(t, ctx, scheme.Get())

	t.Log("Setting up the manager with reconcilers")
	mgr, logs := NewManager(t, ctx, cfg, scheme.Get(), WithKonnectCacheIndices(ctx))
	factory := sdkmocks.NewMockSDKFactory(t)
	sdk := factory.SDK
	StartReconcilers(ctx, t, mgr, logs,
		konnect.NewKonnectEntityReconciler(factory, false, mgr.GetClient(),
			konnect.WithKonnectEntitySyncPeriod[configurationv1alpha1.KongDataPlaneClientCertificate](konnectInfiniteSyncTime),
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

	t.Log("Setting up SDK expectations on KongDataPlaneClientCertificate creation")
	const dpCertID = "dp-cert-id"
	sdk.DataPlaneCertificatesSDK.EXPECT().CreateDataplaneCertificate(mock.Anything, cp.GetKonnectStatus().GetKonnectID(),
		mock.MatchedBy(func(input *sdkkonnectcomp.DataPlaneClientCertificateRequest) bool {
			return input.Cert == deploy.TestValidCACertPEM
		}),
	).Return(&sdkkonnectops.CreateDataplaneCertificateResponse{
		DataPlaneClientCertificate: &sdkkonnectcomp.DataPlaneClientCertificate{
			Item: &sdkkonnectcomp.DataPlaneClientCertificateItem{
				ID:   lo.ToPtr(dpCertID),
				Cert: lo.ToPtr(deploy.TestValidCACertPEM),
			},
		},
	}, nil)

	w := setupWatch[configurationv1alpha1.KongDataPlaneClientCertificateList](t, ctx, cl, client.InNamespace(ns.Name))

	t.Log("Creating KongDataPlaneClientCertificate")
	createdCert := deploy.KongDataPlaneClientCertificateAttachedToCP(t, ctx, clientNamespaced,
		deploy.WithKonnectNamespacedRefControlPlaneRef(cp),
	)

	t.Log("Waiting for KongDataPlaneClientCertificate to be programmed")
	watchFor(t, ctx, w, watch.Modified, func(c *configurationv1alpha1.KongDataPlaneClientCertificate) bool {
		if c.GetName() != createdCert.GetName() {
			return false
		}
		return lo.ContainsBy(c.Status.Conditions, func(condition metav1.Condition) bool {
			return condition.Type == konnectv1alpha1.KonnectEntityProgrammedConditionType &&
				condition.Status == metav1.ConditionTrue
		})
	}, "KongDataPlaneClientCertificate's Programmed condition should be true eventually")

	t.Log("Waiting for KongDataPlaneClientCertificate to be created in the SDK")
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.True(c, factory.SDK.CACertificatesSDK.AssertExpectations(t))
	}, waitTime, tickTime)

	t.Log("Setting up SDK expectations on KongDataPlaneClientCertificate deletion")
	sdk.DataPlaneCertificatesSDK.EXPECT().DeleteDataplaneCertificate(mock.Anything, cp.GetKonnectStatus().GetKonnectID(), dpCertID).
		Return(&sdkkonnectops.DeleteDataplaneCertificateResponse{}, nil)

	t.Log("Deleting KongDataPlaneClientCertificate")
	require.NoError(t, cl.Delete(ctx, createdCert))

	t.Log("Waiting for KongDataPlaneClientCertificate to be deleted in the SDK")
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.True(c, factory.SDK.CACertificatesSDK.AssertExpectations(t))
	}, waitTime, tickTime)

	t.Run("should handle konnectID control plane reference", func(t *testing.T) {
		t.Log("Setting up SDK expectations on KongDataPlaneClientCertificate creation")
		const dpCertID = "dp-cert-id-with-konnectid-cp-ref"
		sdk.DataPlaneCertificatesSDK.EXPECT().CreateDataplaneCertificate(mock.Anything, cp.GetKonnectStatus().GetKonnectID(),
			mock.MatchedBy(func(input *sdkkonnectcomp.DataPlaneClientCertificateRequest) bool {
				return input.Cert == deploy.TestValidCACertPEM
			}),
		).Return(&sdkkonnectops.CreateDataplaneCertificateResponse{
			DataPlaneClientCertificate: &sdkkonnectcomp.DataPlaneClientCertificate{
				Item: &sdkkonnectcomp.DataPlaneClientCertificateItem{
					ID:   lo.ToPtr(dpCertID),
					Cert: lo.ToPtr(deploy.TestValidCACertPEM),
				},
			},
		}, nil)

		t.Log("Creating KongDataPlaneClientCertificate with ControlPlaneRef type=konnectID")
		createdCert := deploy.KongDataPlaneClientCertificateAttachedToCP(t, ctx, clientNamespaced,
			deploy.WithKonnectIDControlPlaneRef(cp),
		)

		t.Log("Waiting for KongDataPlaneClientCertificate to be programmed")
		watchFor(t, ctx, w, watch.Modified, func(c *configurationv1alpha1.KongDataPlaneClientCertificate) bool {
			if c.GetName() != createdCert.GetName() {
				return false
			}
			if c.GetControlPlaneRef().Type != configurationv1alpha1.ControlPlaneRefKonnectID {
				return false
			}
			return lo.ContainsBy(c.Status.Conditions, func(condition metav1.Condition) bool {
				return condition.Type == konnectv1alpha1.KonnectEntityProgrammedConditionType &&
					condition.Status == metav1.ConditionTrue
			})
		}, "KongDataPlaneClientCertificate's Programmed condition should be true eventually")

		eventuallyAssertSDKExpectations(t, factory.SDK.CACertificatesSDK, waitTime, tickTime)
	})

	t.Run("removing referenced CP sets the status conditions properly", func(t *testing.T) {
		const (
			id = "abc-12345"
		)

		t.Log("Creating KonnectAPIAuthConfiguration and KonnectGatewayControlPlane")
		apiAuth := deploy.KonnectAPIAuthConfigurationWithProgrammed(t, ctx, clientNamespaced)
		cp := deploy.KonnectGatewayControlPlaneWithID(t, ctx, clientNamespaced, apiAuth)

		w := setupWatch[configurationv1alpha1.KongDataPlaneClientCertificateList](t, ctx, cl, client.InNamespace(ns.Name))

		t.Log("Setting up SDK expectations on KongDataPlaneClientCertificate creation")
		sdk.DataPlaneCertificatesSDK.EXPECT().
			CreateDataplaneCertificate(
				mock.Anything,
				cp.GetKonnectID(),
				mock.Anything,
			).
			Return(
				&sdkkonnectops.CreateDataplaneCertificateResponse{
					DataPlaneClientCertificate: &sdkkonnectcomp.DataPlaneClientCertificate{
						Item: &sdkkonnectcomp.DataPlaneClientCertificateItem{
							ID: lo.ToPtr(id),
						},
					},
				},
				nil,
			)

		created := deploy.KongDataPlaneClientCertificateAttachedToCP(t, ctx, clientNamespaced,
			deploy.WithKonnectIDControlPlaneRef(cp),
		)
		eventuallyAssertSDKExpectations(t, factory.SDK.DataPlaneCertificatesSDK, waitTime, tickTime)

		t.Log("Waiting for object to be programmed and get Konnect ID")
		watchFor(t, ctx, w, watch.Modified, conditionProgrammedIsSetToTrue(created, id),
			fmt.Sprintf("DataPlaneClientCertificate didn't get Programmed status condition or didn't get the correct %s Konnect ID assigned", id))

		t.Log("Deleting KonnectGatewayControlPlane")
		require.NoError(t, clientNamespaced.Delete(ctx, cp))

		t.Log("Waiting for DataPlaneClientCertificate to be get Programmed and ControlPlaneRefValid conditions with status=False")
		watchFor(t, ctx, w, watch.Modified,
			conditionsAreSetWhenReferencedControlPlaneIsMissing(created),
			"KongDataPlaneClientCertificate didn't get Programmed and/or ControlPlaneRefValid status condition set to False",
		)
	})
}
