package specialized

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/kong/gateway-operator/api/v1alpha1"
	"github.com/kong/gateway-operator/controller/pkg/log"

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
)

// -----------------------------------------------------------------------------
// AIGatewayReconciler - Owned Resource Create/Update
// -----------------------------------------------------------------------------

func (r *AIGatewayReconciler) createOrUpdateHttpRoute(
	ctx context.Context,
	logger logr.Logger,
	aiGateway *v1alpha1.AIGateway,
	httpRoute *gatewayv1.HTTPRoute,
) (bool, error) {
	log.Trace(logger, "checking for any existing httproute for aigateway", aiGateway)

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
			log.Info(logger, "creating httproute for aigateway", aiGateway)
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
	aiGateway *v1alpha1.AIGateway,
	kongPlugin *configurationv1.KongPlugin,
) (bool, error) {
	log.Trace(logger, "checking for any existing plugin for aigateway", aiGateway)

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
			log.Info(logger, "creating plugin for aigateway", aiGateway)
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
	aiGateway *v1alpha1.AIGateway,
	gateway *gatewayv1.Gateway,
) (bool, error) {
	log.Trace(logger, "checking for any existing gateway for aigateway", aiGateway)
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
			log.Info(logger, "creating gateway for aigateway", aiGateway)
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
	aiGateway *v1alpha1.AIGateway,
	service *corev1.Service,
) (bool, error) {
	log.Trace(logger, "checking for any existing service for aigateway", aiGateway)

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
			log.Info(logger, "creating service for aigateway", aiGateway)
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
	change, err := r.createOrUpdateGateway(ctx, logger, aiGateway, aiGatewayToGateway(aiGateway))
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

	log.Trace(logger, "configuring sink service for aigateway", aiGateway)
	aiGatewaySinkService := aiCloudGatewayToKubeSvc(aiGateway)
	changed, err := r.createOrUpdateSvc(ctx, logger, aiGateway, aiGatewaySinkService)
	if changed {
		changes = true
	}
	if err != nil {
		return changes, err
	}

	log.Trace(logger, "retrieving the cloud provider credentials secret for aigateway", aiGateway)
	if aiGateway.Spec.CloudProviderCredentials == nil {
		return changes, fmt.Errorf("ai gateway '%s' requires secret reference for Cloud Provider API keys", aiGateway.Name)
	}
	credentialSecretName := aiGateway.Spec.CloudProviderCredentials.Name
	credentialSecretNamespace := aiGateway.Namespace
	if aiGateway.Spec.CloudProviderCredentials.Namespace != nil {
		credentialSecretNamespace = *aiGateway.Spec.CloudProviderCredentials.Namespace
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

	log.Trace(logger, "generating routes and plugins for aigateway", aiGateway)
	for _, v := range aiGateway.Spec.LargeLanguageModels.CloudHosted {
		cloudHostedLLM := v

		log.Trace(logger, "determining whether we have API keys configured for cloud provider", aiGateway)
		credentialData, ok := credentialSecret.Data[string(cloudHostedLLM.AICloudProvider.Name)]
		if !ok {
			return changes, fmt.Errorf(
				"ai gateway '%s' references provider '%s' but it has no API key stored in the credentials secret",
				aiGateway.Name, string(cloudHostedLLM.AICloudProvider.Name),
			)
		}

		log.Trace(logger, "configuring the base aiproxy plugin for aigateway", aiGateway)
		aiProxyPlugin, err := aiCloudGatewayToKongPlugin(&cloudHostedLLM, aiGateway, &credentialData)
		if err != nil {
			return changes, err
		}
		changed, err := r.createOrUpdatePlugin(ctx, logger, aiGateway, aiProxyPlugin)
		if changed {
			changes = true
		}
		if err != nil {
			return changes, err
		}

		log.Trace(logger, "configuring the ai prompt decorator plugin for aigateway", aiGateway)
		decoratorPlugin, err := aiCloudGatewayToKongPromptDecoratorPlugin(&cloudHostedLLM, aiGateway)
		if err != nil {
			return changes, err
		}
		if decoratorPlugin != nil {
			changed, err := r.createOrUpdatePlugin(ctx, logger, aiGateway, decoratorPlugin)
			if changed {
				changes = true
			}
			if err != nil {
				return changes, err
			}
		}

		log.Trace(logger, "configuring an httproute for aigateway", aiGateway)
		plugins := []string{aiProxyPlugin.Name}
		if decoratorPlugin != nil {
			plugins = append(plugins, decoratorPlugin.Name)
		}
		httpRoute := aiCloudGatewayToHTTPRoute(&cloudHostedLLM, aiGateway, aiGatewaySinkService, plugins)
		changed, err = r.createOrUpdateHttpRoute(ctx, logger, aiGateway, httpRoute)
		if changed {
			changes = true
		}
		if err != nil {
			return changes, err
		}
	}

	return changes, nil
}
