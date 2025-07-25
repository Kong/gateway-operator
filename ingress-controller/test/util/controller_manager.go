package util

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/kong/kubernetes-testing-framework/pkg/clusters"
	ktfkong "github.com/kong/kubernetes-testing-framework/pkg/clusters/addons/kong"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kong/kong-operator/ingress-controller/internal/cmd/rootcmd/config"
	"github.com/kong/kong-operator/ingress-controller/internal/logging"
	"github.com/kong/kong-operator/ingress-controller/pkg/manager"
	managercfg "github.com/kong/kong-operator/ingress-controller/pkg/manager/config"
	"github.com/kong/kong-operator/ingress-controller/test/consts"
)

// logOutput is a file to use for manager log output other than stderr.
var logOutput = os.Getenv("TEST_KONG_KIC_MANAGER_LOG_OUTPUT")

// -----------------------------------------------------------------------------
// Testing Utility Functions - Controller Manager
// -----------------------------------------------------------------------------

// PrepareClusterForRunningControllerManager prepares the provided cluster for running
// the controller manager.
// It creates kong's namespace, deploys its RBAC manifests and CRDs.
func PrepareClusterForRunningControllerManager(
	ctx context.Context,
	cluster clusters.Cluster,
) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: consts.ControllerNamespace,
		},
	}
	nsClient := cluster.Client().CoreV1().Namespaces()
	if _, err := nsClient.Create(ctx, ns, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed creating %s namespace: %w", ns.Name, err)
		}
	}

	// Deploy all RBACs required for testing to the cluster.
	if err := DeployRBACsForCluster(ctx, cluster); err != nil {
		return fmt.Errorf("failed to deploy RBACs: %w", err)
	}

	// Deploy all CRDs required for testing to the cluster.
	if err := DeployCRDsForCluster(ctx, cluster); err != nil {
		return fmt.Errorf("failed to deploy CRDs: %w", err)
	}

	return nil
}

// DeployControllerManagerForCluster deploys all the base CRDs needed for the
// controller manager to function, and also runs a copy of the controller
// manager on a provided test cluster.
//
// Controller managers started this way will run in the background in a goroutine:
// The caller must use the provided context.Context to stop the controller manager
// from running when they're done with it.
//
// This accepts both a list of command line flags with arguments as well as config options
// for the manager. The reason for this is that there exist some flags that are
// not exposed via the CLI interface yet like e.g. GracefulShutdownTimeout.
//
// It returns a context cancellation func which will stop the manager and an error.
func DeployControllerManagerForCluster(
	ctx context.Context,
	logger logr.Logger,
	cluster clusters.Cluster,
	kongAddon *ktfkong.Addon,
	additionalFlags []string,
	opts ...managercfg.Opt,
) (func(), error) {
	if kongAddon == nil {
		// Ensure that the provided test cluster has a kongAddon deployed to it.
		for _, addon := range cluster.ListAddons() {
			a, ok := addon.(*ktfkong.Addon)
			if ok {
				kongAddon = a
			}
		}
	}
	if kongAddon == nil {
		return nil, fmt.Errorf("no Kong addon found loaded in cluster %s", cluster.Name())
	}

	// determine the proxy admin URL for the Kong Gateway on this cluster:
	proxyAdminURL, err := kongAddon.ProxyAdminURL(ctx, cluster)
	if err != nil {
		return nil, fmt.Errorf("couldn't determine Kong Gateway Admin URL on cluster %s: %w", cluster.Name(), err)
	}

	// create a tempfile to hold the cluster kubeconfig that will be used for the controller
	// generate a temporary kubeconfig since we're going to be using the helm CLI
	kubeconfig, err := clusters.TempKubeconfig(cluster)
	if err != nil {
		return nil, err
	}

	// render all controller manager flag options
	controllerManagerFlags := []string{
		fmt.Sprintf("--kong-admin-url=%s", proxyAdminURL),
		fmt.Sprintf("--kubeconfig=%s", kubeconfig.Name()),
		"--election-id=integrationtests.konghq.com",
		"--log-format=text",
		fmt.Sprintf("--publish-service=%s/ingress-controller-kong-proxy", kongAddon.Namespace()),
		fmt.Sprintf("--publish-service-udp=%s/ingress-controller-kong-udp-proxy", kongAddon.Namespace()),
	}
	controllerManagerFlags = append(controllerManagerFlags, additionalFlags...)

	clicfg := config.NewCLIConfig()
	clicfg.Impersonate = "system:serviceaccount:kong:kong-serviceaccount"

	// parsing the controller configuration flags
	// TODO(czeslavo): we may switch to using managercfg.Config directly instead of relying on CLI flags here.
	flags := clicfg.FlagSet()
	if err := flags.Parse(controllerManagerFlags); err != nil {
		os.Remove(kubeconfig.Name())
		return nil, fmt.Errorf("failed to parse manager flags: %w", err)
	}

	for _, opt := range opts {
		opt(clicfg.Config)
	}

	ctx, cancel := context.WithCancel(ctx)
	// run the controller in the background
	go func() {
		defer os.Remove(kubeconfig.Name())
		fmt.Fprintf(os.Stderr, "INFO: Starting Controller Manager for Cluster %s with Configuration: %+v\n", cluster.Name(), clicfg)
		mid, err := manager.NewID("kic-" + cluster.Name())
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Problems with creating Manager ID: %s\n", err)
			os.Exit(1)
		}
		m, err := manager.NewManager(ctx, mid, logger, *clicfg.Config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Problems with set up of Controller Manager: %s\n", err)
			os.Exit(1)
		}
		if err := m.Run(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Problems with running Controller Manager: %s\n", err)
			os.Exit(1)
		}
	}()

	return cancel, nil
}

// SetupLoggers sets up the loggers for the controller manager.
// The controller manager logs needs to be setup before the 30s timeout passes.
// Given the cluster deployment takes time, there is a chance that the timeout
// will pass before the controller manager logs are setup.
// This function can be used to sets up the loggers for the controller manager
// before the cluster deployment.
func SetupLoggers(logLevel string, logFormat string) (logr.Logger, string, error) {
	// construct the config for the logger
	var err error
	output := os.Stderr
	if logOutput != "" {
		output, err = os.OpenFile(logOutput, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
		if err != nil {
			return logr.Logger{}, logOutput, err
		}
	}
	config := managercfg.Config{
		LogLevel:  logLevel,
		LogFormat: logFormat,
	}

	logger, err := logging.SetupLoggers(config, output)
	// Prevents controller-runtime from logging
	// [controller-runtime] log.SetLogger(...) was never called; logs will not be displayed.
	return logger, "", err
}
