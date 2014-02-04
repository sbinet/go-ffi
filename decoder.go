package ffi

import (
	"fmt"
	"reflect"
)

// NewDecoder returns a new decoder that reads from v.
func NewDecoder(v Value) *Decoder {
	dec := &Decoder{cval: v}
	return dec
}

// A Decoder reads Go objects from a C-binary blob
type Decoder struct {
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
	return dec.decode_value(rv)
}

func (dec *Decoder) decode_value(v reflect.Value) (err error) {
	rt := v.Type()
	switch rt.Kind() {
	case reflect.Ptr:
		rt = rt.Elem()
		v = v.Elem()
	case reflect.Slice:
		v = v
	default:
		v = v.Elem()
	}

	v.Set(dec.cval.GoValue())
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("ffi.Decoder: %v", r)
		}
	}()
	return
}

// EOF
