package errors

import (
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DataPlaneNotSetError is a custom object that must be raised when a specific OwnerReference
// is expected to be on an object, but it is not found.
type DataPlaneNotSetError struct {
	object metav1.Object

	message string
}

func (err *DataPlaneNotSetError) Error() string {
	return err.message
}

func NewDataPlaneNotSetError(obj metav1.Object) error {
	return &DataPlaneNotSetError{
		object:  obj,
		message: fmt.Sprintf("no dataplane name set on controlplan %s/%s spec", obj.GetNamespace(), obj.GetName()),
	}
}

func IsDataPlaneNotSet(err error) bool {
	var onwRefErr *DataPlaneNotSetError
	return errors.As(err, &onwRefErr)
}

// ObjectMissingParametersRefError is a custom object that must be raised when the
// .spec.ParametersRef field of the given object is nil
type ObjectMissingParametersRefError struct {
	object metav1.Object

	message string
}

func (err *ObjectMissingParametersRefError) Error() string {
	return err.message
}

func NewObjectMissingParametersRef(obj metav1.Object) error {
	return &ObjectMissingParametersRefError{
		object:  obj,
		message: fmt.Sprintf("object %s/%s has not reference to related objects", obj.GetNamespace(), obj.GetName()),
	}
}

func IsObjectMissingParametersRef(err error) bool {
	var onwRefErr *ObjectMissingParametersRefError
	return errors.As(err, &onwRefErr)
}
