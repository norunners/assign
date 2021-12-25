// Package assign assigns values of any source to Go values.
package assign

import (
	"reflect"
)

// ToFrom assigns a Source value to the given Go value with options.
// See Assigner.To for additional details.
func ToFrom(dst, src interface{}, options ...Option) error {
	return From(src, options...).To(dst)
}

// Assigner assigns values of any source to Go values.
type Assigner struct {
	src   Source
	tags  []string
	cycle bool
}

// From creates a new Assigner from the given source and options.
// The Source value is determined by the Of function from source.
// By default, the `assign` tag is used and cyclical path checks are enabled.
// See Option to change the defaults.
func From(src interface{}, options ...Option) *Assigner {
	a := &Assigner{
		src:   Of(src),
		tags:  []string{"assign"},
		cycle: true,
	}
	for _, option := range options {
		option(a)
	}
	return a
}

// To assigns the Source to the given Go value.
// The Go value can be any supported type but must be a pointer that is not nil
// or a reflect.Value of a pointer that is not nil.
// Multiple Go values can be assigned from the same Source.
// Struct fields that are not exported are ignored.
// A Go value may be partially assigned when an error occurs.
// See error.go for error type details.
func (a *Assigner) To(dst interface{}) error {
	dv := valueOf(dst)
	if dv.Kind() != reflect.Ptr {
		return newError(dv.Type(), a.src.Kind())
	}
	if dv.IsNil() {
		return newError(reflect.TypeOf(nil), a.src.Kind())
	}
	var md *metadata
	if a.cycle {
		md = &metadata{visited: map[uintptr]struct{}{
			dv.Pointer(): {},
		}}
	}
	return a.assignRecover(dv.Elem(), md)
}

// metadata is used to track cyclical paths.
type metadata struct {
	visited map[uintptr]struct{}
	cur     uintptr
}

// assignRecover recovers unexpected assign panics.
// Please report unexpected panics.
func (a *Assigner) assignRecover(dv reflect.Value, md *metadata) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = ErrorPanic{Rec: rec}
		}
	}()

	err = a.assign(dv, a.src, md)
	return
}

// assign recursively assigns to a value.
func (a *Assigner) assign(dv reflect.Value, sv Source, md *metadata) error {
	// CanSet of dst handles fields that are not exported.
	// Skip of src handles invalid or zero values.
	// All these cases are expected to be ignored without assignment.
	if !dv.CanSet() || sv.Skip() {
		return nil
	}
	// The visit logic of source handles circular paths.
	if a.visit(sv, md) {
		return ErrorCycle{
			Dst: dv.Type(),
			Src: sv.Kind(),
		}
	}

	if _, ok := elemSet[sv.Kind()]; ok {
		return a.assign(dv, sv.Elem(), md)
	}

	switch dk := dv.Kind(); dk {
	case reflect.Ptr:
		return a.assignPointer(dv, sv, md)
	case reflect.Struct:
		return a.assignStruct(dv, sv, md)
	case reflect.Map:
		return a.assignMap(dv, sv, md)
	case reflect.Slice:
		return a.assignSlice(dv, sv, md)
	case reflect.Array:
		return a.assignArray(dv, sv, md)
	default:
		return a.assignBasic(dv, sv)
	}
}

// assignPointer assigns to a pointer.
func (a *Assigner) assignPointer(dp reflect.Value, sp Source, md *metadata) error {
	if dp.IsNil() {
		dp.Set(reflect.New(dp.Type().Elem()))
	}
	return a.assign(dp.Elem(), sp, md)
}

// assignBasic assigns to a basic value.
func (a *Assigner) assignBasic(db reflect.Value, sb Source) error {
	sv := reflect.ValueOf(sb.Interface())
	dt := db.Type()
	if st := sv.Type(); !st.ConvertibleTo(dt) {
		return newError(dt, st.Kind())
	}
	db.Set(sv.Convert(dt))
	return nil
}

