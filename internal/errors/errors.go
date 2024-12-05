package errors

import (
	"errors"
	"fmt"
)

// -----------------------------------------------------------------------------
// Objects conversion - Errors
// -----------------------------------------------------------------------------

// ErrUnexpectedObject is a custom error that must be used when the cast of a object to an expected
// type fails.
var ErrUnexpectedObject = errors.New("unexpected object type provided")

// -----------------------------------------------------------------------------
// Gateway - Errors
// -----------------------------------------------------------------------------

// ErrUnsupportedGatewayClass is an error which indicates that a provided GatewayClass
// is not supported.
type ErrUnsupportedGatewayClass struct {
	reason string
}

func NewErrUnsupportedGateway(reason string) ErrUnsupportedGatewayClass {
	return ErrUnsupportedGatewayClass{reason: reason}
}

func (e ErrUnsupportedGatewayClass) Error() string {
	return fmt.Sprintf("unsupported gateway class: %s", e.reason)
}

// -----------------------------------------------------------------------------
// GatewayClass - Errors
// -----------------------------------------------------------------------------

// ErrObjectMissingParametersRef is a custom error that must be used when the
// .spec.ParametersRef field of the given object is nil
var ErrObjectMissingParametersRef = errors.New("no reference to related objects")

// -----------------------------------------------------------------------------
// ControlPlane - Errors
// -----------------------------------------------------------------------------

// ErrDataPlaneNotSet is a custom error that must be used when a specific OwnerReference
// is expected to be on an object, but it is not found.
var ErrDataPlaneNotSet = errors.New("no dataplane name set")

// ErrNoDataPlanePods is a custom error that must be used when the DataPlane Deployment
// referenced by the ControlPlane has no pods ready yet.
var ErrNoDataPlanePods = errors.New("no dataplane pods existing yet")

// -----------------------------------------------------------------------------
// Version Strings - Errors
// -----------------------------------------------------------------------------

// ErrInvalidSemverVersion is a custom error that indicates a provided
// version string (which we were expecting to be in the format of
// <Major>.<Minor>.<Patch>) was invalid, and not in the expected format.
var ErrInvalidSemverVersion = errors.New("not a valid semver version")
