package ffi_test

import (
	//"bytes"
	"encoding/binary"
	//"fmt"
	"reflect"
	"testing"

	"github.com/sbinet/go-ffi/pkg/ffi"
)

func TestGetSetBuiltinValue(t *testing.T) {

	{
		const val = 42
		for _, tt := range []struct {
			n   string
			t   ffi.Type
			val interface{}
		}{
			{"int", ffi.C_int, int64(val)},
			{"int8", ffi.C_int8, int64(val)},
			{"int16", ffi.C_int16, int64(val)},
			{"int32", ffi.C_int32, int64(val)},
			{"int64", ffi.C_int64, int64(val)},
		} {
			cval := ffi.New(tt.t)
			eq(t, tt.n, cval.Type().Name())
			eq(t, tt.t.Kind(), cval.Kind())
			eq(t, reflect.Zero(reflect.TypeOf(tt.val)).Int(), cval.Int())
			cval.SetInt(val)
			eq(t, tt.val, cval.Int())
		}
	}

	{
		const val = 42
		for _, tt := range []struct {
			n   string
			t   ffi.Type
			val interface{}
		}{
			{"unsigned int", ffi.C_uint, uint64(val)},
			{"uint8", ffi.C_uint8, uint64(val)},
			{"uint16", ffi.C_uint16, uint64(val)},
			{"uint32", ffi.C_uint32, uint64(val)},
			{"uint64", ffi.C_uint64, uint64(val)},
		} {
			cval := ffi.New(tt.t)
			eq(t, tt.n, cval.Type().Name())
			eq(t, tt.t.Kind(), cval.Kind())
			eq(t, reflect.Zero(reflect.TypeOf(tt.val)).Uint(), cval.Uint())
			cval.SetUint(val)
			eq(t, tt.val, cval.Uint())
		}
	}

	{
		const val = -66.0
		for _, tt := range []struct {
			n   string
			t   ffi.Type
			val interface{}
		}{
			{"float", ffi.C_float, float64(val)},
			{"double", ffi.C_double, float64(val)},
			//FIXME: Go has no equivalent for long double...
			//{"long double", ffi.C_longdouble, float128(val)},
		} {
			cval := ffi.New(tt.t)
			eq(t, tt.n, cval.Type().Name())
			eq(t, tt.t.Kind(), cval.Kind())
			eq(t, reflect.Zero(reflect.TypeOf(tt.val)).Float(), cval.Float())
			cval.SetFloat(val)
			eq(t, tt.val, cval.Float())
		}
	}

	{
		const val = -66
		cval := ffi.New(ffi.C_int64)
		cptr := cval.Addr()
		cval.SetInt(val)
		eq(t, int64(val), cval.Int())
		eq(t, int64(val), cptr.Elem().Int())
		cval.SetInt(0)
		eq(t, int64(0), cptr.Elem().Int())
		cptr.Elem().SetInt(val)
		eq(t, int64(val), cval.Int())
		eq(t, int64(val), cptr.Elem().Int())

	}
}

