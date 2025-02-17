package specialized

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kong/gateway-operator/api/v1alpha1"
	"github.com/kong/gateway-operator/controller/pkg/log"
	"github.com/kong/gateway-operator/pkg/utils/kubernetes"

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
)

// -----------------------------------------------------------------------------
// AIGatewayReconciler - Owned Resource Create/Update
// -----------------------------------------------------------------------------

func (r *AIGatewayReconciler) createOrUpdateHttpRoute(
	ctx context.Context,
	logger logr.Logger,
	httpRoute *gatewayv1.HTTPRoute,
) (bool, error) {
	log.Trace(logger, "checking for any existing httproute for aigateway")

	// TODO - use GenerateName
	//
	// See: https://github.com/Kong/gateway-operator/issues/137
	found := &gatewayv1.HTTPRoute{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      httpRoute.Name,
		Namespace: httpRoute.Namespace,
	}, found)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info(logger, "creating httproute for aigateway")
			return true, r.Client.Create(ctx, httpRoute)
		}
		return false, err
	}

	// TODO - implement patching
	//
	// See: https://github.com/Kong/gateway-operator/issues/137

	return false, nil
}

func (r *AIGatewayReconciler) createOrUpdatePlugin(
	ctx context.Context,
	logger logr.Logger,
	kongPlugin *configurationv1.KongPlugin,
) (bool, error) {
	log.Trace(logger, "checking for any existing plugin for aigateway")

	// TODO - use GenerateName
	//
	// See: https://github.com/Kong/gateway-operator/issues/137
	found := &configurationv1.KongPlugin{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      kongPlugin.Name,
		Namespace: kongPlugin.Namespace,
	}, found)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info(logger, "creating plugin for aigateway")
			return true, r.Client.Create(ctx, kongPlugin)
		}
		return false, err
	}

	// TODO - implement patching
	//
	// See: https://github.com/Kong/gateway-operator/issues/137

	return false, nil
}

func (r *AIGatewayReconciler) createOrUpdateGateway(
	ctx context.Context,
	logger logr.Logger,
	gateway *gatewayv1.Gateway,
) (bool, error) {
	log.Trace(logger, "checking for any existing gateway for aigateway")
	found := &gatewayv1.Gateway{}

	// TODO - use GenerateName
	//
	// See: https://github.com/Kong/gateway-operator/issues/137
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      gateway.Name,
		Namespace: gateway.Namespace,
	}, found)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info(logger, "creating gateway for aigateway")
			return true, r.Client.Create(ctx, gateway)
		}

		return false, err
	}

	// TODO - implement patching
	//
	// See: https://github.com/Kong/gateway-operator/issues/137

	return false, nil
}

func (r *AIGatewayReconciler) createOrUpdateSvc(
	ctx context.Context,
	logger logr.Logger,
	service *corev1.Service,
) (bool, error) {
	log.Trace(logger, "checking for any existing service for aigateway")

	// TODO - use GenerateName
	//
	// See: https://github.com/Kong/gateway-operator/issues/137
	found := &corev1.Service{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      service.Name,
		Namespace: service.Namespace,
	}, found)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info(logger, "creating service for aigateway")
			return true, r.Client.Create(ctx, service)
		}

		return false, err
	}

	// TODO - implement patching
	//
	// See: https://github.com/Kong/gateway-operator/issues/137

	return false, nil
}

// -----------------------------------------------------------------------------
// AIGatewayReconciler - Owned Resource Management
// -----------------------------------------------------------------------------

func (r *AIGatewayReconciler) manageGateway(
	ctx context.Context,
	logger logr.Logger,
	aiGateway *v1alpha1.AIGateway,
) (
	bool, // whether any changes were made
	error,
) {
	change, err := r.createOrUpdateGateway(ctx, logger, aiGatewayToGateway(aiGateway))
	if change {
		return true, err
	}
	if err != nil {
		return change, fmt.Errorf("could not reconcile Gateway, %w", err)
	}

	return change, nil
}

