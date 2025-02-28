package mocks

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkops "github.com/kong/gateway-operator/controller/konnect/ops/sdk"
)

type MockSDKWrapper struct {
	ControlPlaneSDK             *MockControlPlaneSDK
	CloudGatewaysSDK            *MockCloudGatewaysSDK
	ControlPlaneGroupSDK        *MockControlPlaneGroupSDK
	ServicesSDK                 *MockServicesSDK
	RoutesSDK                   *MockRoutesSDK
	ConsumersSDK                *MockConsumersSDK
	ConsumerGroupSDK            *MockConsumerGroupSDK
	PluginSDK                   *MockPluginSDK
	UpstreamsSDK                *MockUpstreamsSDK
	TargetsSDK                  *MockTargetsSDK
	MeSDK                       *MockMeSDK
	KongCredentialsBasicAuthSDK *MockKongCredentialBasicAuthSDK
	KongCredentialsAPIKeySDK    *MockKongCredentialAPIKeySDK
	KongCredentialsACLSDK       *MockKongCredentialACLSDK
	KongCredentialsJWTSDK       *MockKongCredentialJWTSDK
	KongCredentialsHMACSDK      *MockKongCredentialHMACSDK
	CACertificatesSDK           *MockCACertificatesSDK
	CertificatesSDK             *MockCertificatesSDK
	VaultSDK                    *MockVaultSDK
	KeysSDK                     *MockKeysSDK
	KeySetsSDK                  *MockKeySetsSDK
	SNIsSDK                     *MockSNIsSDK
	DataPlaneCertificatesSDK    *MockDataPlaneClientCertificatesSDK
}

var _ sdkops.SDKWrapper = MockSDKWrapper{}

func NewMockSDKWrapperWithT(t *testing.T) *MockSDKWrapper {
	return &MockSDKWrapper{
		ControlPlaneSDK:             NewMockControlPlaneSDK(t),
		ControlPlaneGroupSDK:        NewMockControlPlaneGroupSDK(t),
		CloudGatewaysSDK:            NewMockCloudGatewaysSDK(t),
		ServicesSDK:                 NewMockServicesSDK(t),
		RoutesSDK:                   NewMockRoutesSDK(t),
		ConsumersSDK:                NewMockConsumersSDK(t),
		ConsumerGroupSDK:            NewMockConsumerGroupSDK(t),
		PluginSDK:                   NewMockPluginSDK(t),
		UpstreamsSDK:                NewMockUpstreamsSDK(t),
		TargetsSDK:                  NewMockTargetsSDK(t),
		MeSDK:                       NewMockMeSDK(t),
		KongCredentialsBasicAuthSDK: NewMockKongCredentialBasicAuthSDK(t),
		KongCredentialsAPIKeySDK:    NewMockKongCredentialAPIKeySDK(t),
		KongCredentialsACLSDK:       NewMockKongCredentialACLSDK(t),
		KongCredentialsJWTSDK:       NewMockKongCredentialJWTSDK(t),
		KongCredentialsHMACSDK:      NewMockKongCredentialHMACSDK(t),
		CACertificatesSDK:           NewMockCACertificatesSDK(t),
		CertificatesSDK:             NewMockCertificatesSDK(t),
		VaultSDK:                    NewMockVaultSDK(t),
		KeysSDK:                     NewMockKeysSDK(t),
		KeySetsSDK:                  NewMockKeySetsSDK(t),
		SNIsSDK:                     NewMockSNIsSDK(t),
		DataPlaneCertificatesSDK:    NewMockDataPlaneClientCertificatesSDK(t),
	}
}

const (
	mockSDKServerURL = "http://mock-api.konnect.test"
)

func (m MockSDKWrapper) GetServerURL() string {
	return mockSDKServerURL
}

func (m MockSDKWrapper) GetControlPlaneSDK() sdkops.ControlPlaneSDK {
	return m.ControlPlaneSDK
}

func (m MockSDKWrapper) GetControlPlaneGroupSDK() sdkops.ControlPlaneGroupSDK {
	return m.ControlPlaneGroupSDK
}

func (m MockSDKWrapper) GetServicesSDK() sdkops.ServicesSDK {
	return m.ServicesSDK
}

func (m MockSDKWrapper) GetRoutesSDK() sdkops.RoutesSDK {
	return m.RoutesSDK
}

func (m MockSDKWrapper) GetConsumersSDK() sdkops.ConsumersSDK {
	return m.ConsumersSDK
}

func (m MockSDKWrapper) GetConsumerGroupsSDK() sdkops.ConsumerGroupSDK {
	return m.ConsumerGroupSDK
}

func (m MockSDKWrapper) GetPluginSDK() sdkops.PluginSDK {
	return m.PluginSDK
}

func (m MockSDKWrapper) GetUpstreamsSDK() sdkops.UpstreamsSDK {
	return m.UpstreamsSDK
}

func (m MockSDKWrapper) GetBasicAuthCredentialsSDK() sdkops.KongCredentialBasicAuthSDK {
	return m.KongCredentialsBasicAuthSDK
}

func (m MockSDKWrapper) GetAPIKeyCredentialsSDK() sdkops.KongCredentialAPIKeySDK {
	return m.KongCredentialsAPIKeySDK
}

func (m MockSDKWrapper) GetACLCredentialsSDK() sdkops.KongCredentialACLSDK {
	return m.KongCredentialsACLSDK
}

func (m MockSDKWrapper) GetJWTCredentialsSDK() sdkops.KongCredentialJWTSDK {
	return m.KongCredentialsJWTSDK
}

func (m MockSDKWrapper) GetHMACCredentialsSDK() sdkops.KongCredentialHMACSDK {
	return m.KongCredentialsHMACSDK
}

func (m MockSDKWrapper) GetTargetsSDK() sdkops.TargetsSDK {
	return m.TargetsSDK
}

func (m MockSDKWrapper) GetVaultSDK() sdkops.VaultSDK {
	return m.VaultSDK
}

func (m MockSDKWrapper) GetMeSDK() sdkops.MeSDK {
	return m.MeSDK
}

func (m MockSDKWrapper) GetCACertificatesSDK() sdkops.CACertificatesSDK {
	return m.CACertificatesSDK
}

func (m MockSDKWrapper) GetCertificatesSDK() sdkops.CertificatesSDK {
	return m.CertificatesSDK
}

func (m MockSDKWrapper) GetKeysSDK() sdkops.KeysSDK {
	return m.KeysSDK
}

func (m MockSDKWrapper) GetKeySetsSDK() sdkops.KeySetsSDK {
	return m.KeySetsSDK
}

func (m MockSDKWrapper) GetSNIsSDK() sdkops.SNIsSDK {
	return m.SNIsSDK
}

func (m MockSDKWrapper) GetDataPlaneCertificatesSDK() sdkops.DataPlaneClientCertificatesSDK {
	return m.DataPlaneCertificatesSDK
}

func (m MockSDKWrapper) GetCloudGatewaysSDK() sdkops.CloudGatewaysSDK {
	return m.CloudGatewaysSDK
}

type MockSDKFactory struct {
	t   *testing.T
	SDK *MockSDKWrapper
}

var _ sdkops.SDKFactory = MockSDKFactory{}

func NewMockSDKFactory(t *testing.T) *MockSDKFactory {
	return &MockSDKFactory{
		t:   t,
		SDK: NewMockSDKWrapperWithT(t),
	}
}

func (m MockSDKFactory) NewKonnectSDK(_ string, _ sdkops.SDKToken) sdkops.SDKWrapper {
	require.NotNil(m.t, m.SDK)
	return *m.SDK
}
