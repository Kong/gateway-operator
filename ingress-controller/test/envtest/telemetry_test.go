//go:build envtest

package envtest

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilversion "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/apimachinery/pkg/version"
	discoveryclient "k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	gatewayclient "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"

	"github.com/kong/kong-operator/ingress-controller/internal/gatewayapi"
	"github.com/kong/kong-operator/ingress-controller/internal/k8s"
	"github.com/kong/kong-operator/ingress-controller/pkg/manager"
)

func TestTelemetry(t *testing.T) {
	t.Parallel()

	ts := NewTelemetryServer(t)
	defer ts.Stop(t)
	ts.Start(t.Context(), t)

	t.Log("configuring envtest and creating K8s objects for telemetry test")
	envcfg := Setup(t, scheme.Scheme)
	// Let's have long duration due too rate limiter in K8s client.
	cfg, err := manager.NewConfig(
		WithDefaultEnvTestsConfig(envcfg),
	)
	require.NoError(t, err)
	c, err := k8s.GetKubeconfig(cfg)
	require.NoError(t, err)
	createK8sObjectsForTelemetryTest(t.Context(), t, c)

	t.Log("starting the controller manager")
	go func() {
		RunManager(t.Context(), t, envcfg, AdminAPIOptFns(),
			WithDefaultEnvTestsConfig(envcfg),
			WithTelemetry(ts.Endpoint(), 100*time.Millisecond),
		)
	}()

	dcl, err := discoveryclient.NewDiscoveryClientForConfig(envcfg)
	require.NoError(t, err)
	k8sVersion, err := dcl.ServerVersion()
	require.NoError(t, err)

	t.Log("verifying that eventually we get an expected telemetry report")
	const (
		waitTime = 3 * time.Second
		tickTime = 10 * time.Millisecond
	)
	require.Eventuallyf(t, func() bool {
		select {
		case report := <-ts.ReportChan():
			return verifyTelemetryReport(t, k8sVersion, string(report))
		case <-time.After(tickTime):
			return false
		}
	}, waitTime, tickTime, "telemetry report never matched expected value")
}

func createK8sObjectsForTelemetryTest(ctx context.Context, t *testing.T, cfg *rest.Config) {
	t.Helper()
	cl, err := kubernetes.NewForConfig(cfg)
	require.NoError(t, err)

	gcl, err := gatewayclient.NewForConfig(cfg)
	require.NoError(t, err)

	const additionalNamespace = "test-ns-1"
	_, err = cl.CoreV1().Namespaces().Create(
		ctx,
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: additionalNamespace},
		},
		metav1.CreateOptions{},
	)
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		_, err = gcl.GatewayV1().GatewayClasses().Create(
			ctx,
			&gatewayapi.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("test-%d", i)},
				Spec: gatewayapi.GatewayClassSpec{
					ControllerName: "test.com/gateway-controller",
				},
			},
			metav1.CreateOptions{},
		)
		require.NoError(t, err)

		_, err = cl.CoreV1().Nodes().Create(
			ctx,
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("test-%d", i)},
			},
			metav1.CreateOptions{},
		)
		require.NoError(t, err)

		for _, namespace := range []string{metav1.NamespaceDefault, additionalNamespace} {
			_, err := cl.CoreV1().Services(namespace).Create(
				ctx,
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("test-%d", i)},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Port: 443,
							},
						},
					},
				},
				metav1.CreateOptions{},
			)
			require.NoError(t, err)

			_, err = cl.CoreV1().Pods(namespace).Create(
				ctx,
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("test-%d", i)},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "test", Image: "test"},
						},
					},
				},
				metav1.CreateOptions{},
			)
			require.NoError(t, err)

			_, err = cl.NetworkingV1().Ingresses(namespace).Create(
				ctx,
				&netv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("test-%d", i)},
					Spec: netv1.IngressSpec{
						Rules: []netv1.IngressRule{
							{
								Host: fmt.Sprintf("test-%d.example.com", i),
							},
						},
					},
				},
				metav1.CreateOptions{},
			)
			require.NoError(t, err)

			_, err = gcl.GatewayV1().Gateways(namespace).Create(
				ctx,
				&gatewayapi.Gateway{
					ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("test-%d", i)},
					Spec: gatewayapi.GatewaySpec{
						GatewayClassName: gatewayapi.ObjectName("test"),
						Listeners: []gatewayapi.Listener{
							{
								Name:     "test",
								Port:     443,
								Protocol: "HTTP",
							},
						},
					},
				},
				metav1.CreateOptions{},
			)
			require.NoError(t, err)

			_, err = gcl.GatewayV1().HTTPRoutes(namespace).Create(
				ctx,
				&gatewayapi.HTTPRoute{
					ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("test-%d", i)},
					Spec:       gatewayapi.HTTPRouteSpec{},
				},
				metav1.CreateOptions{},
			)
			require.NoError(t, err)

			_, err = gcl.GatewayV1().GRPCRoutes(namespace).Create(
				ctx,
				&gatewayapi.GRPCRoute{
					ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("test-%d", i)},
					Spec:       gatewayapi.GRPCRouteSpec{},
				},
				metav1.CreateOptions{},
			)
			require.NoError(t, err)

			_, err = gcl.GatewayV1alpha2().TCPRoutes(namespace).Create(
				ctx,
				&gatewayapi.TCPRoute{
					ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("test-%d", i)},
					Spec: gatewayapi.TCPRouteSpec{
						Rules: []gatewayapi.TCPRouteRule{{}},
					},
				},
				metav1.CreateOptions{},
			)
			require.NoError(t, err)

			_, err = gcl.GatewayV1alpha2().UDPRoutes(namespace).Create(
				ctx,
				&gatewayapi.UDPRoute{
					ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("test-%d", i)},
					Spec: gatewayapi.UDPRouteSpec{
						Rules: []gatewayapi.UDPRouteRule{{}},
					},
				},
				metav1.CreateOptions{},
			)
			require.NoError(t, err)

			_, err = gcl.GatewayV1alpha2().TLSRoutes(namespace).Create(
				ctx,
				&gatewayapi.TLSRoute{
					ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("test-%d", i)},
					Spec: gatewayapi.TLSRouteSpec{
						Rules: []gatewayapi.TLSRouteRule{{}},
					},
				},
				metav1.CreateOptions{},
			)
			require.NoError(t, err)

			_, err = gcl.GatewayV1beta1().ReferenceGrants(namespace).Create(
				ctx,
				&gatewayapi.ReferenceGrant{
					ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("test-%d", i)},
					Spec: gatewayapi.ReferenceGrantSpec{
						From: []gatewayapi.ReferenceGrantFrom{
							{
								Kind:      "test",
								Namespace: metav1.NamespaceDefault,
							},
						},
						To: []gatewayapi.ReferenceGrantTo{
							{
								Kind: "test",
							},
						},
					},
				},
				metav1.CreateOptions{},
			)
			require.NoError(t, err)
		}
	}
}

