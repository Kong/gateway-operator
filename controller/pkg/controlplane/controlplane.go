package controlplane

import (
	"fmt"
	"os"
	"reflect"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"

	operatorv1beta1 "github.com/kong/gateway-operator/api/v1beta1"
	cputils "github.com/kong/gateway-operator/internal/utils/controlplane"
	"github.com/kong/gateway-operator/internal/versions"
	"github.com/kong/gateway-operator/pkg/consts"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
	k8scompare "github.com/kong/gateway-operator/pkg/utils/kubernetes/compare"
	"github.com/kong/gateway-operator/pkg/vars"
)

// DefaultsArgs contains the parameters to pass to setControlPlaneDefaults
type DefaultsArgs struct {
	Namespace                   string
	ControlPlaneName            string
	DataPlaneIngressServiceName string
	DataPlaneAdminServiceName   string
	OwnedByGateway              string
	AnonymousReportsEnabled     bool
	Konnect                     bool
}

// -----------------------------------------------------------------------------
// ControlPlane - Private Functions
// -----------------------------------------------------------------------------

// SetDefaults updates the environment variables of control plane
// and returns true if env field is changed.
func SetDefaults(
	spec *operatorv1beta1.ControlPlaneOptions,
	args DefaultsArgs,
) bool {
	changed := false

	// set env POD_NAMESPACE. should be always from `metadata.namespace` of pod.
	envSourceMetadataNamespace := &corev1.EnvVarSource{
		FieldRef: &corev1.ObjectFieldSelector{
			APIVersion: "v1",
			FieldPath:  "metadata.namespace",
		},
	}
	if spec.Deployment.PodTemplateSpec == nil {
		spec.Deployment.PodTemplateSpec = &corev1.PodTemplateSpec{}
	}

	podSpec := &spec.Deployment.PodTemplateSpec.Spec
	container := k8sutils.GetPodContainerByName(podSpec, consts.ControlPlaneControllerContainerName)
	if container == nil {
		container = &corev1.Container{
			Name: consts.ControlPlaneControllerContainerName,
		}
	}
	dontOverride := make(map[string]struct{})
	for _, envVar := range container.Env {
		dontOverride[envVar.Name] = struct{}{}
	}

	const podNamespaceEnvVarName = cputils.PodNamespaceEnvVarName
	if !reflect.DeepEqual(envSourceMetadataNamespace, k8sutils.EnvVarSourceByName(container.Env, podNamespaceEnvVarName)) {
		container.Env = k8sutils.UpdateEnvSource(container.Env, podNamespaceEnvVarName, envSourceMetadataNamespace)
		changed = true
	}

	// due to the anonymous reports being enabled by default
	// if the flag is set to false, we need to set the env var to false
	if k8sutils.EnvValueByName(container.Env, cputils.ControllerAnonymousReportsEnvVarName) != fmt.Sprintf("%t", args.AnonymousReportsEnabled) {
		container.Env = k8sutils.UpdateEnv(container.Env, cputils.ControllerAnonymousReportsEnvVarName, fmt.Sprintf("%t", args.AnonymousReportsEnabled))
		changed = true
	}

	// set env POD_NAME. should be always from `metadata.name` of pod.
	envSourceMetadataName := &corev1.EnvVarSource{
		FieldRef: &corev1.ObjectFieldSelector{
			APIVersion: "v1",
			FieldPath:  "metadata.name",
		},
	}
	if !reflect.DeepEqual(envSourceMetadataName, k8sutils.EnvVarSourceByName(container.Env, cputils.PodNameEnvVarName)) {
		container.Env = k8sutils.UpdateEnvSource(container.Env, cputils.PodNameEnvVarName, envSourceMetadataName)
		changed = true
	}

	if ctrlName := vars.ControllerName(); k8sutils.EnvValueByName(container.Env, cputils.ControllerGatewayAPIControllerNameEnvVarName) != ctrlName {
		container.Env = k8sutils.UpdateEnv(container.Env, cputils.ControllerGatewayAPIControllerNameEnvVarName, ctrlName)
		changed = true
	}

	if args.Namespace != "" && args.DataPlaneIngressServiceName != "" {
		if _, isOverrideDisabled := dontOverride[cputils.ControllerPublishServiceEnvVarName]; !isOverrideDisabled {
			publishServiceNN := k8stypes.NamespacedName{Namespace: args.Namespace, Name: args.DataPlaneIngressServiceName}.String()
			if k8sutils.EnvValueByName(container.Env, cputils.ControllerPublishServiceEnvVarName) != publishServiceNN {
				container.Env = k8sutils.UpdateEnv(container.Env, cputils.ControllerPublishServiceEnvVarName, publishServiceNN)
				changed = true
			}
		}
	}

	if args.Namespace != "" && args.DataPlaneAdminServiceName != "" {
		dataPlaneAdminServiceNN := k8stypes.NamespacedName{Namespace: args.Namespace, Name: args.DataPlaneAdminServiceName}.String()
		if _, isOverrideDisabled := dontOverride[cputils.ControllerKongAdminSVCEnvVarName]; !isOverrideDisabled {
			if k8sutils.EnvValueByName(container.Env, cputils.ControllerKongAdminSVCEnvVarName) != dataPlaneAdminServiceNN {
				container.Env = k8sutils.UpdateEnv(container.Env, cputils.ControllerKongAdminSVCEnvVarName, dataPlaneAdminServiceNN)
				changed = true
			}
		}

		if _, isOverrideDisabled := dontOverride[cputils.ControllerKongAdminSVCPortNamesEnvVarName]; !isOverrideDisabled {
			if k8sutils.EnvValueByName(container.Env, cputils.ControllerKongAdminSVCPortNamesEnvVarName) != consts.DataPlaneAdminServicePortName {
				container.Env = k8sutils.UpdateEnv(container.Env, cputils.ControllerKongAdminSVCPortNamesEnvVarName, consts.DataPlaneAdminServicePortName)
				changed = true
			}
		}
	}
	if _, isOverrideDisabled := dontOverride[cputils.ControllerGatewayDiscoveryDNSStrategyEnvVarName]; !isOverrideDisabled {
		if k8sutils.EnvValueByName(container.Env, cputils.ControllerGatewayDiscoveryDNSStrategyEnvVarName) != consts.DataPlaneServiceDNSDiscoveryStrategy {
			container.Env = k8sutils.UpdateEnv(container.Env, cputils.ControllerGatewayDiscoveryDNSStrategyEnvVarName, consts.DataPlaneServiceDNSDiscoveryStrategy)
			changed = true
		}
	}

	if args.OwnedByGateway != "" {
		// If the controlplane is managed by a gateway, the controlplane may take some time to properly connect to the dataplane,
		// as the controlplane and the dataplane are deployed together. For this reason, we set the env var CONTROLLER_KONG_ADMIN_INIT_RETRY_DELAY
		// to 5s (the default value is 1s) to:
		// - reduce spamming of "retrying connection to the dataplane i/60";
		// - avoid crash of the controlplane pod when the dataplane is particularly slow to start (it happens quite rarely).
		if _, isOverrideDisabled := dontOverride[cputils.ControllerKongAdminInitRetryDelayEnvVarName]; !isOverrideDisabled {
			if k8sutils.EnvValueByName(container.Env, cputils.ControllerKongAdminInitRetryDelayEnvVarName) != consts.DataPlaneInitRetryDelay {
				container.Env = k8sutils.UpdateEnv(container.Env, cputils.ControllerKongAdminInitRetryDelayEnvVarName, consts.DataPlaneInitRetryDelay)
				changed = true
			}
		}

		if _, isOverrideDisabled := dontOverride[cputils.ControllerGatewayToReconcileEnvVarName]; !isOverrideDisabled {
			gatewayOwner := fmt.Sprintf("%s/%s", args.Namespace, args.OwnedByGateway)
			if k8sutils.EnvValueByName(container.Env, cputils.ControllerGatewayToReconcileEnvVarName) != gatewayOwner {
				container.Env = k8sutils.UpdateEnv(container.Env, cputils.ControllerGatewayToReconcileEnvVarName, gatewayOwner)
				changed = true
			}
		}
	}
	// This uses a different check for ownership. this function gets invoked twice for gateway-managed ControlPlanes,
	// once from the Gateway controller, which preps its own copy of the ControlPlane config before spawning a ControlPlane,
	// and once from the ControlPlane controller. the Gateway controller only has the spec and lacks meta, whereas the
	// ControlPlane controller doesn't have the args.ManagedByGateway

	if _, isOverrideDisabled := dontOverride[cputils.ControllerKongAdminTLSClientCertFileEnvVarName]; !isOverrideDisabled {
		if k8sutils.EnvValueByName(container.Env, cputils.ControllerKongAdminTLSClientCertFileEnvVarName) != consts.TLSCRTPath {
			container.Env = k8sutils.UpdateEnv(container.Env, cputils.ControllerKongAdminTLSClientCertFileEnvVarName, consts.TLSCRTPath)
			changed = true
		}
	}
	if _, isOverrideDisabled := dontOverride[cputils.ControllerKongAdminTLSClientKeyFileEnvVarName]; !isOverrideDisabled {
		if k8sutils.EnvValueByName(container.Env, cputils.ControllerKongAdminTLSClientKeyFileEnvVarName) != consts.TLSKeyPath {
			container.Env = k8sutils.UpdateEnv(container.Env, cputils.ControllerKongAdminTLSClientKeyFileEnvVarName, consts.TLSKeyPath)
			changed = true
		}
	}
	if _, isOverrideDisabled := dontOverride[cputils.ControllerKongAdminCACertFileEnvVarName]; !isOverrideDisabled {
		if k8sutils.EnvValueByName(container.Env, cputils.ControllerKongAdminCACertFileEnvVarName) != consts.TLSCACRTPath {
			container.Env = k8sutils.UpdateEnv(container.Env, cputils.ControllerKongAdminCACertFileEnvVarName, consts.TLSCACRTPath)
			changed = true
		}
	}

	if args.ControlPlaneName != "" {
		electionID := fmt.Sprintf("%s.konghq.com", args.ControlPlaneName)
		if _, isOverrideDisabled := dontOverride[cputils.ControllerElectionIDEnvVarName]; !isOverrideDisabled {
			if k8sutils.EnvValueByName(container.Env, cputils.ControllerElectionIDEnvVarName) != electionID {
				container.Env = k8sutils.UpdateEnv(container.Env, cputils.ControllerElectionIDEnvVarName, electionID)
				changed = true
			}
		}
	}

	if _, isOverrideDisabled := dontOverride[cputils.ControllerAdmissionWebhookListenEnvVarName]; !isOverrideDisabled {
		if k8sutils.EnvValueByName(container.Env, cputils.ControllerAdmissionWebhookListenEnvVarName) != consts.ControlPlaneAdmissionWebhookEnvVarValue {
			container.Env = k8sutils.UpdateEnv(container.Env, cputils.ControllerAdmissionWebhookListenEnvVarName, consts.ControlPlaneAdmissionWebhookEnvVarValue)
			changed = true
		}
	}

	k8sutils.SetPodContainer(podSpec, container)

	return changed
}

