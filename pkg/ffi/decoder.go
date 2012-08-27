package ffi

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

// NewDecoder returns a new decoder that reads from v.
func NewDecoder(v Value) *Decoder {
	dec := &Decoder{r: nil, cval: v}
	return dec
}

// A Decoder reads Go objects from a C-binary blob
type Decoder struct {
	r io.Reader
	cval Value
}

func (dec *Decoder) Decode(v interface{}) error {
	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)
	// FIXME ?
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	// make sure we can decode this value v from dec.cval
	ct := ctype_from_gotype(rt)
	if !is_compatible(ct, dec.cval.Type()) {
		return fmt.Errorf("ffi.Decode: can not decode go-type [%s] (with c-type [%s]) from c-type [%s]", rt.Name(), ct.Name(), dec.cval.Type().Name())
	}
	dec.r = NewReader(dec.cval)
	return dec.decode_value(rv)
}

func (dec *Decoder) decode_value(v reflect.Value) (err error) {
	rt := v.Type()
	switch rt.Kind() {
	case reflect.Ptr:
		rt = rt.Elem()
		v = v
		//v = v.Elem()
	case reflect.Slice:
		v = v
	}
	data := v.Interface()
	switch rt.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = binary.Read(dec.r, g_native_endian, data)

	case reflect.Int:
		switch g_int_sz {
		case 4:
			val := int32(0)
			err = binary.Read(dec.r, g_native_endian, &val)
			if err == nil {
				v.Elem().SetInt(int64(val))
			}
		case 8:
			val := int64(0)
			err = binary.Read(dec.r, g_native_endian, &val)
			if err == nil {
				v.Elem().SetInt(val)
			}
		}
	case reflect.Uint:
		switch g_uint_sz {
		case 4:
			val := uint32(0)
			err = binary.Read(dec.r, g_native_endian, &val)
			if err == nil {
				v.Elem().SetUint(uint64(val))
			}
		case 8:
			val := uint64(0)
			err = binary.Read(dec.r, g_native_endian, &val)
			if err == nil {
				v.Elem().SetUint(val)
			}
		}

	case reflect.Float32:
		err = binary.Read(dec.r, g_native_endian, data)

	case reflect.Float64:
		err = binary.Read(dec.r, g_native_endian, data)

	case reflect.Array:
		v := v.Elem()
		for i := 0; i < rt.Len(); i++ {
			cval := dec.cval.Index(i)
			data = v.Index(i).Addr().Interface()
			w := NewReader(cval)
			err = binary.Read(w, g_native_endian, data)
			if err != nil {
				return err
			}
		}
	case reflect.Ptr:
		panic("unimplemented")

	case reflect.Struct:
		v := v.Elem()
		for i := 0; i < rt.NumField(); i++ {
			field := dec.cval.Field(i)
			data = v.Field(i).Addr().Interface()
			w := NewReader(field)
			err = binary.Read(w, g_native_endian, data)
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
