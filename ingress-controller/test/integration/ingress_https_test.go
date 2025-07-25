//go:build integration_tests

package integration

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kong/go-kong/kong"
	"github.com/kong/kubernetes-testing-framework/pkg/clusters"
	"github.com/kong/kubernetes-testing-framework/pkg/utils/kubernetes/generators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kong/kong-operator/ingress-controller/test"
	"github.com/kong/kong-operator/ingress-controller/test/consts"
	"github.com/kong/kong-operator/ingress-controller/test/helpers/certificate"
	"github.com/kong/kong-operator/ingress-controller/test/internal/helpers"
)

func TestHTTPSRedirect(t *testing.T) {
	RunWhenKongExpressionRouter(t.Context(), t)
	ctx := t.Context()

	t.Parallel()
	ns, cleaner := helpers.Setup(ctx, t, env)

	t.Log("creating an HTTP container via deployment to test redirect functionality")
	container := generators.NewContainer("alsohttpbin", test.HTTPBinImage, test.HTTPBinPort)
	deployment := generators.NewDeploymentForContainer(container)
	opts := metav1.CreateOptions{}
	_, err := env.Cluster().Client().AppsV1().Deployments(ns.Name).Create(ctx, deployment, opts)
	assert.NoError(t, err)
	cleaner.Add(deployment)

	t.Logf("exposing deployment %s via Service", deployment.Name)
	service := generators.NewServiceForDeployment(deployment, corev1.ServiceTypeClusterIP)
	service, err = env.Cluster().Client().CoreV1().Services(ns.Name).Create(ctx, service, opts)
	assert.NoError(t, err)
	cleaner.Add(service)

	t.Logf("exposing Service %s via Ingress", service.Name)
	ingress := generators.NewIngressForService("/test_https_redirect", map[string]string{
		"konghq.com/protocols":                  "https",
		"konghq.com/https-redirect-status-code": "301",
	}, service)
	ingress.Spec.IngressClassName = kong.String(consts.IngressClass)
	assert.NoError(t, clusters.DeployIngress(ctx, env.Cluster(), ns.Name, ingress))
	cleaner.Add(ingress)

	t.Log("waiting for Ingress to be operational and properly redirect")
	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Second * 3,
	}
	assert.Eventually(t, func() bool {
		resp, err := client.Get(fmt.Sprintf("%s/test_https_redirect", proxyHTTPURL))
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusMovedPermanently
	}, ingressWait, waitTick)
}