// GenerateImage returns the image to use for the control plane.
func GenerateImage(opts *operatorv1beta1.ControlPlaneOptions, validators ...versions.VersionValidationOption) (string, error) {
	container := k8sutils.GetPodContainerByName(&opts.Deployment.PodTemplateSpec.Spec, consts.ControlPlaneControllerContainerName)
	if container == nil {
		// This is just a safeguard against running the operator without an admission webhook
		// (which would prevent admission of a ControlPlane without an image specified)
		// to prevent panics.
		return "", fmt.Errorf("unsupported ControlPlane without image")
	}
	if container.Image != "" {
		for _, v := range validators {
			supported, err := v(container.Image)
			if err != nil {
				return "", err
			}
			if !supported {
				return "", fmt.Errorf("unsupported ControlPlane image %s", container.Image)
			}
		}
		return container.Image, nil
	}

	if relatedKongControllerImage := os.Getenv("RELATED_IMAGE_KONG_CONTROLLER"); relatedKongControllerImage != "" {
		// RELATED_IMAGE_KONG_CONTROLLER is set by the operator-sdk when building the operator bundle.
		// https://github.com/Kong/gateway-operator-archive/issues/261
		return relatedKongControllerImage, nil
	}

	return consts.DefaultControlPlaneImage, nil // TODO: https://github.com/Kong/gateway-operator-archive/issues/20
}

