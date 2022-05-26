//go:build integration_tests
// +build integration_tests

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/kong/kubernetes-testing-framework/pkg/clusters"
	"github.com/kong/kubernetes-testing-framework/pkg/clusters/addons/metallb"
	"github.com/kong/kubernetes-testing-framework/pkg/clusters/types/kind"
	"github.com/kong/kubernetes-testing-framework/pkg/environments"
	"k8s.io/client-go/kubernetes"

	"github.com/kong/gateway-operator/internal/manager"
	"github.com/kong/gateway-operator/pkg/clientset"
)

// -----------------------------------------------------------------------------
// Testing Vars - Environment Overrideable
// -----------------------------------------------------------------------------

var (
	existingClusterName  = os.Getenv("KONG_TEST_CLUSTER")
	controllerManagerOut = os.Getenv("KONG_CONTROLLER_OUT")
)

// -----------------------------------------------------------------------------
// Testing Vars - Testing Environment
// -----------------------------------------------------------------------------

var (
	ctx    context.Context
	cancel context.CancelFunc
	env    environments.Environment

	k8sClient      *kubernetes.Clientset
	operatorClient *clientset.Clientset
)

// -----------------------------------------------------------------------------
// Testing Main
// -----------------------------------------------------------------------------

func TestMain(m *testing.M) {
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	var skipClusterCleanup bool
	var existingCluster clusters.Cluster
	var err error

	fmt.Println("INFO: setting up test cluster")
	if existingClusterName != "" {
		existingCluster, err = kind.NewFromExisting(existingClusterName)
		exitOnErr(err)
		skipClusterCleanup = true
		fmt.Printf("INFO: using existing kind cluster (name: %s)\n", existingCluster.Name())
	}

	fmt.Println("INFO: setting up test environment")
	envBuilder := environments.NewBuilder()
	if existingCluster != nil {
		envBuilder.WithExistingCluster(existingCluster)
	}
	env, err = envBuilder.WithAddons(metallb.New()).Build(ctx)
	exitOnErr(err)

	fmt.Printf("INFO: waiting for cluster %s and all addons to become ready\n", env.Cluster().Name())
	exitOnErr(<-env.WaitForReady(ctx))

	fmt.Println("INFO: initializing Kubernetes API clients")
	k8sClient = env.Cluster().Client()
	operatorClient, err = clientset.NewForConfig(env.Cluster().Config())
	exitOnErr(err)

	fmt.Println("INFO: deploying CRDs to test cluster")
	exitOnErr(clusters.KustomizeDeployForCluster(ctx, env.Cluster(), "../../config/crd"))

	fmt.Println("INFO: starting the operator's controller manager")
	go startControllerManager()

	fmt.Println("INFO: environment is ready, starting tests")
	code := m.Run()

	if !skipClusterCleanup {
		fmt.Println("INFO: cleaning up testing cluster and environment")
		exitOnErr(env.Cleanup(ctx))
	}

	os.Exit(code)
}

// -----------------------------------------------------------------------------
// Testing Main - Helper Functions
// -----------------------------------------------------------------------------

func exitOnErr(err error) {
	if err != nil {
		if env != nil {
			env.Cleanup(ctx) //nolint:errcheck
		}
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}
}

func startControllerManager() {
	cfg := manager.DefaultConfig
	cfg.LeaderElection = false
	cfg.DevelopmentMode = true

	if controllerManagerOut != "stdout" {
		out, err := os.CreateTemp(os.TempDir(), "gateway-operator-controller-logs")
		exitOnErr(err)
		cfg.Out = out
		fmt.Printf("INFO: controller output is being logged to %s\n", out.Name())
		defer out.Close()
	}

	exitOnErr(manager.Run(cfg))
}
