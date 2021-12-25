package assign

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAssign(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		src     Source
		options []Option
		exp     All
	}{
		{
			name: "all value",
			src:  Of(allValue),
			exp:  allValue,
		},
		{
			name: "all pointer",
			src:  Of(pallValue),
			exp:  pallValue,
		},
		{
			name:    "all value without cycle",
			src:     panicker{Source: Of(allValue)},
			options: []Option{WithoutCycle()},
			exp:     allValue,
		},
		{
			name:    "all pointer without cycle",
			src:     panicker{Source: Of(pallValue)},
			options: []Option{WithoutCycle()},
			exp:     pallValue,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			dst := All{}

			if err := ToFrom(&dst, test.src, test.options...); err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if diff := cmp.Diff(test.exp, dst); diff != "" {
				t.Errorf("(-expected +actual):\n%s\n%+v", diff, dst)
			}
		})
	}
}

func TestAssignFieldValues(t *testing.T) {
	t.Parallel()
	srcAll := reflect.ValueOf(allValue)
	dstAll := reflect.ValueOf(&All{}).Elem()
	n := srcAll.NumField()
	for i := 0; i < n; i++ {
		i := i
		typ := srcAll.Type()
		name := typ.Field(i).Name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			df := dstAll.Field(i)
			sf := srcAll.Field(i)

			if err := ToFrom(df.Addr(), sf); err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			dst := df.Interface()
			src := sf.Interface()
			if diff := cmp.Diff(src, dst); diff != "" {
				t.Errorf("(-expected +actual):\n%s\n%+v", diff, dst)
			}
		})
	}
}

func TestAssignFieldPointers(t *testing.T) {
	t.Parallel()
	srcAll := reflect.ValueOf(pallValue)
	dstAll := reflect.ValueOf(&All{}).Elem()
	n := srcAll.NumField()
	for i := 0; i < n; i++ {
		i := i
		typ := srcAll.Type()
		name := typ.Field(i).Name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			df := dstAll.Field(i)
			sf := srcAll.Field(i)

			if err := ToFrom(df.Addr(), sf); err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			dst := df.Interface()
			src := sf.Interface()
			if diff := cmp.Diff(src, dst); diff != "" {
				t.Errorf("(-expected +actual):\n%s\n%+v", diff, dst)
			}
		})
	}
}

func TestAssignSliceAndArray(t *testing.T) {
	t.Parallel()

	dst := ListsDestination{}
	src := ListsSource{
		SliceToArray:      []interface{}{"0", 1, 2.3},
		SliceToArrayShort: []interface{}{"0", 1, 2.3},
		SliceToArrayLong:  []interface{}{"0", 1},
		ArrayToSlice:      [3]interface{}{"0", 1, 2.3},
		ArrayToSliceShort: [3]interface{}{"0", 1},
	}
	exp := ListsDestination{
		SliceToArray:      [3]interface{}{"0", 1, 2.3},
		SliceToArrayShort: [2]interface{}{"0", 1},
		SliceToArrayLong:  [3]interface{}{"0", 1},
		ArrayToSlice:      []interface{}{"0", 1, 2.3},
		ArrayToSliceShort: []interface{}{"0", 1, nil},
	}

	if err := ToFrom(&dst, src); err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if diff := cmp.Diff(exp, dst); diff != "" {
		t.Errorf("(-expected +actual):\n%s%+v", diff, dst)
	}
}

func TestAssignNotExported(t *testing.T) {
	t.Parallel()
	src := Fields{
		Exported:    "0",
		notExported: 1,
	}
	dst := Fields{}
	if err := ToFrom(&dst, src); err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	exp := Fields{
		Exported: "0",
	}
	// A regular equals expression is used below because
	// cmp.Diff does not handle fields that are not exported.
	if exp != dst {
		t.Errorf("expected: %+v but found %+v", exp, dst)
	}
}

func TestAssignWithTags(t *testing.T) {
	t.Parallel()

	src := Tags{
		FooGoToBar: "one",
		BarGoToFoo: "two",
		BazGoToQux: "three",
		QuxGoToBaz: "four",
	}
	tests := []struct {
		name    string
		options []Option
		exp     Tags
	}{
		{
			name: "assign",
			exp: Tags{
				FooGoToBar: "two",
				BarGoToFoo: "one",
				BazGoToQux: "three",
				QuxGoToBaz: "four",
			},
		},
		{
			name:    "assign and json",
			options: []Option{WithTags("json")},
			exp: Tags{
				FooGoToBar: "two",
				BarGoToFoo: "one",
				BazGoToQux: "four",
				QuxGoToBaz: "three",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			dst := Tags{}

			if err := ToFrom(&dst, src, test.options...); err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if diff := cmp.Diff(test.exp, dst); diff != "" {
				t.Errorf("(-expected +actual):\n%s\n%+v", diff, dst)
			}
		})
	}
}

