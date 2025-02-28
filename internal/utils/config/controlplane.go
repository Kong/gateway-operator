package config

import (
	"fmt"
	"strings"

	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

// KongInKonnectDefaults returns the map of Konnect-related env vars properly configured.
func KICInKonnectDefaults(konnectExtensionStatus konnectv1alpha1.KonnectExtensionStatus) (map[string]string, error) {
	newEnvSet := make(map[string]string, len(kongInKonnectClusterTypeControlPlaneDefaults))
	var template map[string]string

	switch konnectExtensionStatus.Konnect.ClusterType {
	case konnectv1alpha1.ClusterTypeK8sIngressController:
		template = map[string]string{
			"CONTROLLER_KONNECT_SYNC_ENABLED": "true",
		}
	case konnectv1alpha1.ClusterTypeControlPlane:
		return nil, fmt.Errorf("unsupported Konnect cluster type: %s", konnectExtensionStatus.Konnect.ClusterType)
	default:
		// default never happens as the validation is at the CRD level
		panic(fmt.Sprintf("unsupported Konnect cluster type: %s", konnectExtensionStatus.Konnect.ClusterType))
	}

	for k, v := range template {
		v = strings.ReplaceAll(v, "<CONTROL-PLANE-ENDPOINT>", sanitizeEndpoint(konnectExtensionStatus.Konnect.Endpoints.ControlPlaneEndpoint))
		v = strings.ReplaceAll(v, "<TELEMETRY-ENDPOINT>", sanitizeEndpoint(konnectExtensionStatus.Konnect.Endpoints.TelemetryEndpoint))
		newEnvSet[k] = v
	}

	return newEnvSet, nil
}
