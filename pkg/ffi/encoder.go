package ffi

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

// NewEncoder returns a new encoder that writes to v.
func NewEncoder(v Value) *Encoder {
	enc := &Encoder{w: nil, cval: v}
	return enc
}

// An Encoder writes Go objects to a C-binary blob
type Encoder struct {
	w io.Writer
	cval Value
}

func (enc *Encoder) Encode(v interface{}) error {
	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)
	// make sure we can encode this value v into enc.cval
	ct := ctype_from_gotype(rt)
	//if ct.Name() != enc.cval.Type().Name() {
	if !is_compatible(ct, enc.cval.Type()) {
		return fmt.Errorf("ffi.Encode: can not encode go-type [%s] (with c-type [%s]) into c-type [%s]", rt.Name(), ct.Name(), enc.cval.Type().Name())
	}
	enc.w = NewWriter(enc.cval)
	return enc.encode_value(rv)
}

var g_int_sz = reflect.TypeOf(int(0)).Size()
var g_uint_sz = reflect.TypeOf(uint(0)).Size()

func (enc *Encoder) encode_value(v reflect.Value) (err error) {
	rt := v.Type()
	data := v.Interface()
	switch rt.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = binary.Write(enc.w, g_native_endian, data)

	case reflect.Int:
		switch g_int_sz {
		case 4:
			err = binary.Write(enc.w, g_native_endian, int32(v.Int()))
		case 8:
			err = binary.Write(enc.w, g_native_endian, int64(v.Int()))
		}
	case reflect.Uint:
		switch g_uint_sz {
		case 4:
			err = binary.Write(enc.w, g_native_endian, int32(v.Uint()))
		case 8:
			err = binary.Write(enc.w, g_native_endian, int64(v.Uint()))
		}

	case reflect.Float32:
		err = binary.Write(enc.w, g_native_endian, data)

	case reflect.Float64:
		err = binary.Write(enc.w, g_native_endian, data)

	case reflect.Array:
		for i := 0; i < rt.Len(); i++ {
			cval := enc.cval.Index(i)
			data = v.Index(i).Interface()
			w := NewWriter(cval)
			err = binary.Write(w, g_native_endian, data)
			if err != nil {
				return err
			}
		}
	case reflect.Ptr:
		panic("unimplemented")
	case reflect.Struct:
		for i := 0; i < rt.NumField(); i++ {
			field := enc.cval.Field(i)
			data = v.Field(i).Interface()
			w := NewWriter(field)
			err = binary.Write(w, g_native_endian, data)
			if err != nil {
				return err
			}
		}

	case reflect.String:
		panic("unimplemented")
	default:
		panic("unhandled kind [" + rt.Kind().String() + "]")
	}
	return err
}

// EOF