/*
func TestAssignWithoutCycle(t *testing.T) {
	t.Parallel()

	exp := &Small{Field: "1"}
	src := panicker{Source: Of(exp)}
	dst := &Small{}

	// Verify cyclical checks are disabled as Pointer does not panic.
	if err := ToFrom(dst, src, WithoutCycle()); err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if diff := cmp.Diff(exp, dst); diff != "" {
		t.Errorf("(-expected +actual):\n%s\n%+v", diff, dst)
	}
}
*/

func TestAssignErrorType(t *testing.T) {
	t.Parallel()

	fn := func() {}
	ch := make(chan struct{})
	tests := []struct {
		name string
		val  interface{}
	}{
		{
			name: "bool",
			val:  &allValue.Bool,
		},
		{
			name: "string",
			val:  &allValue.String,
		},
		{
			name: "slice",
			val:  &allValue.Slice,
		},
		{
			name: "map",
			val:  &allValue.Map,
		},
		{
			name: "struct",
			val:  &allValue.Small,
		},
		{
			name: "func",
			val:  &fn,
		},
		{
			name: "chan",
			val:  &ch,
		},
	}

	expErr := ErrorType{}
	for _, ti := range tests {
		for _, tj := range tests {
			if ti.name == tj.name {
				continue
			}
			ti, tj := ti, tj
			t.Run(fmt.Sprintf("assign to destination: %s from source: %s", ti.name, tj.name), func(t *testing.T) {
				t.Parallel()
				if err := ToFrom(ti.val, tj.val); !errors.As(err, &expErr) {
					t.Errorf("expected type: %T but found: %T", expErr, err)
				}
			})
			t.Run(fmt.Sprintf("assign to destination: %s from source: %s", tj.name, ti.name), func(t *testing.T) {
				t.Parallel()
				if err := ToFrom(tj.val, ti.val); !errors.As(err, &expErr) {
					t.Errorf("expected type: %T but found: %T", expErr, err)
				}
			})
		}
	}
}

func TestAssignNotPointerAndNilPointer(t *testing.T) {
	t.Parallel()
	dstAll := reflect.ValueOf(&All{}).Elem()
	n := dstAll.NumField()

	for i := 0; i < n; i++ {
		i := i
		typ := dstAll.Type()
		name := typ.Field(i).Name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			df := dstAll.Field(i)
			dt := df.Type()
			if dt.Kind() == reflect.Ptr && df.IsNil() {
				dt = reflect.TypeOf(nil)
			}
			exp := newError(dt, reflect.Invalid)

			if err := ToFrom(df, nil); err != exp {
				t.Errorf("expected: %+v but found %+v", exp, err)
			}
		})
	}
}

func TestAssignErrorCycle(t *testing.T) {
	t.Parallel()
	srcCycle := reflect.ValueOf(newCycle())
	dstCycle := reflect.ValueOf(&Cycle{}).Elem()
	expErr := ErrorCycle{}

	n := srcCycle.NumField()
	for i := 0; i < n; i++ {
		i := i
		typ := srcCycle.Type()
		name := typ.Field(i).Name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			df := dstCycle.Field(i)
			sf := srcCycle.Field(i)

			if err := ToFrom(df.Addr(), sf); !errors.As(err, &expErr) {
				t.Errorf("expected type: %T but found: %T", expErr, err)
				return
			}
		})
	}
}

func TestAssignErrorPanic(t *testing.T) {
	t.Parallel()

	src := panicker{Source: Of(&allValue)}
	expErr := ErrorPanic{}
	// Verify cyclical checks are enabled as Pointer panics.
	if err := ToFrom(&All{}, src); !errors.As(err, &expErr) {
		t.Errorf("expected type: %T but found: %T", expErr, err)
		return
	}
}

type All struct {
	Bool    bool
	Int     int
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	Uint    uint
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	Uintptr uintptr
	Float32 float32
	Float64 float64

	PBool    *bool
	PInt     *int
	PInt8    *int8
	PInt16   *int16
	PInt32   *int32
	PInt64   *int64
	PUint    *uint
	PUint8   *uint8
	PUint16  *uint16
	PUint32  *uint32
	PUint64  *uint64
	PUintptr *uintptr
	PFloat32 *float32
	PFloat64 *float64

	String  string
	PString *string

	Map   map[string]Small
	MapP  map[string]*Small
	PMap  *map[string]Small
	PMapP *map[string]*Small

	EmptyMap map[string]Small
	NilMap   map[string]Small

	Slice   []Small
	SliceP  []*Small
	PSlice  *[]Small
	PSliceP *[]*Small

	EmptySlice []Small
	NilSlice   []Small

	StringSlice []string
	ByteSlice   []byte

	Array   [2]Small
	ArrayP  [2]*Small
	PArray  *[2]Small
	PArrayP *[2]*Small

	EmptyArray [0]Small
	ZeroArray  [2]Small

	StringArray [2]string
	ByteArray   [2]byte

	Small   Small
	PSmall  *Small
	PPSmall **Small

	Interface  interface{}
	PInterface *interface{}
}

