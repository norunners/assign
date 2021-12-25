package assign

import (
	"reflect"
)

// Source represents any value which can be assigned to a Go value.
type Source interface {
	// Kind is the reflection based kind of the type.
	Kind() reflect.Kind
	// Elem is the contained value of the pointer or interface.
	Elem() Source
	// FieldByName retrieves the field value of the struct by name.
	FieldByName(string) Source
	// Len is the length of the map, slice or array.
	Len() int
	// Index retrieves the element of the slice or array at the index.
	Index(int) Source
	// Pointer retrieves the underlying pointer of the value.
	// This is used to prevent circular paths.
	// Disable this from being used with the WithoutCycle option.
	Pointer() uintptr
	// MapRange provides a map iterator.
	MapRange() MapIter
	// Skip is how values are skipped for assignment.
	// This can be useful for invalid or zero values.
	Skip() bool
	// Interface provides the underlying Go value.
	// This is needed for basic Go types, e.g. bool, int, string, etc.
	Interface() interface{}
}

// Of provides a Source from any given value.
// Handles Source directly, otherwise defaults to goSource
// which handles reflect.Value as well.
func Of(i interface{}) Source {
	if val, ok := i.(Source); ok {
		return val
	}
	return &goSource{val: valueOf(i)}
}

// valueOf provides the reflection value of the given Go value
// even if the Go value is a reflection value.
func valueOf(i interface{}) reflect.Value {
	if val, ok := i.(reflect.Value); ok {
		return val
	}
	return reflect.ValueOf(i)
}

// goSource satisfies Source for Go values.
type goSource struct {
	val reflect.Value
}

func (v *goSource) Kind() reflect.Kind {
	return v.val.Kind()
}

func (v *goSource) Elem() Source {
	return &goSource{val: v.val.Elem()}
}

func (v *goSource) FieldByName(name string) Source {
	return &goSource{val: v.val.FieldByName(name)}
}

func (v *goSource) Len() int {
	return v.val.Len()
}

func (v *goSource) Index(i int) Source {
	return &goSource{val: v.val.Index(i)}
}

func (v *goSource) Pointer() uintptr {
	return v.val.Pointer()
}

func (v *goSource) MapRange() MapIter {
	return &GoMapIter{it: v.val.MapRange()}
}

func (v *goSource) Skip() bool {
	return !v.val.IsValid() || v.val.IsZero()
}

func (v *goSource) Interface() interface{} {
	return v.val.Interface()
}

var _ Source = (*goSource)(nil)

// MapIter provides a way to iterate over maps types.
type MapIter interface {
	Next() bool
	Key() Source
	Value() Source
}

// GoMapIter satisfies MapIter.
type GoMapIter struct {
	it *reflect.MapIter
}

func (m *GoMapIter) Next() bool {
	return m.it.Next()
}

func (m *GoMapIter) Key() Source {
	return &goSource{val: m.it.Key()}
}

func (m *GoMapIter) Value() Source {
	return &goSource{val: m.it.Value()}
}

var _ MapIter = (*GoMapIter)(nil)
