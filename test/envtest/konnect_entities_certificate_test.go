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
	"github.com/kong/gateway-operator/controller/konnect/ops"
	"github.com/kong/gateway-operator/modules/manager/scheme"
	"github.com/kong/gateway-operator/test/helpers/deploy"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

func TestKongCertificate(t *testing.T) {
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
			konnect.WithKonnectEntitySyncPeriod[configurationv1alpha1.KongCertificate](konnectInfiniteSyncTime),
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

	t.Log("Setting up SDK expectations on KongCertificate creation")
	sdk.CertificatesSDK.EXPECT().CreateCertificate(mock.Anything, cp.GetKonnectStatus().GetKonnectID(),
		mock.MatchedBy(func(input sdkkonnectcomp.CertificateInput) bool {
			return input.Cert == deploy.TestValidCertPEM && input.Key == deploy.TestValidCertKeyPEM
		}),
	).Return(&sdkkonnectops.CreateCertificateResponse{
		Certificate: &sdkkonnectcomp.Certificate{
			ID: lo.ToPtr("cert-12345"),
		},
	}, nil)

	t.Log("Setting up a watch for KongCertificate events")
	w := setupWatch[configurationv1alpha1.KongCertificateList](t, ctx, cl, client.InNamespace(ns.Name))

	t.Log("Creating KongCertificate")
	createdCert := deploy.KongCertificateAttachedToCP(t, ctx, clientNamespaced, cp)

	t.Log("Waiting for KongCertificate to be programmed")
	watchFor(t, ctx, w, watch.Modified, func(c *configurationv1alpha1.KongCertificate) bool {
		if c.GetName() != createdCert.GetName() {
			return false
		}
		return lo.ContainsBy(c.Status.Conditions, func(condition metav1.Condition) bool {
			return condition.Type == konnectv1alpha1.KonnectEntityProgrammedConditionType &&
				condition.Status == metav1.ConditionTrue
		})
	}, "KongCertificate's Programmed condition should be true eventually")

	t.Log("Waiting for KongCertificate to be created in the SDK")
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.True(c, factory.SDK.CertificatesSDK.AssertExpectations(t))
	}, waitTime, tickTime)

	t.Log("Setting up SDK expectations on KongCertificate update")
	sdk.CertificatesSDK.EXPECT().UpsertCertificate(mock.Anything, mock.MatchedBy(func(r sdkkonnectops.UpsertCertificateRequest) bool {
		return r.CertificateID == "cert-12345" &&
			lo.Contains(r.Certificate.Tags, "addedTag")
	})).Return(&sdkkonnectops.UpsertCertificateResponse{}, nil)

	t.Log("Patching KongCertificate")
	certToPatch := createdCert.DeepCopy()
	certToPatch.Spec.Tags = append(certToPatch.Spec.Tags, "addedTag")
	require.NoError(t, clientNamespaced.Patch(ctx, certToPatch, client.MergeFrom(createdCert)))

	t.Log("Waiting for KongCertificate to be updated in the SDK")
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.True(c, factory.SDK.CertificatesSDK.AssertExpectations(t))
	}, waitTime, tickTime)

	t.Log("Setting up SDK expectations on KongCertificate deletion")
	sdk.CertificatesSDK.EXPECT().DeleteCertificate(mock.Anything, cp.GetKonnectStatus().GetKonnectID(), "cert-12345").
		Return(&sdkkonnectops.DeleteCertificateResponse{}, nil)

	t.Log("Deleting KongCertificate")
	require.NoError(t, cl.Delete(ctx, createdCert))

	t.Log("Waiting for KongCertificate to be deleted in the SDK")
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.True(c, factory.SDK.CertificatesSDK.AssertExpectations(t))
	}, waitTime, tickTime)
}