func TestGetSetArrayValue(t *testing.T) {

	{
		const val = 42
		for _, tt := range []struct {
			n   string
			len int
			t   ffi.Type
			val interface{}
		}{
			{"uint8[10]", 10, ffi.C_uint8, [10]uint8{}},
			{"uint16[10]", 10, ffi.C_uint16, [10]uint16{}},
			{"uint32[10]", 10, ffi.C_uint32, [10]uint32{}},
			{"uint64[10]", 10, ffi.C_uint64, [10]uint64{}},
		} {
			ctyp, err := ffi.NewArrayType(tt.len, tt.t)
			if err != nil {
				t.Errorf(err.Error())
			}
			cval := ffi.New(ctyp)
			eq(t, tt.n, cval.Type().Name())
			eq(t, ctyp.Kind(), cval.Kind())
			gtyp := reflect.TypeOf(tt.val)
			gval := reflect.New(gtyp).Elem()
			eq(t, gval.Len(), cval.Len())
			for i := 0; i < gval.Len(); i++ {
				eq(t, gval.Index(i).Uint(), cval.Index(i).Uint())
				gval.Index(i).SetUint(val)
				cval.Index(i).SetUint(val)
				eq(t, gval.Index(i).Uint(), cval.Index(i).Uint())
			}
		}
	}

	{
		const val = 42
		for _, tt := range []struct {
			n   string
			len int
			t   ffi.Type
			val interface{}
		}{
			{"int8[10]", 10, ffi.C_int8, [10]int8{}},
			{"int16[10]", 10, ffi.C_int16, [10]int16{}},
			{"int32[10]", 10, ffi.C_int32, [10]int32{}},
			{"int64[10]", 10, ffi.C_int64, [10]int64{}},
		} {
			ctyp, err := ffi.NewArrayType(tt.len, tt.t)
			if err != nil {
				t.Errorf(err.Error())
			}
			cval := ffi.New(ctyp)
			eq(t, tt.n, cval.Type().Name())
			eq(t, ctyp.Kind(), cval.Kind())
			gtyp := reflect.TypeOf(tt.val)
			gval := reflect.New(gtyp).Elem()
			eq(t, gval.Len(), cval.Len())
			for i := 0; i < gval.Len(); i++ {
				eq(t, gval.Index(i).Int(), cval.Index(i).Int())
				gval.Index(i).SetInt(val)
				cval.Index(i).SetInt(val)
				eq(t, gval.Index(i).Int(), cval.Index(i).Int())
			}
		}
	}

	{
		const val = -66.2
		for _, tt := range []struct {
			n   string
			len int
			t   ffi.Type
			val interface{}
		}{
			{"float[10]", 10, ffi.C_float, [10]float32{}},
			{"double[10]", 10, ffi.C_double, [10]float64{}},
			// FIXME: go has no long double equivalent
			//{"long double[10]", 10, ffi.C_longdouble, [10]float128{}},
		} {
			ctyp, err := ffi.NewArrayType(tt.len, tt.t)
			if err != nil {
				t.Errorf(err.Error())
			}
			cval := ffi.New(ctyp)
			eq(t, tt.n, cval.Type().Name())
			eq(t, ctyp.Kind(), cval.Kind())
			gtyp := reflect.TypeOf(tt.val)
			gval := reflect.New(gtyp).Elem()
			eq(t, gval.Len(), cval.Len())
			for i := 0; i < gval.Len(); i++ {
				eq(t, gval.Index(i).Float(), cval.Index(i).Float())
				gval.Index(i).SetFloat(val)
				cval.Index(i).SetFloat(val)
				eq(t, gval.Index(i).Float(), cval.Index(i).Float())
			}
		}
	}

}

func TestGetSetStructValue(t *testing.T) {

	const val = 42
	arr10, err := ffi.NewArrayType(10, ffi.C_int32)
	if err != nil {
		t.Errorf(err.Error())
	}

	ctyp, err := ffi.NewStructType(
		"struct_ssv",
		[]ffi.Field{
			{"F1", ffi.C_uint16},
			{"F2", arr10},
			{"F3", ffi.C_int32},
			{"F4", ffi.C_uint16},
		})
	eq(t, "struct_ssv", ctyp.Name())
	eq(t, ffi.Struct, ctyp.Kind())
	eq(t, 4, ctyp.NumField())

	cval := ffi.New(ctyp)
	eq(t, ctyp.Kind(), cval.Kind())
	eq(t, ctyp.NumField(), cval.NumField())
	eq(t, uint64(0), cval.Field(0).Uint())
	for i := 0; i < arr10.Len(); i++ {
		eq(t, int64(0), cval.Field(1).Index(i).Int())
	}
	eq(t, int64(0), cval.Field(2).Int())
	eq(t, uint64(0), cval.Field(3).Uint())

	// set everything to 'val'
	cval.Field(0).SetUint(val)
	for i := 0; i < arr10.Len(); i++ {
		cval.Field(1).Index(i).SetInt(val)
	}
	cval.Field(2).SetInt(val)
	cval.Field(3).SetUint(val)

	// test values back
	eq(t, uint64(val), cval.Field(0).Uint())
	for i := 0; i < arr10.Len(); i++ {
		eq(t, int64(val), cval.Field(1).Index(i).Int())
	}
	eq(t, int64(val), cval.Field(2).Int())
	eq(t, uint64(val), cval.Field(3).Uint())

	// test values back - by field name
	eq(t, uint64(val), cval.FieldByName("F1").Uint())
	for i := 0; i < arr10.Len(); i++ {
		eq(t, int64(val), cval.FieldByName("F2").Index(i).Int())
	}
	eq(t, int64(val), cval.FieldByName("F3").Int())
	eq(t, uint64(val), cval.FieldByName("F4").Uint())

}

