package sendconfig

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"

	"github.com/kong/kong-operator/ingress-controller/internal/dataplane/deckgen"
	"github.com/kong/kong-operator/ingress-controller/internal/diagnostics"
	"github.com/kong/kong-operator/ingress-controller/internal/logging"
	"github.com/kong/kong-operator/ingress-controller/internal/metrics"
	"github.com/kong/kong-operator/ingress-controller/internal/store"
	"github.com/kong/kong-operator/ingress-controller/internal/util"
)

// -----------------------------------------------------------------------------
// Sendconfig - Public Functions
// -----------------------------------------------------------------------------

type UpdateStrategyResolver interface {
	ResolveUpdateStrategy(client UpdateClient, diagnostic *diagnostics.Client) UpdateStrategy
}

type AdminAPIClient interface {
	AdminAPIClient() *kong.Client
	LastConfigSHA() []byte
	SetLastConfigSHA([]byte)
	SetLastCacheStoresHash(store.SnapshotHash)
	BaseRootURL() string
	PluginSchemaStore() *util.PluginSchemaStore

	IsKonnect() bool
	KonnectControlPlane() string
}

// PerformUpdate writes `targetContent` to Kong Admin API specified by `kongConfig`.
func PerformUpdate(
	ctx context.Context,
	logger logr.Logger,
	client AdminAPIClient,
	config Config,
	targetContent *file.Content,
	customEntities CustomEntitiesByType,
	promMetrics metrics.Recorder,
	updateStrategyResolver UpdateStrategyResolver,
	configChangeDetector ConfigurationChangeDetector,
	diagnostic *diagnostics.Client,
	isFallback bool,
) ([]byte, error) {
	oldSHA := client.LastConfigSHA()
	newSHA, err := deckgen.GenerateSHA(targetContent, customEntities)
	if err != nil {
		return oldSHA, fmt.Errorf("failed to generate SHA for target content: %w", err)
	}

	// Disable checking whether the actual SHA is as generated when:
	// - reverse sync is enabled
	// - or running in DB mode and in fallback mode
	//
	// We do this because in db mode, when applying fails, some entities are applied successfully, then the Kong gateway may be in a "partial success" state.
	if !config.EnableReverseSync && (config.InMemory || !isFallback) {
		configurationChanged, err := configChangeDetector.HasConfigurationChanged(ctx, oldSHA, newSHA, targetContent, client.AdminAPIClient())
		if err != nil {
			return nil, fmt.Errorf("failed to detect configuration change: %w", err)
		}
		if !configurationChanged {
			if client.IsKonnect() {
				logger.V(logging.DebugLevel).Info("No configuration change, skipping sync to Konnect")
			} else {
				logger.V(logging.DebugLevel).Info("No configuration change, skipping sync to Kong")
			}
			return oldSHA, nil
		}
	}

	updateStrategy := updateStrategyResolver.ResolveUpdateStrategy(client, diagnostic)
	logger = logger.WithValues("update_strategy", updateStrategy.Type())
	timeStart := time.Now()
	size, err := updateStrategy.Update(ctx, ContentWithHash{
		Content:        targetContent,
		CustomEntities: customEntities,
		Hash:           newSHA,
	})
	duration := time.Since(timeStart)

	metricsProtocol := updateStrategy.MetricsProtocol()
	if err != nil {
		// For UpdateError, record the failure and return the error.
		var updateError UpdateError
		if errors.As(err, &updateError) {
			if isFallback {
				promMetrics.RecordFallbackPushFailure(metricsProtocol, duration, updateError.ConfigSize(), client.BaseRootURL(), len(updateError.ResourceFailures()), updateError.err)
			} else {
				promMetrics.RecordPushFailure(metricsProtocol, duration, updateError.ConfigSize(), client.BaseRootURL(), len(updateError.ResourceFailures()), updateError.err)
			}
			return nil, updateError
		}

		// Any other error, simply return it and skip metrics recording - we have no details to record.
		return nil, fmt.Errorf("config update failed: %w", err)
	}

	if isFallback {
		promMetrics.RecordFallbackPushSuccess(metricsProtocol, duration, size, client.BaseRootURL())
	} else {
		promMetrics.RecordPushSuccess(metricsProtocol, duration, size, client.BaseRootURL())
	}

	if client.IsKonnect() {
		logger.V(logging.InfoLevel).Info("Successfully synced configuration to Konnect", "duration", duration.Truncate(time.Millisecond).String())
	} else {
		logger.V(logging.InfoLevel).Info("Successfully synced configuration to Kong", "duration", duration.Truncate(time.Millisecond).String())
	}

	return newSHA, nil
}
