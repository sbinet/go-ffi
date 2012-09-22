package ffi_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/sbinet/go-ffi/pkg/ffi"
)

func TestBuiltinTypes(t *testing.T) {
	for _, table := range []struct {
		n  string
		t  ffi.Type
		rt reflect.Type
	}{
		{"unsigned char", ffi.C_uchar, reflect.TypeOf(byte(0))},
		{"char", ffi.C_char, reflect.TypeOf(byte(0))},

		{"int8", ffi.C_int8, reflect.TypeOf(int8(0))},
		{"uint8", ffi.C_uint8, reflect.TypeOf(uint8(0))},
		{"int16", ffi.C_int16, reflect.TypeOf(int16(0))},
		{"uint16", ffi.C_uint16, reflect.TypeOf(uint16(0))},
		{"int32", ffi.C_int32, reflect.TypeOf(int32(0))},
		{"uint32", ffi.C_uint32, reflect.TypeOf(uint32(0))},
		{"int64", ffi.C_int64, reflect.TypeOf(int64(0))},
		{"uint64", ffi.C_uint64, reflect.TypeOf(uint64(0))},

		{"float", ffi.C_float, reflect.TypeOf(float32(0))},
		{"double", ffi.C_double, reflect.TypeOf(float64(0))},
		//FIXME: use float128 when/if available
		//{"long double", ffi.C_longdouble, reflect.TypeOf(complex128(0))},

		{"*", ffi.C_pointer, reflect.TypeOf((*int)(nil))},
	} {
		if table.n != table.t.Name() {
			t.Errorf("expected [%s], got [%s]", table.n, table.t.Name())
		}
		if table.t.Size() != table.rt.Size() {
			t.Errorf("expected [%d], got [%d] (type=%q)", table.t.Size(), table.rt.Size(), table.n)
		}
	}
}

func TestNewStructType(t *testing.T) {

	arr10, err := ffi.NewArrayType(10, ffi.C_int32)
	if err != nil {
		t.Errorf(err.Error())
	}

	for _, table := range []struct {
		name    string
		fields  []ffi.Field
		size    uintptr
		offsets []uintptr
	}{
		{"struct_0",
			[]ffi.Field{{"a", ffi.C_int}},
			ffi.C_int.Size(),
			[]uintptr{0},
		},
		{"struct_1",
			[]ffi.Field{
				{"a", ffi.C_int},
				{"b", ffi.C_int},
			},
			ffi.C_int.Size() + ffi.C_int.Size(),
			[]uintptr{0, ffi.C_int.Size()},
		},
		{"struct_2",
			[]ffi.Field{
				{"F1", ffi.C_uint8},
				{"F2", ffi.C_int16},
				{"F3", ffi.C_int32},
				{"F4", ffi.C_uint8},
			},
			12,
			[]uintptr{0, 2, 4, 8},
		},
		{"struct_3",
			[]ffi.Field{
				{"F1", ffi.C_uint8},
				{"F2", arr10},
				{"F3", ffi.C_int32},
				{"F4", ffi.C_uint8},
			},
			56,
			[]uintptr{0, 8, 48, 52},
		},
	} {
		typ, err := ffi.NewStructType(table.name, table.fields)
		if err != nil {
			t.Errorf(err.Error())
		}
		eq(t, table.name, typ.Name())
		//eq(t, table.size, typ.Size())
		if table.size != typ.Size() {
			t.Errorf("expected size [%d] got [%d] (type=%q)", table.size, typ.Size(), table.name)
		}
		eq(t, len(table.offsets), typ.NumField())
		for i := 0; i < typ.NumField(); i++ {
			if table.offsets[i] != typ.Field(i).Offset {
				t.Errorf("type=%q field=%d: expected offset [%d]. got [%d]", table.name, i, table.offsets[i], typ.Field(i).Offset)
			}
			//eq(t, table.offsets[i], typ.Field(i).Offset)
		}
		eq(t, ffi.Struct, typ.Kind())
	}

	// test type mismatch
	n := "struct_type_err"
	st, err := ffi.NewStructType(n, []ffi.Field{{"a", ffi.C_int}})
	if err != nil {
		t.Errorf(err.Error())
	}
	{
		// check we get the exact same instance
		st_dup, err := ffi.NewStructType(n, []ffi.Field{{"a", ffi.C_int}})
		if err != nil {
			t.Errorf(err.Error())
		}
		if !reflect.DeepEqual(st_dup, st) {
			t.Errorf("NewStructType is not idem-potent")
		}
	}
	{
		_, err := ffi.NewStructType(
			n,
			[]ffi.Field{{"a", ffi.C_int}, {"b", ffi.C_int}})
		if err == nil {
			t.Errorf("failed to raise an error")
		}
		errmsg := fmt.Sprintf("ffi.NewStructType: inconsistent re-declaration of [%s]", n)
		if err.Error() != errmsg {
			t.Errorf("failed to detect number of fields differ: %v", err)
		}
	}
	{
		_, err := ffi.NewStructType(n, []ffi.Field{{"b", ffi.C_int}})
		if err == nil {
			t.Errorf("failed to raise an error")
		}
		errmsg := fmt.Sprintf("ffi.NewStructType: inconsistent re-declaration of [%s] (field #0 name mismatch)", n)
		if err.Error() != errmsg {
			t.Errorf("failed to detect field-name mismatch: %v", err)
		}
	}
	{
		_, err := ffi.NewStructType(n, []ffi.Field{{"a", ffi.C_uint}})
		if err == nil {
			t.Errorf("failed to raise an error")
		}
		errmsg := fmt.Sprintf("ffi.NewStructType: inconsistent re-declaration of [%s] (field #0 type mismatch)", n)
		if err.Error() != errmsg {
			t.Errorf("failed to detect field-type mismatch: %v", err)
		}
	}
}

