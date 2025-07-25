package adminapi_test

import (
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/kong/kong-operator/ingress-controller/internal/adminapi"
	managercfg "github.com/kong/kong-operator/ingress-controller/pkg/manager/config"
	"github.com/kong/kong-operator/ingress-controller/test/mocks"
)

func TestClientFactory_CreateAdminAPIClientAttachesPodReference(t *testing.T) {
	factory := adminapi.NewClientFactoryForWorkspace(logr.Discard(), "workspace", managercfg.AdminAPIClientConfig{}, "")

	adminAPIHandler := mocks.NewAdminAPIHandler(t)
	adminAPIServer := httptest.NewServer(adminAPIHandler)
	t.Cleanup(func() { adminAPIServer.Close() })

	client, err := factory.CreateAdminAPIClient(t.Context(), adminapi.DiscoveredAdminAPI{
		Address: adminAPIServer.URL,
		PodRef: k8stypes.NamespacedName{
			Namespace: "namespace",
			Name:      "name",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, client)

	ref, ok := client.PodReference()
	require.True(t, ok, "expected pod reference to be attached to the client")
	require.Equal(t, k8stypes.NamespacedName{
		Namespace: "namespace",
		Name:      "name",
	}, ref)
}
