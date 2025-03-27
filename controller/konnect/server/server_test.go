package server_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kong/gateway-operator/controller/konnect/server"

	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

func TestServer(t *testing.T) {
	testCases := []struct {
		name                  string
		input                 string
		expectedURL           string
		expectedRegion        server.Region
		expectedErrorContains string
	}{
		{
			name:           "valid URL",
			input:          "https://us.konghq.com:8000",
			expectedURL:    "https://us.konghq.com:8000",
			expectedRegion: server.RegionUS,
		},
		{
			name:           "valid hostname",
			input:          "us.konghq.com",
			expectedURL:    "https://us.konghq.com",
			expectedRegion: server.RegionUS,
		},
		{
			name:                  "invalid URL",
			input:                 "not-a-valid-url:\\us.konghq.com",
			expectedErrorContains: "failed to parse region from hostname",
		},
		{
			name:                  "unknown region",
			input:                 "unknown.konghq.com",
			expectedErrorContains: "failed to parse region from hostname: unknown region",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := server.NewServer[konnectv1alpha1.KonnectGatewayControlPlane](tc.input)
			if tc.expectedErrorContains != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expectedErrorContains)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedURL, got.URL())
			assert.Equal(t, tc.expectedRegion, got.Region())
		})
	}

	t.Run("KonnectCloudGatewayNetwork", func(t *testing.T) {
		konnectTestCases := []struct {
			name           string
			input          string
			expectedURL    string
			expectedRegion server.Region
		}{
			{
				name:           "us",
				input:          "us.api.konghq.com",
				expectedURL:    "https://global.api.konghq.com",
				expectedRegion: server.RegionGlobal,
			},
			{
				name:           "eu",
				input:          "eu.api.konghq.com",
				expectedURL:    "https://global.api.konghq.com",
				expectedRegion: server.RegionGlobal,
			},
		}
		for _, tc := range konnectTestCases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := server.NewServer[konnectv1alpha1.KonnectCloudGatewayNetwork](tc.input)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedURL, got.URL())
				assert.Equal(t, tc.expectedRegion, got.Region())
			})
		}
	})
}
