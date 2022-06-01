package consts

// -----------------------------------------------------------------------------
// Consts - Standard Kubernetes Object Labels
// -----------------------------------------------------------------------------

const (
	// GatewayOperatorControlledLabel is the label that is used for objects which
	// were created by this operator.
	GatewayOperatorControlledLabel = "konghq.com/gateway-operator"

	// DataPlaneManagedLabelValue indicates that an object's lifecycle is managed
	// by the dataplane controller.
	DataPlaneManagedLabelValue = "dataplane"
)

// -----------------------------------------------------------------------------
// Consts - Container Images
// -----------------------------------------------------------------------------

const (
	// DefaultDataPlaneBaseImage is the base container image that can be used
	// by default for a DataPlane resource if all other attempts to dynamically
	// decide an image fail.
	DefaultDataPlaneBaseImage = "kong"

	// DefaultDataPlaneTag is the base container image tag that can be used
	// by default for a DataPlane resource if all other attempts to dynamically
	// decide an image tag fail.
	DefaultDataPlaneTag = "2.8" // TODO: create CI job that will patch update this automatically

	// DefaultDataPlaneImage is the default container image that can be used if
	// all other attempts to dynamically decide the default image fail.
	DefaultDataPlaneImage = DefaultDataPlaneBaseImage + ":" + DefaultDataPlaneTag
)