// assignStruct assigns to a struct.
func (a *Assigner) assignStruct(ds reflect.Value, ss Source, md *metadata) error {
	dt := ds.Type()
	if sk := ss.Kind(); sk != reflect.Struct {
		return newError(dt, sk)
	}
	n := ds.NumField()
	for i := 0; i < n; i++ {
		df := ds.Field(i)
		dn := a.nameOf(dt.Field(i))
		sf := ss.FieldByName(dn)
		if err := a.assign(df, sf, md); err != nil {
			return err
		}
	}
	return nil
}

// assignMap assigns to a map.
func (a *Assigner) assignMap(dm reflect.Value, sm Source, md *metadata) error {
	dt := dm.Type()
	if sk := sm.Kind(); sk != reflect.Map {
		return newError(dt, sk)
	}
	kt := dt.Key()
	vt := dt.Elem()
	if dm.IsNil() {
		dm.Set(reflect.MakeMapWithSize(dt, sm.Len()))
	}

	for mi := sm.MapRange(); mi.Next(); {
		dk := reflect.New(kt).Elem()
		sk := mi.Key()
		if err := a.assign(dk, sk, md); err != nil {
			return err
		}
		dv := reflect.New(vt).Elem()
		sv := mi.Value()
		if err := a.assign(dv, sv, md); err != nil {
			return err
		}
		dm.SetMapIndex(dk, dv)
	}
	return nil
}

// assignSlice assigns to a slice.
func (a *Assigner) assignSlice(ds reflect.Value, ss Source, md *metadata) error {
	dt := ds.Type()
	sk := ss.Kind()
	if _, ok := listSet[sk]; !ok {
		return newError(dt, sk)
	}
	if ds.IsNil() {
		n := ss.Len()
		ds.Set(reflect.MakeSlice(dt, n, n))
	}
	return a.assignList(ds, ss, md)
}

// assignArray assigns to an array.
func (a *Assigner) assignArray(da reflect.Value, sa Source, md *metadata) error {
	sk := sa.Kind()
	if _, ok := listSet[sk]; !ok {
		return newError(da.Type(), sk)
	}
	return a.assignList(da, sa, md)
}

// assignList assigns both slices and arrays to each other.
// Varying lengths are permitted.
func (a *Assigner) assignList(dl reflect.Value, sl Source, md *metadata) error {
	n := sl.Len()
	if dn := dl.Len(); n > dn {
		n = dn
	}
	for i := 0; i < n; i++ {
		de := dl.Index(i)
		se := sl.Index(i)
		if err := a.assign(de, se, md); err != nil {
			return err
		}
	}
	return nil
}

// visit track pointers and checks for cyclical paths.
// Disable calls to Source.Pointer with the WithoutCycle option.
func (a *Assigner) visit(v Source, md *metadata) bool {
	if !a.cycle {
		return false
	}
	// This case occurs when the Source is a kind that has no pointer.
	if _, ok := ptrSet[v.Kind()]; !ok {
		// Reset the current pointer as the Source has encountered
		// a value to be assigned.
		md.cur = 0
		return false
	}
	ptr := v.Pointer()
	// This case occurs when the Source has not yet been assigned
	// because the destination is still being traversed.
	// The pointer has already been visited since it equals the current pointer.
	if ptr == md.cur {
		return false
	}
	md.cur = ptr
	_, ok := md.visited[ptr]
	if !ok {
		md.visited[ptr] = struct{}{}
	}
	return ok
}

// nameOf returns the first name matched by tag key, otherwise the field name.
// The first and default tag key is `assign`, see WithTags option to include tag keys.
func (a *Assigner) nameOf(sf reflect.StructField) string {
	for _, tag := range a.tags {
		if name := sf.Tag.Get(tag); name != "" {
			return name
		}
	}
	return sf.Name
}

var (
	ptrSet = map[reflect.Kind]struct{}{
		reflect.Ptr:           {},
		reflect.Map:           {},
		reflect.Slice:         {},
		reflect.Chan:          {},
		reflect.Func:          {},
		reflect.UnsafePointer: {},
	}
	elemSet = map[reflect.Kind]struct{}{
		reflect.Ptr:       {},
		reflect.Interface: {},
	}
	listSet = map[reflect.Kind]struct{}{
		reflect.Slice: {},
		reflect.Array: {},
	}
)