func TestNewArrayType(t *testing.T) {

	s_t, err := ffi.NewStructType("s_0", []ffi.Field{{"a", ffi.C_int32}})
	if err != nil {
		t.Errorf(err.Error())
	}

	p_s_t, err := ffi.NewPointerType(s_t)
	if err != nil {
		t.Errorf(err.Error())
	}

	for _, table := range []struct {
		name string
		n    int
		elem ffi.Type
	}{
		{"uint8[10]", 10, ffi.C_uint8},
		{"uint16[10]", 10, ffi.C_uint16},
		{"uint32[10]", 10, ffi.C_uint32},
		{"uint64[10]", 10, ffi.C_uint64},
		{"int8[10]", 10, ffi.C_int8},
		{"int16[10]", 10, ffi.C_int16},
		{"int32[10]", 10, ffi.C_int32},
		{"int64[10]", 10, ffi.C_int64},

		{"float[10]", 10, ffi.C_float},
		{"double[10]", 10, ffi.C_double},

		{"s_0[10]", 10, s_t},
		{"s_0*[10]", 10, p_s_t},
	} {
		typ, err := ffi.NewArrayType(table.n, table.elem)
		if err != nil {
			t.Errorf(err.Error())
		}
		eq(t, table.name, typ.Name())
		eq(t, table.elem, typ.Elem())
		eq(t, uintptr(table.n)*table.elem.Size(), typ.Size())
		eq(t, table.n, typ.Len())
		eq(t, ffi.Array, typ.Kind())
	}
}

func TestNewSliceType(t *testing.T) {

	s_t, err := ffi.NewStructType("s_0", []ffi.Field{{"a", ffi.C_int32}})
	if err != nil {
		t.Errorf(err.Error())
	}

	p_s_t, err := ffi.NewPointerType(s_t)
	if err != nil {
		t.Errorf(err.Error())
	}

	for _, table := range []struct {
		name string
		elem ffi.Type
	}{
		{"uint8[]", ffi.C_uint8},
		{"uint16[]", ffi.C_uint16},
		{"uint32[]", ffi.C_uint32},
		{"uint64[]", ffi.C_uint64},
		{"int8[]", ffi.C_int8},
		{"int16[]", ffi.C_int16},
		{"int32[]", ffi.C_int32},
		{"int64[]", ffi.C_int64},

		{"float[]", ffi.C_float},
		{"double[]", ffi.C_double},

		{"s_0[]", s_t},
		{"s_0*[]", p_s_t},
	} {
		typ, err := ffi.NewSliceType(table.elem)
		if err != nil {
			t.Errorf(err.Error())
		}
		eq(t, table.name, typ.Name())
		eq(t, table.elem, typ.Elem())
		eq(t, (2*ffi.C_int.Size())+ffi.C_pointer.Size(), typ.Size())
		//eq(t, table.n, typ.Len())
		eq(t, ffi.Slice, typ.Kind())
	}
}

func TestNewPointerType(t *testing.T) {
	s_t, err := ffi.NewStructType("s_0", []ffi.Field{{"a", ffi.C_int32}})
	if err != nil {
		t.Errorf(err.Error())
	}

	p_s_t, err := ffi.NewPointerType(s_t)
	if err != nil {
		t.Errorf(err.Error())
	}

	for _, table := range []struct {
		name string
		elem ffi.Type
	}{
		{"int8*", ffi.C_int8},
		{"int16*", ffi.C_int16},
		{"int32*", ffi.C_int32},
		{"int64*", ffi.C_int64},
		{"uint8*", ffi.C_uint8},
		{"uint16*", ffi.C_uint16},
		{"uint32*", ffi.C_uint32},
		{"uint64*", ffi.C_uint64},

		{"float*", ffi.C_float},
		{"double*", ffi.C_double},

		{"s_0*", s_t},
		{"s_0**", p_s_t},
	} {
		typ, err := ffi.NewPointerType(table.elem)
		if err != nil {
			t.Errorf(err.Error())
		}
		eq(t, table.name, typ.Name())
		eq(t, table.elem, typ.Elem())
		eq(t, ffi.C_pointer.Size(), typ.Size())
		eq(t, ffi.Ptr, typ.Kind())
	}
}

// EOF
