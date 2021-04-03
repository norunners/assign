package assign

import (
	"fmt"
	"reflect"
)

// ErrorType handles the invalid assign of types case.
type ErrorType struct {
	// Dst is the reflection type of the Go value.
	// The type is nil when the Go value passed is a nil pointer.
	Dst reflect.Type
	// Src is the reflection kind of the source.
	Src reflect.Kind
}

// newError creates a new ErrorType.
func newError(dst reflect.Type, src reflect.Kind) ErrorType {
	return ErrorType{
		Dst: dst,
		Src: src,
	}
}

func (e ErrorType) Error() string {
	return fmt.Sprintf("failed to assign to type: %v from source kind: %v", e.Dst, e.Src)
}

// ErrorCycle handles the cyclical paths case.
type ErrorCycle struct {
	Dst reflect.Type
	Src reflect.Kind
}

func (e ErrorCycle) Error() string {
	return fmt.Sprintf("cyclical assign found at type: %q and source kind: %q", e.Dst, e.Src)
}

// ErrorPanic handles the panic case.
// Please report all panics as they are unexpected.
type ErrorPanic struct {
	Rec interface{}
}

func (e ErrorPanic) Error() string {
	return fmt.Sprintf("unexpected panic: %+v", e.Rec)
}