// -----------------------------------------------------------------------------
// ControlPlane - Private Functions - Equality Checks
// -----------------------------------------------------------------------------
func SpecDeepEqual(spec1, spec2 *operatorv1beta1.ControlPlaneOptions, envVarsToIgnore ...string) bool {
	if !k8scompare.ControlPlaneDeploymentOptionsDeepEqual(&spec1.Deployment, &spec2.Deployment, envVarsToIgnore...) ||
		!reflect.DeepEqual(spec1.DataPlane, spec2.DataPlane) {
		return false
	}

	if !reflect.DeepEqual(spec1.Extensions, spec2.Extensions) {
		return false
	}

	return true
}

// DeduceAnonymousReportsEnabled returns the value of the anonymous reports enabled
// based on the environment variable `CONTROLLER_ANONYMOUS_REPORTS` in the control plane
// pod template spec and operator development mode setting.
//
// This allows users to override the setting that is a derivative of the operator development mode
// using the environment variable `CONTROLLER_ANONYMOUS_REPORTS` in the control plane pod template spec.
func DeduceAnonymousReportsEnabled(developmentMode bool, cpOpts *operatorv1beta1.ControlPlaneOptions) bool {
	pts := cpOpts.Deployment.PodTemplateSpec
	if pts == nil {
		return !developmentMode
	}

	container := k8sutils.GetPodContainerByName(&pts.Spec, consts.ControlPlaneControllerContainerName)
	if container == nil {
		return !developmentMode
	}

	env := k8sutils.EnvValueByName(container.Env, "CONTROLLER_ANONYMOUS_REPORTS")
	if v, err := strconv.ParseBool(env); len(env) > 0 && err == nil {
		return v
	}

	return !developmentMode
}