func TestBinaryIO(t *testing.T) {

	order := g_native_endian
	{
		const val = 42
		for _, tt := range []struct {
			n   string
			t   ffi.Type
			val interface{}
		}{
			{"int", ffi.C_int, int64(val)},
			{"int8", ffi.C_int8, int64(val)},
			{"int16", ffi.C_int16, int64(val)},
			{"int32", ffi.C_int32, int64(val)},
			{"int64", ffi.C_int64, int64(val)},
		} {
			cval := ffi.New(tt.t)
			w := ffi.NewWriter(cval)
			err := binary.Write(w, order, tt.val)
			if err != nil {
				t.Errorf(err.Error())
			}
			eq(t, tt.val, cval.Int())
		}
	}

	{
		const val = 42
		for _, tt := range []struct {
			n   string
			t   ffi.Type
			val interface{}
		}{
			{"uint", ffi.C_uint, uint64(val)},
			{"uint8", ffi.C_uint8, uint64(val)},
			{"uint16", ffi.C_uint16, uint64(val)},
			{"uint32", ffi.C_uint32, uint64(val)},
			{"uint64", ffi.C_uint64, uint64(val)},
		} {
			cval := ffi.New(tt.t)
			w := ffi.NewWriter(cval)
			err := binary.Write(w, order, tt.val)
			if err != nil {
				t.Errorf(err.Error())
			}
			eq(t, tt.val, cval.Uint())
		}
	}

	{
		const val = 42.0
		for _, tt := range []struct {
			n   string
			t   ffi.Type
			val interface{}
		}{
			{"float", ffi.C_float, float32(val)},
			{"double", ffi.C_double, float64(val)},
		} {
			cval := ffi.New(tt.t)
			w := ffi.NewWriter(cval)
			err := binary.Write(w, order, tt.val)
			if err != nil {
				t.Errorf(err.Error())
			}
			eq(t, float64(val), cval.Float())
		}
	}

	{

		const val = 42
		ctyp, err := ffi.NewStructType(
			"struct_uints",
			[]ffi.Field{
			{"F1", ffi.C_uint8},
			{"F2", ffi.C_uint16},
			{"F3", ffi.C_uint32},
			{"F4", ffi.C_uint64},
		})
		eq(t, ffi.Struct, ctyp.Kind())
		eq(t, 4, ctyp.NumField())

		cval := ffi.New(ctyp)
		eq(t, ctyp.Kind(), cval.Kind())
		eq(t, ctyp.NumField(), cval.NumField())
		for i := 0; i<ctyp.NumField(); i++ {
			eq(t, uint64(0), cval.Field(i).Uint())
		}

		values := []interface{}{
			uint8(val),
			uint16(val),
			uint32(val),
			uint64(val),
		}

		for i := 0; i < cval.NumField(); i++ {
			field := cval.Field(i)
			w := ffi.NewWriter(field)
			err = binary.Write(w, order, values[i])
			if err != nil {
				t.Errorf(err.Error())
			}
		}

		// test values back
		eq(t, uint64(val), cval.Field(0).Uint())
		eq(t, uint64(val), cval.Field(1).Uint())
		eq(t, uint64(val), cval.Field(2).Uint())
		eq(t, uint64(val), cval.Field(3).Uint())
	}

	{

		const val = 42
		ctyp, err := ffi.NewStructType(
			"struct_ints",
			[]ffi.Field{
			{"F1", ffi.C_int8},
			{"F2", ffi.C_int16},
			{"F3", ffi.C_int32},
			{"F4", ffi.C_int64},
		})
		eq(t, ffi.Struct, ctyp.Kind())
		eq(t, 4, ctyp.NumField())

		cval := ffi.New(ctyp)
		eq(t, ctyp.Kind(), cval.Kind())
		eq(t, ctyp.NumField(), cval.NumField())
		for i := 0; i<ctyp.NumField(); i++ {
			eq(t, int64(0), cval.Field(i).Int())
		}

		values := []interface{}{
			uint8(val),
			uint16(val),
			uint32(val),
			uint64(val),
		}

		for i := 0; i < cval.NumField(); i++ {
			field := cval.Field(i)
			w := ffi.NewWriter(field)
			err = binary.Write(w, order, values[i])
			if err != nil {
				t.Errorf(err.Error())
			}
		}

		// test values back
		eq(t, int64(val), cval.Field(0).Int())
		eq(t, int64(val), cval.Field(1).Int())
		eq(t, int64(val), cval.Field(2).Int())
		eq(t, int64(val), cval.Field(3).Int())
	}

	{
		const val = 42
		arr10, err := ffi.NewArrayType(10, ffi.C_int32)
		if err != nil {
			t.Errorf(err.Error())
		}

		ctyp, err := ffi.NewStructType(
			"struct_ssv",
			[]ffi.Field{
			{"F1", ffi.C_uint16},
			{"F2", arr10},
			{"F3", ffi.C_int32},
			{"F4", ffi.C_uint16},
		})
		eq(t, ffi.Struct, ctyp.Kind())
		eq(t, 4, ctyp.NumField())

		cval := ffi.New(ctyp)
		eq(t, ctyp.Kind(), cval.Kind())
		eq(t, ctyp.NumField(), cval.NumField())
		eq(t, uint64(0), cval.Field(0).Uint())
		for i := 0; i < arr10.Len(); i++ {
			eq(t, int64(0), cval.Field(1).Index(i).Int())
		}
		eq(t, int64(0), cval.Field(2).Int())
		eq(t, uint64(0), cval.Field(3).Uint())

		values := []interface{}{
			uint16(val),
			// note we use an array
			[10]int32{val, val, val, val, val,
				val, val, val, val, val},
			int32(val),
			uint16(val),
		}

		for i, value := range values {
			field := cval.Field(i)
			w := ffi.NewWriter(field)
			err = binary.Write(w, order, value)
			if err != nil {
				t.Errorf(err.Error())
			}
		}
		// test values back
		eq(t, uint64(val), cval.Field(0).Uint())
		for i := 0; i < arr10.Len(); i++ {
			eq(t, int64(val), cval.Field(1).Index(i).Int())
		}
		eq(t, int64(val), cval.Field(2).Int())
		eq(t, uint64(val), cval.Field(3).Uint())
	}

	{
		const val = 42
		arr10, err := ffi.NewArrayType(10, ffi.C_int32)
		if err != nil {
			t.Errorf(err.Error())
		}

		ctyp, err := ffi.NewStructType(
			"struct_ssv",
			[]ffi.Field{
			{"F1", ffi.C_uint16},
			{"F2", arr10},
			{"F3", ffi.C_int32},
			{"F4", ffi.C_uint16},
		})
		eq(t, ffi.Struct, ctyp.Kind())
		eq(t, 4, ctyp.NumField())

		cval := ffi.New(ctyp)
		eq(t, ctyp.Kind(), cval.Kind())
		eq(t, ctyp.NumField(), cval.NumField())
		eq(t, uint64(0), cval.Field(0).Uint())
		for i := 0; i < arr10.Len(); i++ {
			eq(t, int64(0), cval.Field(1).Index(i).Int())
		}
		eq(t, int64(0), cval.Field(2).Int())
		eq(t, uint64(0), cval.Field(3).Uint())

		values := []interface{}{
			uint16(val),
			// note we use a slice
			[]int32{val, val, val, val, val,
				val, val, val, val, val},
			int32(val),
			uint16(val),
		}

		for i, value := range values {
			field := cval.Field(i)
			w := ffi.NewWriter(field)
			err = binary.Write(w, order, value)
			if err != nil {
				t.Errorf(err.Error())
			}
		}
		// test values back
		eq(t, uint64(val), cval.Field(0).Uint())
		for i := 0; i < arr10.Len(); i++ {
			eq(t, int64(val), cval.Field(1).Index(i).Int())
		}
		eq(t, int64(val), cval.Field(2).Int())
		eq(t, uint64(val), cval.Field(3).Uint())
	}
}

// EOF
