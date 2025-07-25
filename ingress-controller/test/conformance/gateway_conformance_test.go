//go:build conformance_tests

package conformance

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/gateway-api/conformance"
	conformancev1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
	"sigs.k8s.io/yaml"

	dpconf "github.com/kong/kong-operator/ingress-controller/internal/dataplane/config"
	"github.com/kong/kong-operator/ingress-controller/pkg/metadata"
	"github.com/kong/kong-operator/ingress-controller/test/internal/testenv"
)

var skippedTestsForTraditionalRoutes = []string{
	// core conformance
	tests.HTTPRouteHeaderMatching.ShortName,
	// NOTE: Skipped tests.GRPCRouteHeaderMatching.ShortName and
	// tests.GRPCExactMethodMatching.ShortName because in traditional mode,
	// when wanting to proxy different gRPC services and route requests based on Header or Method,
	// it is necessary to create separate catch-all routes for them.
	// However, Kong does not define priority behavior in this situation unless priorities are manually added.
	// ref: https://github.com/Kong/kubernetes-ingress-controller/issues/6144
	tests.GRPCRouteHeaderMatching.ShortName,
	tests.GRPCExactMethodMatching.ShortName,
}

var skippedTestsForExpressionRoutes = []string{
	// When processing this scenario, the Kong's expressions router requires `priority`
	// to be specified for routes.
	// We cannot provide that for routes that are part of the conformance suite.
	tests.GRPCRouteListenerHostnameMatching.ShortName,
}

var traditionalRoutesSupportedFeatures = []features.FeatureName{
	// core features
	features.SupportGateway,
	features.SupportHTTPRoute,
	features.SupportGRPCRoute,
	// extended features
	features.SupportHTTPRouteResponseHeaderModification,
	features.SupportHTTPRoutePathRewrite,
	features.SupportHTTPRouteHostRewrite,
	// TODO: https://github.com/Kong/kubernetes-ingress-controller/issues/5868
	// Temporarily disabled and tracking through the following issue.
	// suite.SupportHTTPRouteBackendTimeout,
}

var expressionRoutesSupportedFeatures = []features.FeatureName{
	// core features
	features.SupportGateway,
	features.SupportHTTPRoute,
	features.SupportGRPCRoute,
	// extended features
	features.SupportHTTPRouteQueryParamMatching,
	features.SupportHTTPRouteMethodMatching,
	features.SupportHTTPRouteResponseHeaderModification,
	features.SupportHTTPRoutePathRewrite,
	features.SupportHTTPRouteHostRewrite,
	// TODO: https://github.com/Kong/kubernetes-ingress-controller/issues/5868
	// Temporarily disabled and tracking through the following issue.
	// features.SupportHTTPRouteBackendTimeout,
}

func TestGatewayConformance(t *testing.T) {
	k8sClient, gatewayClassName := prepareEnvForGatewayConformanceTests(t)

	// Conformance tests are run for both available router flavours:
	// traditional_compatible and expressions.
	var (
		skippedTests      []string
		supportedFeatures []features.FeatureName
		mode              string
	)
	switch rf := testenv.KongRouterFlavor(); rf {
	case dpconf.RouterFlavorTraditionalCompatible:
		skippedTests = skippedTestsForTraditionalRoutes
		supportedFeatures = traditionalRoutesSupportedFeatures
		mode = string(dpconf.RouterFlavorTraditionalCompatible)
	case dpconf.RouterFlavorExpressions:
		skippedTests = skippedTestsForExpressionRoutes
		supportedFeatures = expressionRoutesSupportedFeatures
		mode = string(dpconf.RouterFlavorExpressions)
	default:
		t.Fatalf("unsupported KongRouterFlavor: %s", rf)
	}

	opts := conformance.DefaultOptions(t)
	opts.GatewayClassName = gatewayClassName
	opts.Debug = true
	opts.Mode = mode
	opts.CleanupBaseResources = !testenv.IsCI()
	opts.BaseManifests = conformanceTestsBaseManifests
	opts.SupportedFeatures = sets.New(supportedFeatures...)
	opts.SkipTests = skippedTests
	opts.ConformanceProfiles = sets.New(
		suite.GatewayHTTPConformanceProfileName,
		suite.GatewayGRPCConformanceProfileName,
	)
	opts.Implementation = conformancev1.Implementation{
		Organization: metadata.Organization,
		Project:      metadata.ProjectName,
		URL:          metadata.ProjectURL,
		Version:      metadata.Release,
		Contact: []string{
			path.Join(metadata.ProjectURL, "/issues/new/choose"),
		},
	}
	cSuite, err := suite.NewConformanceTestSuite(opts)
	require.NoError(t, err)

	t.Log("starting the gateway conformance test suite")
	cSuite.Setup(t, tests.ConformanceTests)

	go patchGatewayClassToPassTestGatewayClassObservedGenerationBump(ctx, t, k8sClient)

	// To work with individual tests only, you can disable the normal Run call and construct a slice containing a
	// single test only, e.g.:
	//
	// cSuite.Run(t, []suite.ConformanceTest{tests.GatewayClassObservedGenerationBump})
	// To work with individual tests only, you can disable the normal Run call and construct a slice containing a
	// single test only, e.g.:
	//
	// require.NoError(t, cSuite.Run(t, []suite.ConformanceTest{tests.HTTPRouteRewritePath}))
	require.NoError(t, cSuite.Run(t, tests.ConformanceTests))

	const reportFileName = "kong-kubernetes-ingress-controller.yaml"
	t.Log("saving the gateway conformance test report to file:", reportFileName)
	report, err := cSuite.Report()
	require.NoError(t, err)
	rawReport, err := yaml.Marshal(report)
	require.NoError(t, err)
	// Save report in root of the repository, file name is in .gitignore.
	require.NoError(t, os.WriteFile("../../"+reportFileName, rawReport, 0o600))
}
