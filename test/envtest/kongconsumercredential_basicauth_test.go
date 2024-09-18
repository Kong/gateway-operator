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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kong/gateway-operator/controller/konnect"
	"github.com/kong/gateway-operator/controller/konnect/ops"
	"github.com/kong/gateway-operator/modules/manager"
	"github.com/kong/gateway-operator/modules/manager/scheme"

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	"github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

func TestKongConsumerCredential_BasicAuth(t *testing.T) {
	t.Parallel()
	ctx, cancel := Context(t, context.Background())
	defer cancel()

	// Setup up the envtest environment.
	cfg, ns := Setup(t, ctx, scheme.Get())

	mgr, logs := NewManager(t, ctx, cfg, scheme.Get())

	clientNamespaced := client.NewNamespacedClient(mgr.GetClient(), ns.Name)

	apiAuth := deployKonnectAPIAuthConfigurationWithProgrammed(t, ctx, clientNamespaced)
	cp := deployKonnectGatewayControlPlaneWithID(t, ctx, clientNamespaced, apiAuth)

	consumerID := uuid.NewString()
	consumer := deployKongConsumerWithProgrammed(t, ctx, clientNamespaced, &configurationv1.KongConsumer{
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

	password := "password"
	username := "username"
	credentialBasicAuth := &configurationv1alpha1.CredentialBasicAuth{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "basic-auth-",
		},
		Spec: configurationv1alpha1.CredentialBasicAuthSpec{
			ConsumerRef: corev1.LocalObjectReference{
				Name: consumer.Name,
			},
			CredentialBasicAuthAPISpec: configurationv1alpha1.CredentialBasicAuthAPISpec{
				Password: password,
				Username: username,
			},
		},
	}
	require.NoError(t, clientNamespaced.Create(ctx, credentialBasicAuth))
	t.Logf("deployed %s CredentialBasicAuth resource", client.ObjectKeyFromObject(credentialBasicAuth))

	basicAuthID := uuid.NewString()
	tags := []string{
		"k8s-generation:1",
		"k8s-group:configuration.konghq.com",
		"k8s-kind:CredentialBasicAuth",
		"k8s-name:" + credentialBasicAuth.Name,
		"k8s-uid:" + string(credentialBasicAuth.GetUID()),
		"k8s-version:v1alpha1",
		"k8s-namespace:" + ns.Name,
	}

	factory := ops.NewMockSDKFactory(t)
	factory.SDK.BasicAuthCredentials.EXPECT().
		CreateBasicAuthWithConsumer(
			mock.Anything,
			sdkkonnectops.CreateBasicAuthWithConsumerRequest{
				ControlPlaneID:              cp.GetKonnectStatus().GetKonnectID(),
				ConsumerIDForNestedEntities: consumerID,
				BasicAuthWithoutParents: sdkkonnectcomp.BasicAuthWithoutParents{
					Password: lo.ToPtr(password),
					Username: lo.ToPtr(username),
					Tags:     tags,
				},
			},
		).
		Return(
			&sdkkonnectops.CreateBasicAuthWithConsumerResponse{
				BasicAuth: &sdkkonnectcomp.BasicAuth{
					ID: lo.ToPtr(basicAuthID),
				},
			},
			nil,
		)
	factory.SDK.BasicAuthCredentials.EXPECT().
		UpsertBasicAuthWithConsumer(mock.Anything, mock.Anything, mock.Anything).Maybe().
		Return(
			&sdkkonnectops.UpsertBasicAuthWithConsumerResponse{
				BasicAuth: &sdkkonnectcomp.BasicAuth{
					ID: lo.ToPtr(basicAuthID),
				},
			},
			nil,
		)

	require.NoError(t, manager.SetupCacheIndicesForKonnectTypes(ctx, mgr, false))
	reconcilers := []Reconciler{
		konnect.NewKonnectEntityReconciler(factory, false, mgr.GetClient(),
			konnect.WithKonnectEntitySyncPeriod[configurationv1alpha1.CredentialBasicAuth](konnectSyncTime),
		),
	}

	StartReconcilers(ctx, t, mgr, logs, reconcilers...)

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.True(c, factory.SDK.BasicAuthCredentials.AssertExpectations(t))
	}, waitTime, tickTime)

	factory.SDK.BasicAuthCredentials.EXPECT().
		DeleteBasicAuthWithConsumer(
			mock.Anything,
			sdkkonnectops.DeleteBasicAuthWithConsumerRequest{
				ControlPlaneID:              cp.GetKonnectStatus().GetKonnectID(),
				ConsumerIDForNestedEntities: consumerID,
				BasicAuthID:                 basicAuthID,
			},
		).
		Return(
			&sdkkonnectops.DeleteBasicAuthWithConsumerResponse{
				StatusCode: 200,
			},
			nil,
		)
	require.NoError(t, clientNamespaced.Delete(ctx, credentialBasicAuth))

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.True(c, factory.SDK.BasicAuthCredentials.AssertExpectations(t))
	}, waitTime, tickTime)
}