func TestHTTPSIngress(t *testing.T) {
	ctx := t.Context()

	t.Parallel()
	ns, cleaner := helpers.Setup(ctx, t, env)

	certPool := x509.NewCertPool()
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	testTransport := http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			switch addr {
			case "foo.example:443", "bar.example:443", "baz.example:443":
				addr = proxyHTTPSURL.Host
			default:
			}
			return dialer.DialContext(ctx, network, addr)
		},
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			RootCAs:    certPool,
		},
	}
	httpcStatic := http.Client{
		Timeout:   httpcTimeout,
		Transport: &testTransport,
	}

	t.Log("deploying a minimal HTTP container deployment to test Ingress routes")
	container := generators.NewContainer("httpbin", test.HTTPBinImage, test.HTTPBinPort)
	deployment := generators.NewDeploymentForContainer(container)
	deployment, err := env.Cluster().Client().AppsV1().Deployments(ns.Name).Create(ctx, deployment, metav1.CreateOptions{})
	assert.NoError(t, err)
	cleaner.Add(deployment)

	t.Logf("exposing deployment %s via service", deployment.Name)
	service := generators.NewServiceForDeployment(deployment, corev1.ServiceTypeLoadBalancer)
	_, err = env.Cluster().Client().CoreV1().Services(ns.Name).Create(ctx, service, metav1.CreateOptions{})
	assert.NoError(t, err)
	cleaner.Add(service)

	t.Logf("creating an ingress for service %s with ingress.class %s", service.Name, consts.IngressClass)
	ingress1 := generators.NewIngressForService("/foo", map[string]string{
		"konghq.com/strip-path": "true",
	}, service)
	ingress1.Spec.IngressClassName = kong.String(consts.IngressClass)
	ingress2 := generators.NewIngressForService("/bar", map[string]string{
		"konghq.com/strip-path": "true",
	}, service)
	ingress2.Spec.IngressClassName = kong.String(consts.IngressClass)

	t.Log("configuring ingress tls spec")
	ingress1.Spec.TLS = []netv1.IngressTLS{{SecretName: "secret1", Hosts: []string{"foo.example"}}}
	ingress1.Name = "ingress1"
	ingress2.Spec.TLS = []netv1.IngressTLS{{SecretName: "secret2", Hosts: []string{"bar.example", "baz.example"}}}
	ingress2.Name = "ingress2"

	t.Log("configuring secrets")
	fooExampleTLSCert, fooExampleTLSKey := certificate.MustGenerateCertPEMFormat(
		certificate.WithCommonName("secure-foo-bar"), certificate.WithDNSNames("secure-foo-bar", "foo.example"),
	)
	require.True(t, certPool.AppendCertsFromPEM(fooExampleTLSCert))
	barExampleTLSCert, barExampleTLSKey := certificate.MustGenerateCertPEMFormat(
		certificate.WithCommonName("foo.com"), certificate.WithDNSNames("foo.com", "bar.example", "baz.example"),
	)
	require.True(t, certPool.AppendCertsFromPEM(barExampleTLSCert))

	secrets := []*corev1.Secret{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret1",
				Namespace: ns.Name,
			},
			Data: map[string][]byte{
				"tls.crt": fooExampleTLSCert,
				"tls.key": fooExampleTLSKey,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret2",
				Namespace: ns.Name,
			},
			Data: map[string][]byte{
				"tls.crt": barExampleTLSCert,
				"tls.key": barExampleTLSKey,
			},
		},
	}

	// Since we updated the logic of secret controller to only process secrets that are referred by
	// other controlled objects (service, ingress, gateway, ...), we should make sure that ingresses
	// created before and after referred secret created both works.
	// so here we interleave the creating process of deploying 2 ingresses and secrets.
	t.Log("deploying secrets and ingresses")
	require.NoError(t, clusters.DeployIngress(ctx, env.Cluster(), ns.Name, ingress1))
	cleaner.Add(ingress1)

	secret1, err := env.Cluster().Client().CoreV1().Secrets(ns.Name).Create(ctx, secrets[0], metav1.CreateOptions{})
	assert.NoError(t, err)
	cleaner.Add(secret1)

	secret2, err := env.Cluster().Client().CoreV1().Secrets(ns.Name).Create(ctx, secrets[1], metav1.CreateOptions{})
	assert.NoError(t, err)
	cleaner.Add(secret2)

	require.NoError(t, clusters.DeployIngress(ctx, env.Cluster(), ns.Name, ingress2))
	cleaner.Add(ingress2)

	t.Log("checking first ingress status readiness")
	require.Eventually(t, func() bool {
		lbstatus, err := clusters.GetIngressLoadbalancerStatus(ctx, env.Cluster(), ns.Name, ingress1)
		if err != nil {
			return false
		}
		for _, ingress := range lbstatus.Ingress {
			if len(ingress.Hostname) > 0 || len(ingress.IP) > 0 {
				t.Logf("networkingv1 ingress1 hostname %s or ip %s is ready to redirect traffic.", ingress.Hostname, ingress.IP)
				return true
			}
		}
		return false
	}, statusWait, waitTick, true)

	t.Log("checking second ingress status readiness")
	assert.Eventually(t, func() bool {
		lbstatus, err := clusters.GetIngressLoadbalancerStatus(ctx, env.Cluster(), ns.Name, ingress2)
		if err != nil {
			return false
		}
		for _, ingress := range lbstatus.Ingress {
			if len(ingress.Hostname) > 0 || len(ingress.IP) > 0 {
				t.Logf("networkingv1 ingress2 hostname %s or ip %s is ready to redirect traffic.", ingress.Hostname, ingress.IP)
				return true
			}
		}
		return false
	}, statusWait, waitTick, true)

	t.Log("waiting for routes from Ingress to be operational with expected certificate")
	assert.Eventually(t, func() bool {
		resp, err := httpcStatic.Get("https://foo.example:443/foo")
		if err != nil {
			t.Logf("WARNING: error while waiting for https://foo.example:443/foo: %v", err)
			return false
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			b := new(bytes.Buffer)
			n, err := b.ReadFrom(resp.Body)
			require.NoError(t, err)
			require.True(t, n > 0)
			return strings.Contains(b.String(), "<title>httpbin.org</title>") && resp.TLS.PeerCertificates[0].Subject.CommonName == "secure-foo-bar"
		}
		return false
	}, ingressWait, waitTick, true)

	t.Log("waiting for routes from Ingress to be operational with expected certificate")
	assert.Eventually(t, func() bool {
		resp, err := httpcStatic.Get("https://bar.example:443/bar")
		if err != nil {
			t.Logf("WARNING: error while waiting for https://bar.example:443/bar: %v", err)
			return false
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			b := new(bytes.Buffer)
			n, err := b.ReadFrom(resp.Body)
			require.NoError(t, err)
			require.True(t, n > 0)
			return strings.Contains(b.String(), "<title>httpbin.org</title>") && resp.TLS.PeerCertificates[0].Subject.CommonName == "foo.com"
		}
		return false
	}, ingressWait, waitTick, true)

	t.Log("confirm Ingress path routes available on other hostnames")
	assert.Eventually(t, func() bool {
		resp, err := httpcStatic.Get("https://baz.example:443/bar")
		if err != nil {
			t.Logf("WARNING: error while waiting for https://baz.example:443/bar: %v", err)
			return false
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			b := new(bytes.Buffer)
			n, err := b.ReadFrom(resp.Body)
			require.NoError(t, err)
			require.True(t, n > 0)
			return strings.Contains(b.String(), "<title>httpbin.org</title>")
		}
		return false
	}, ingressWait, waitTick)

	ingress2, err = env.Cluster().Client().NetworkingV1().Ingresses(ns.Name).Get(ctx, ingress2.Name, metav1.GetOptions{})
	assert.NoError(t, err)
	ingress2.Annotations["konghq.com/snis"] = "bar.example"
	_, err = env.Cluster().Client().NetworkingV1().Ingresses(ns.Name).Update(ctx, ingress2, metav1.UpdateOptions{})
	assert.NoError(t, err)

	t.Log("confirm Ingress no longer routes without matching SNI")
	assert.Eventually(t, func() bool {
		resp, err := httpcStatic.Get("https://baz.example:443/bar")
		if err != nil {
			t.Logf("WARNING: error while waiting for https://baz.example:443/bar: %v", err)
			return false
		}

		defer resp.Body.Close()
		return resp.StatusCode == http.StatusNotFound
	}, ingressWait, waitTick)

	t.Log("confirm Ingress still routes with matching SNI")
	assert.Eventually(t, func() bool {
		resp, err := httpcStatic.Get("https://bar.example:443/bar")
		if err != nil {
			t.Logf("WARNING: error while waiting for https://bar.example:443/bar: %v", err)
			return false
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			b := new(bytes.Buffer)
			n, err := b.ReadFrom(resp.Body)
			require.NoError(t, err)
			require.True(t, n > 0)
			return strings.Contains(b.String(), "<title>httpbin.org</title>")
		}
		return false
	}, ingressWait, waitTick)
}
