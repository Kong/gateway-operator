package konnect

import "fmt"

type FailedKonnectOpError[T SupportedKonnectEntityType] struct {
	Op  Op
	Err error
}

func (e FailedKonnectOpError[T]) Error() string {
	return fmt.Sprintf("failed to %s %s on Konnect: %v",
		e.Op, entityTypeName[T](), e.Err,
	)
}

func (e FailedKonnectOpError[T]) Unwrap() error {
	return e.Err
}
