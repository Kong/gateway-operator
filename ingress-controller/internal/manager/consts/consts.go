package consts

import "time"

// -----------------------------------------------------------------------------
// Controller Manager - Constants & Vars
// -----------------------------------------------------------------------------

const (
	// HealthzPort is the default port the manager's health service listens on.
	// Changing this will result in a breaking change. Existing deployments may use the literal
	// port number in their liveness and readiness probes, and upgrading to a controller version
	// with a changed HealthzPort will result in crash loops until users update their probe config.
	// Note that there are several stock manifests in this repo that also use the literal port number. If you
	// update this value, search for the old port number and update the stock manifests also.
	HealthzPort = 10254

	// MetricsPort is the default port the manager's metrics service listens on.
	// Similar to HealthzPort, it may be used in existing user deployment configurations, and its
	// literal value is used in several stock manifests, which must be updated along with this value.
	MetricsPort = 10255

	// DiagnosticsPort is the default port of the manager's diagnostics service listens on.
	DiagnosticsPort = 10256

	// KongClientEventRecorderComponentName is a KongClient component name used to identify the events recording component.
	KongClientEventRecorderComponentName = "kong-client"

	// DefaultEnableDrainSupport is the default value for enabling drain support feature.
	// When enabled, terminating endpoints are kept in Kong upstreams with weight=0 for graceful connection draining.
	DefaultEnableDrainSupport = false

	// DefaultConfigMapSelector is the default label selector used to ingest ConfigMaps in the DataPlane sync.
	DefaultConfigMapSelector = "konghq.com/configmap"

	// InstanceIDAnnotationKey is the annotation key used to store the instance ID of particular manager,
	// to corelate events with it.
	InstanceIDAnnotationKey = "konghq.com/instance-id"

	// DefaultGracefulShutdownTimeout is the default timeout for graceful shutdown. It is used by the manager
	// subcomponents to wait for the resources to be cleaned up before shutting down.
	DefaultGracefulShutdownTimeout = time.Second * 5
)
