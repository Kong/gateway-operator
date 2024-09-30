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
	"github.com/kong/gateway-operator/controller/konnect/ops"
	"github.com/kong/gateway-operator/modules/manager/scheme"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	konnectalpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

func TestKongSNI(t *testing.T) {
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
			konnect.WithKonnectEntitySyncPeriod[configurationv1alpha1.KongSNI](konnectInfiniteSyncTime),
		),
	)

	t.Log("Setting up clients")
	cl, err := client.NewWithWatch(mgr.GetConfig(), client.Options{
		Scheme: scheme.Get(),
	})
	require.NoError(t, err)
	clientNamespaced := client.NewNamespacedClient(mgr.GetClient(), ns.Name)

	t.Log("Creating KonnectAPIAuthConfiguration and KonnectGatewayControlPlane")
	apiAuth := deployKonnectAPIAuthConfigurationWithProgrammed(t, ctx, clientNamespaced)
	cp := deployKonnectGatewayControlPlaneWithID(t, ctx, clientNamespaced, apiAuth)

	t.Log("Creating KongCertificate and setting it to Programmed")
	createdCert := deployKongCertificateAttachedToCP(t, ctx, clientNamespaced, cp)
	createdCert.Status = configurationv1alpha1.KongCertificateStatus{
		Konnect: &konnectalpha1.KonnectEntityStatusWithControlPlaneRef{
			KonnectEntityStatus: konnectEntityStatus("cert-12345"),
			ControlPlaneID:      cp.Status.GetKonnectID(),
		},
		Conditions: []metav1.Condition{
			{
				Type:               conditions.KonnectEntityProgrammedConditionType,
				Status:             metav1.ConditionTrue,
				Reason:             conditions.KonnectEntityProgrammedReasonProgrammed,
				ObservedGeneration: createdCert.GetGeneration(),
				LastTransitionTime: metav1.Now(),
			},
		},
	}
	require.NoError(t, clientNamespaced.Status().Update(ctx, createdCert))

	t.Log("Setting up a watch for KongSNI events")
	w := setupWatch[configurationv1alpha1.KongSNIList](t, ctx, cl, client.InNamespace(ns.Name))

	t.Log("Setting up SDK for creating SNI")
	sdk.SNIsSDK.EXPECT().CreateSniWithCertificate(
		mock.Anything,
		mock.MatchedBy(func(req sdkkonnectops.CreateSniWithCertificateRequest) bool {
			return req.ControlPlaneID == cp.Status.ID &&
				req.CertificateID == createdCert.GetKonnectID() &&
				req.SNIWithoutParents.Name == "test.kong-sni.example.com"
		}),
	).Return(&sdkkonnectops.CreateSniWithCertificateResponse{
		Sni: &sdkkonnectcomp.Sni{
			ID: lo.ToPtr("sni-12345"),
		},
	}, nil)

	t.Log("Creating KongSNI")
	createdSNI := deploySNIAttachedToCertificate(t, ctx,
		clientNamespaced,
		"test.kong-sni.example.com", nil,
		createdCert,
	)

	t.Log("Waiting for SNI to be programmed and got Konnect ID")
	watchFor(t, ctx, w, watch.Modified, func(s *configurationv1alpha1.KongSNI) bool {
		return s.GetKonnectID() == "sni-12345" && lo.ContainsBy(s.Status.Conditions,
			func(c metav1.Condition) bool {
				return c.Type == "Programmed" && c.Status == metav1.ConditionTrue
			})
	}, "")

	t.Log("Set up SDK for SNI update")
	sdk.SNIsSDK.EXPECT().UpsertSniWithCertificate(
		mock.Anything,
		mock.MatchedBy(func(req sdkkonnectops.UpsertSniWithCertificateRequest) bool {
			return req.CertificateID == createdCert.GetKonnectID() &&
				req.ControlPlaneID == cp.Status.ID &&
				req.SNIWithoutParents.Name == "test2.kong-sni.example.com"
		}),
	).Return(&sdkkonnectops.UpsertSniWithCertificateResponse{}, nil)

	t.Log("Patching KongSNI")
	sniToPatch := createdSNI.DeepCopy()
	sniToPatch.Spec.KongSNIAPISpec.Name = "test2.kong-sni.example.com"
	require.NoError(t, clientNamespaced.Patch(ctx, sniToPatch, client.MergeFrom(createdSNI)))

	t.Log("Waiting for KongSNI to be updated in the SDK")
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.True(c, factory.SDK.SNIsSDK.AssertExpectations(t))
	}, waitTime, tickTime)

	t.Log("Setting up SDK for deleting SNI")
	sdk.SNIsSDK.EXPECT().DeleteSniWithCertificate(
		mock.Anything,
		sdkkonnectops.DeleteSniWithCertificateRequest{
			ControlPlaneID: cp.Status.ID,
			CertificateID:  createdCert.GetKonnectID(),
			SNIID:          "sni-12345",
		},
	).Return(&sdkkonnectops.DeleteSniWithCertificateResponse{}, nil)

	t.Log("Deleting KongSNI")
	require.NoError(t, clientNamespaced.Delete(ctx, createdSNI))

	t.Log("Waiting for SNI to be deleted in SDK")
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.True(c, factory.SDK.SNIsSDK.AssertExpectations(t))
	}, waitTime, tickTime)
}
