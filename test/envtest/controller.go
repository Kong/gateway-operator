package envtest

import (
	"context"
	"sync"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

// Reconciler represents a reconciler.
type Reconciler interface {
	SetupWithManager(context.Context, ctrl.Manager) error
}

// NewManager returns a manager and a logs observer.
// The logs observer can be used to dump logs if the test fails.
// The returned manager can be used with StartReconcilers() to start a list of
// provided reconcilers with the manager.
func NewManager(t *testing.T, ctx context.Context, cfg *rest.Config, s *runtime.Scheme) (manager.Manager, LogsObserver) {
	_, logger, logs := CreateTestLogger(ctx)

	o := manager.Options{
		Logger: logger,
		Scheme: s,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
		Controller: config.Controller{
			// We don't want to hide panics in tests so expose them by failing the test run rather
			// than silently ignoring them.
			RecoverPanic: lo.ToPtr(false),

			// This is needed because controller-runtime keeps a global list of controller
			// names and panics if there are duplicates.
			// This is a workaround for that in tests.
			// Ref: https://github.com/kubernetes-sigs/controller-runtime/pull/2902#issuecomment-2284194683
			SkipNameValidation: lo.ToPtr(true),
		},
	}

	mgr, err := ctrl.NewManager(cfg, o)
	require.NoError(t, err)
	return mgr, logs
}

// StartReconcilers creates a controller manager and starts the provided reconciler
// as its runnable.
// It also adds a t.Cleanup which waits for the manager to exit so that the test
// can be self contained and logs from different tests' managers don't mix up.
func StartReconcilers(
	ctx context.Context,
	t *testing.T,
	mgr manager.Manager,
	logs LogsObserver,
	reconcilers ...Reconciler,
) {
	t.Helper()

	for _, r := range reconcilers {
		require.NoError(t, r.SetupWithManager(ctx, mgr))
	}

	// This wait group makes it so that we wait for manager to exit.
	// This way we get clean test logs not mixing between tests.
	wg := sync.WaitGroup{}
	wg.Add(1)
	t.Logf("Starting manager for test case %s", t.Name())
	go func() {
		defer wg.Done()
		assert.NoError(t, mgr.Start(ctx))
	}()
	t.Cleanup(func() {
		wg.Wait()
		DumpLogsIfTestFailed(t, logs)
	})
}

// NewControllerClient returns a new controller-runtime Client for provided runtime.Scheme and rest.Config.
func NewControllerClient(t *testing.T, scheme *runtime.Scheme, cfg *rest.Config) ctrlclient.Client {
	client, err := ctrlclient.New(cfg, ctrlclient.Options{
		Scheme: scheme,
	})
	require.NoError(t, err)
	return client
}
