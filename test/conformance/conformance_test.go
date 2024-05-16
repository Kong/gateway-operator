package conformance

import (
	"flag"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	conformancev1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/kong/gateway-operator/internal/metadata"
	testutils "github.com/kong/gateway-operator/pkg/utils/test"
	"github.com/kong/gateway-operator/pkg/vars"
)

var skippedTests = []string{
	// gateway
	tests.GatewayInvalidTLSConfiguration.ShortName,
	tests.GatewayModifyListeners.ShortName,
	tests.GatewayWithAttachedRoutes.ShortName,

	// httproute
	tests.HTTPRouteHeaderMatching.ShortName,
}

var (
	shouldCleanup = flag.Bool("cleanup", true, "indicates whether or not the base test resources such as Gateways should be cleaned up after the run.")
	showDebug     = flag.Bool("debug", false, "indicates whether to execute the conformance tests in debug mode.")
)

func TestGatewayConformance(t *testing.T) {
	t.Parallel()

	t.Log("creating GatewayClass for gateway conformance tests")
	gwc := &gatewayv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: uuid.NewString(),
		},
		Spec: gatewayv1.GatewayClassSpec{
			ControllerName: gatewayv1.GatewayController(vars.ControllerName()),
		},
	}
	require.NoError(t, clients.MgrClient.Create(ctx, gwc))
	defer func() {
		require.NoError(t, clients.MgrClient.Delete(ctx, gwc))
	}()

	// There are no explicit conformance tests for GatewayClass, but we can
	// still run the conformance test suite setup to ensure that the
	// GatewayClass gets accepted.
	t.Log("starting the gateway conformance test suite")
	conformanceTestsBaseManifests := fmt.Sprintf("%s/conformance/base/manifests.yaml", testutils.GatewayRawRepoURL)

	cSuite, err := suite.NewConformanceTestSuite(
		suite.ConformanceOptions{
			Client:               clients.MgrClient,
			GatewayClassName:     gwc.Name,
			Debug:                *showDebug,
			CleanupBaseResources: *shouldCleanup,
			BaseManifests:        conformanceTestsBaseManifests,
			SkipTests:            skippedTests,
			ConformanceProfiles: sets.New(
				suite.GatewayHTTPConformanceProfileName,
			),
			Implementation: conformancev1.Implementation{
				Organization: metadata.Organization,
				Project:      metadata.ProjectName,
				URL:          metadata.ProjectURL,
				Version:      metadata.Release,
				Contact: []string{
					path.Join(metadata.ProjectURL, "/issues/new/choose"),
				},
			},
		},
	)
	require.NoError(t, err)

	t.Log("starting the gateway conformance test suite")
	cSuite.Setup(t, nil)

	// To work with individual tests only, you can disable the normal Run call and construct a slice containing a
	// single test only, e.g.:
	//
	// cSuite.Run(t, []suite.ConformanceTest{tests.GatewayClassObservedGenerationBump})
	// To work with individual tests only, you can disable the normal Run call and construct a slice containing a
	// single test only, e.g.:
	//
	// cSuite.Run(t, []suite.ConformanceTest{tests.HTTPRouteRedirectPortAndScheme})
	require.NoError(t, cSuite.Run(t, tests.ConformanceTests))

	// In the future we'll likely change the file name as https://github.com/kubernetes-sigs/gateway-api/issues/2740 will be implemented.
	// The day it will happen, we'll have to change the .gitignore entry as well.
	const reportFileName = "kong-gateway-operator.yaml"

	t.Log("saving the gateway conformance test report to file:", reportFileName)
	report, err := cSuite.Report()
	require.NoError(t, err)
	rawReport, err := yaml.Marshal(report)
	require.NoError(t, err)
	fmt.Println("INFO: final report")
	fmt.Println(string(rawReport))
	// Save report in root of the repository, file name is in .gitignore.
	require.NoError(t, os.WriteFile("../../"+reportFileName, rawReport, 0o600))
}