func verifyTelemetryReport(t *testing.T, k8sVersion *version.Info, report string) bool {
	t.Helper()
	hostname, err := os.Hostname()
	if err != nil {
		t.Logf("Failed to get hostname: %s", err)
		return false
	}
	semver, err := utilversion.ParseGeneric(k8sVersion.GitVersion)
	if err != nil {
		t.Logf("Failed to parse k8s version: %s", err)
		return false
	}

	// Report contains stanza like:
	// id=57a7a76c-25d0-4394-ab9a-954f7190e39a;
	// uptime=9;
	// that are not stable across runs, so we need to remove them.
	for _, s := range []string{"id", "uptime"} {
		report, err = removeStanzaFromReport(report, s)
		if err != nil {
			// this normally happens during shutdown, when the report is an empty string
			// no point in proceeding if so
			return false
		}
	}

	// this expected report OMITS openshift_version, whereas manager/telemetry.TestCreateManager includes it.
	// the resources created for this test do not include the OpenShift namespace+pod combo, as expected for
	// non-OpenShift clusters. output for non-OpenShift clusters should not include an OpenShift version.
	expectedReport := fmt.Sprintf(
		"<14>"+
			"signal=kic-ping;"+
			"db=off;"+
			"feature-fallbackconfiguration=false;"+
			"feature-fillids=true;"+
			"feature-gateway-service-discovery=false;"+
			"feature-gatewayalpha=false;"+
			"feature-kongcustomentity=true;"+
			"feature-kongservicefacade=false;"+
			"feature-konnect-sync=false;"+
			"feature-rewriteuris=false;"+
			"feature-sanitizekonnectconfigdumps=true;"+
			"hn=%s;"+
			"kv=3.4.1;"+
			"rf=traditional;"+
			"v=NOT_SET;"+
			"k8s_arch=%s;"+
			"k8s_provider=UNKNOWN;"+
			"k8sv=%s;"+
			"k8sv_semver=%s;"+
			"k8s_gatewayclasses_count=2;"+
			"k8s_gateways_count=4;"+
			"k8s_grpcroutes_count=4;"+
			"k8s_httproutes_count=4;"+
			"k8s_ingresses_count=4;"+
			"k8s_nodes_count=2;"+
			"k8s_pods_count=4;"+
			"k8s_referencegrants_count=4;"+
			"k8s_services_count=5;"+
			"k8s_tcproutes_count=4;"+
			"k8s_tlsroutes_count=4;"+
			"k8s_udproutes_count=4;"+
			"\n",
		hostname,
		k8sVersion.Platform,
		k8sVersion.GitVersion,
		"v"+semver.String(),
	)
	if diff := cmp.Diff(expectedReport, report); diff != "" {
		t.Logf("telemetry report mismatch (-want +got):\n%s", diff)
		return false
	}
	return true
}

// removeStanzaFromReport removes stanza from report. Report contains stanzas like:
// id=57a7a76c-25d0-4394-ab9a-954f7190e39a;uptime=9;k8s_services_count=5; etc.
// Pass e.g. "uptime" to remove the whole uptime=9; from the report.
func removeStanzaFromReport(report string, stanza string) (string, error) {
	const idStanzaEnd = ";"
	stanza += "="
	start := strings.Index(report, stanza)
	if start == -1 {
		return "", fmt.Errorf("stanza %q not found in report: %s", stanza, report)
	}
	end := strings.Index(report[start:], idStanzaEnd)
	if end == -1 {
		return "", errors.New("stanza end not found in report")
	}
	end += start
	return report[:start] + report[end+1:], nil
}
