package translator

import (
	"errors"
	"fmt"

	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/kong/kong-operator/ingress-controller/internal/dataplane/translator/subtranslator"
	"github.com/kong/kong-operator/ingress-controller/internal/gatewayapi"
)

// -----------------------------------------------------------------------------
// Translate HTTPRoute - IngressRules Translation
// -----------------------------------------------------------------------------

// ingressRulesFromHTTPRoutes processes a list of HTTPRoute objects and translates
// then into Kong configuration objects.
func (t *Translator) ingressRulesFromHTTPRoutes() ingressRules {
	result := newIngressRules()

	httpRouteList, err := t.storer.ListHTTPRoutes()
	if err != nil {
		t.logger.Error(err, "Failed to list HTTPRoutes")
		return result
	}

	httpRoutesToTranslate := make([]*gatewayapi.HTTPRoute, 0, len(httpRouteList))
	for _, httproute := range httpRouteList {
		// Validate each HTTPRoute before translating and register translation failures if an HTTPRoute is invalid.
		if err := validateHTTPRoute(httproute, t.featureFlags); err != nil {
			t.registerTranslationFailure(fmt.Sprintf("HTTPRoute can't be routed: %v", err), httproute)
			continue
		}
		httpRoutesToTranslate = append(httpRoutesToTranslate, httproute)
	}

	t.ingressRulesFromHTTPRoutesWithCombinedService(httpRoutesToTranslate, &result)
	return result
}

func validateHTTPRoute(httproute *gatewayapi.HTTPRoute, featureFlags FeatureFlags) error {
	spec := httproute.Spec

	// validation for HTTPRoutes will happen at a higher layer, but in spite of that we run
	// validation at this level as well as a fallback so that if routes are posted which
	// are invalid somehow make it past validation (e.g. the webhook is not enabled) we can
	// at least try to provide a helpful message about the situation in the manager logs.
	if len(spec.Rules) == 0 {
		return subtranslator.ErrRouteValidationNoRules
	}

	// Kong supports query parameter match only with expression router,
	// so we return error when query param match is specified and expression router is not enabled in the translator.
	if !featureFlags.ExpressionRoutes {
		for _, rule := range spec.Rules {
			for _, match := range rule.Matches {
				if len(match.QueryParams) > 0 {
					return subtranslator.ErrRouteValidationQueryParamMatchesUnsupported
				}
			}
		}
	}

	return nil
}

// ingressRulesFromHTTPRoutesWithCombinedService translates a list of HTTPRoutes to ingress rules.
// When the feature flag CombinedServicesFromDifferentHTTPRoutes is true, it combines rules with same backends
// to a single Kong gateway service across different HTTPRoutes in the same namespace.
// When the feature flag is false, it combines rules with same backends in an HTTPRoute to a Kong gateway service.
// When the feature flag ExpressionRoutes is set to true, expression based Kong routes will be translated from matches of HTTPRoutes.
// Otherwise, traditional Kong routes are translated.
func (t *Translator) ingressRulesFromHTTPRoutesWithCombinedService(httpRoutes []*gatewayapi.HTTPRoute, result *ingressRules) {
	translateOptions := subtranslator.TranslateHTTPRouteToKongstateServiceOptions{
		CombinedServicesFromDifferentHTTPRoutes: t.featureFlags.CombinedServicesFromDifferentHTTPRoutes,
		ExpressionRoutes:                        t.featureFlags.ExpressionRoutes,
		SupportRedirectPlugin:                   t.featureFlags.SupportRedirectPlugin,
	}
	translationResult := subtranslator.TranslateHTTPRoutesToKongstateServices(
		t.logger,
		t.storer,
		httpRoutes,
		translateOptions,
	)
	for serviceName, service := range translationResult.ServiceNameToKongstateService {
		result.ServiceNameToServices[serviceName] = service
		result.ServiceNameToParent[serviceName] = service.Parent
	}
	for _, httproute := range httpRoutes {
		namespacedName := k8stypes.NamespacedName{
			Namespace: httproute.Namespace,
			Name:      httproute.Name,
		}
		translationFailures := translationResult.HTTPRouteNameToTranslationErrors[namespacedName]
		// For HTTPRoutes that errors happened in translation, register translation failures for them.
		if len(translationFailures) > 0 {
			t.failuresCollector.PushResourceFailure(
				fmt.Sprintf("HTTPRoute can't be routed: %v", errors.Join(translationFailures...)),
				httproute,
			)
			continue
		}
		// Register HTTPRoute successfully translated if no translation error found.
		t.registerSuccessfullyTranslatedObject(httproute)
	}
}