func (r *AIGatewayReconciler) configurePlugins(
	ctx context.Context,
	logger logr.Logger,
	aiGateway *v1alpha1.AIGateway,
) (
	bool, // whether any changes were made
	error,
) {
	changes := false

	log.Trace(logger, "configuring sink service for aigateway")
	aiGatewaySinkService := aiCloudGatewayToKubeSvc(aiGateway)
	changed, err := r.createOrUpdateSvc(ctx, logger, aiGatewaySinkService)
	if changed {
		changes = true
	}
	if err != nil {
		return changes, err
	}

	log.Trace(logger, "retrieving the cloud provider credentials secret for aigateway")
	if aiGateway.Spec.CloudProviderCredentials == nil {
		return changes, fmt.Errorf("ai gateway '%s' requires secret reference for Cloud Provider API keys", aiGateway.Name)
	}
	credentialSecretName := aiGateway.Spec.CloudProviderCredentials.Name
	credentialSecretNamespace := aiGateway.Namespace
	if aiGateway.Spec.CloudProviderCredentials.Namespace != nil {
		credentialSecretNamespace = *aiGateway.Spec.CloudProviderCredentials.Namespace
	}

	// check if referencing the credential secret is allowed by referencegrants.
	if !r.secretReferenceAllowedByReferenceGrants(ctx, logger, aiGateway, credentialSecretNamespace, credentialSecretName) {
		log.Info(logger, "Referencing Secret is not allowed by ReferenceGrants",
			"secret_namespace", credentialSecretNamespace, "secret_name", credentialSecretName)
		return false, fmt.Errorf("Referencing Secret %s/%s is not allowed by ReferenceGrants",
			credentialSecretNamespace, credentialSecretName)
	}

	credentialSecret := &corev1.Secret{}
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: credentialSecretNamespace, Name: credentialSecretName}, credentialSecret); err != nil {
		if k8serrors.IsNotFound(err) {
			return changes, nil
		}
		return changes, fmt.Errorf(
			"ai gateway '%s' references secret '%s/%s' but it could not be read, %w",
			aiGateway.Name, credentialSecretNamespace, credentialSecretName, err,
		)
	}

	log.Trace(logger, "generating routes and plugins for aigateway")
	for _, v := range aiGateway.Spec.LargeLanguageModels.CloudHosted {
		cloudHostedLLM := v

		log.Trace(logger, "determining whether we have API keys configured for cloud provider")
		credentialData, ok := credentialSecret.Data[string(cloudHostedLLM.AICloudProvider.Name)]
		if !ok {
			return changes, fmt.Errorf(
				"ai gateway '%s' references provider '%s' but it has no API key stored in the credentials secret",
				aiGateway.Name, string(cloudHostedLLM.AICloudProvider.Name),
			)
		}

		log.Trace(logger, "configuring the base aiproxy plugin for aigateway")
		aiProxyPlugin, err := aiCloudGatewayToKongPlugin(&cloudHostedLLM, aiGateway, &credentialData)
		if err != nil {
			return changes, err
		}
		changed, err := r.createOrUpdatePlugin(ctx, logger, aiProxyPlugin)
		if changed {
			changes = true
		}
		if err != nil {
			return changes, err
		}

		log.Trace(logger, "configuring the ai prompt decorator plugin for aigateway")
		decoratorPlugin, err := aiCloudGatewayToKongPromptDecoratorPlugin(&cloudHostedLLM, aiGateway)
		if err != nil {
			return changes, err
		}
		if decoratorPlugin != nil {
			changed, err := r.createOrUpdatePlugin(ctx, logger, decoratorPlugin)
			if changed {
				changes = true
			}
			if err != nil {
				return changes, err
			}
		}

		log.Trace(logger, "configuring an httproute for aigateway")
		plugins := []string{aiProxyPlugin.Name}
		if decoratorPlugin != nil {
			plugins = append(plugins, decoratorPlugin.Name)
		}
		httpRoute := aiCloudGatewayToHTTPRoute(&cloudHostedLLM, aiGateway, aiGatewaySinkService, plugins)
		changed, err = r.createOrUpdateHttpRoute(ctx, logger, httpRoute)
		if changed {
			changes = true
		}
		if err != nil {
			return changes, err
		}
	}

	return changes, nil
}

// secretReferenceAllowedByReferenceGrants returns true if the AIGateway is allowed to references the Secret `secretNamespace/secretName`.
// Returns true if they are in the same namespace, or there is any RefernceGrant allowing it.
func (r *AIGatewayReconciler) secretReferenceAllowedByReferenceGrants(
	ctx context.Context,
	logger logr.Logger,
	aigateway *v1alpha1.AIGateway,
	secretNamespace string,
	secretName string,
) bool {
	// Same namespace is always allowed.
	if aigateway.Namespace == secretNamespace {
		return true
	}

	// list referencegrants and check if the reference is allowed.
	from := gatewayv1beta1.ReferenceGrantFrom{
		Group:     gatewayv1beta1.Group(v1alpha1.AIGatewayGVR().Group),
		Kind:      gatewayv1beta1.Kind(aigateway.Kind),
		Namespace: gatewayv1beta1.Namespace(aigateway.Namespace),
	}
	to := gatewayv1beta1.ReferenceGrantTo{
		Group: gatewayv1.Group(corev1.GroupName),
		Kind:  gatewayv1.Kind("Secret"),
		Name:  lo.ToPtr(gatewayv1beta1.ObjectName(secretName)),
	}
	allowed, err := kubernetes.AllowedByReferenceGrants(ctx, r.Client, from, secretNamespace, to)
	if err != nil {
		log.Error(logger, err, "failed to check reference grant from aigateway to secret",
			"secret_namespace", secretNamespace,
			"secret_name", secretName,
		)
		return false
	}
	return allowed
}