type Small struct {
	Field string
}

var allValue = All{
	Bool:    true,
	Int:     2,
	Int8:    3,
	Int16:   4,
	Int32:   5,
	Int64:   6,
	Uint:    7,
	Uint8:   8,
	Uint16:  9,
	Uint32:  10,
	Uint64:  11,
	Uintptr: 12,
	Float32: 14.1,
	Float64: 15.1,
	String:  "16",
	Map: map[string]Small{
		"17": {Field: "field17"},
		"18": {Field: "field18"},
	},
	MapP: map[string]*Small{
		"19": {Field: "field19"},
	},
	EmptyMap:    map[string]Small{},
	Slice:       []Small{{Field: "field20"}, {Field: "field21"}},
	SliceP:      []*Small{{Field: "field22"}, {Field: "field23"}},
	EmptySlice:  []Small{},
	StringSlice: []string{"str24", "str25", "str26"},
	ByteSlice:   []byte{27, 28, 29},
	Array:       [2]Small{{Field: "field30"}, {Field: "field31"}},
	ArrayP:      [2]*Small{{Field: "field32"}, {Field: "field33"}},
	StringArray: [2]string{"str35", "str36"},
	ByteArray:   [2]byte{37, 38},
	Small:       Small{Field: "field39"},
	PSmall:      &Small{Field: "field40"},
	Interface:   41.2,
}

var pallValue = All{
	PBool:      &allValue.Bool,
	PInt:       &allValue.Int,
	PInt8:      &allValue.Int8,
	PInt16:     &allValue.Int16,
	PInt32:     &allValue.Int32,
	PInt64:     &allValue.Int64,
	PUint:      &allValue.Uint,
	PUint8:     &allValue.Uint8,
	PUint16:    &allValue.Uint16,
	PUint32:    &allValue.Uint32,
	PUint64:    &allValue.Uint64,
	PUintptr:   &allValue.Uintptr,
	PFloat32:   &allValue.Float32,
	PFloat64:   &allValue.Float64,
	PString:    &allValue.String,
	PMap:       &allValue.Map,
	PMapP:      &allValue.MapP,
	PSlice:     &allValue.Slice,
	PSliceP:    &allValue.SliceP,
	PArray:     &allValue.Array,
	PArrayP:    &allValue.ArrayP,
	PPSmall:    &allValue.PSmall,
	PInterface: &allValue.Interface,
}

type ListsSource struct {
	SliceToArray      []interface{}
	SliceToArrayShort []interface{}
	SliceToArrayLong  []interface{}
	ArrayToSlice      [3]interface{}
	ArrayToSliceShort [3]interface{}
}

type ListsDestination struct {
	SliceToArray      [3]interface{}
	SliceToArrayShort [2]interface{}
	SliceToArrayLong  [3]interface{}
	ArrayToSlice      []interface{}
	ArrayToSliceShort []interface{}
}

type Fields struct {
	Exported    string
	notExported int
}

type Tags struct {
	FooGoToBar string `assign:"BarGoToFoo",json:"FooGoToBar"`
	BarGoToFoo string `assign:"FooGoToBar"json:"BarGoToFoo"`
	BazGoToQux string `json:"QuxGoToBaz"`
	QuxGoToBaz string `json:"BazGoToQux"`
}

// Cycle contains fields that can have cyclical paths.
type Cycle struct {
	Struct *Cycle
	Map    map[struct{}]interface{}
	Slice  []interface{}
	Array  *[1]interface{}
}

// newCycle creates a new Cycle with each field having a cyclical path.
func newCycle() Cycle {
	cycle := Cycle{}
	cycle.Struct = &cycle
	cycle.Map = map[struct{}]interface{}{}
	cycle.Map[struct{}{}] = cycle.Map
	cycle.Slice = []interface{}{}
	cycle.Slice = append(cycle.Slice, &cycle.Slice)
	cycle.Array = &[1]interface{}{}
	cycle.Array[0] = cycle.Array
	return cycle
}

// panicker is a source Source which panics on Pointer.
type panicker struct {
	Source
}

// Pointer always panics.
func (panicker) Pointer() uintptr {
	panic("panicker")
}

var _ Source = (*panicker)(nil)
