package ffi

import (
	"fmt"
	"reflect"
)

// NewEncoder returns a new encoder that writes to v.
func NewEncoder(v Value) *Encoder {
	enc := &Encoder{cval: v}
	return enc
}

// An Encoder writes Go objects to a C-binary blob
type Encoder struct {
	cval Value
}

func (enc *Encoder) Encode(v interface{}) error {
	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)
	//fmt.Printf("::Encode: %v %v\n", rt.Name(), rv)
	// make sure we can encode this value v into enc.cval
	ct := ctype_from_gotype(rt)
	//if ct.Name() != enc.cval.Type().Name() {
	if !is_compatible(ct, enc.cval.Type()) {
		return fmt.Errorf("ffi.Encode: can not encode go-type [%s] (with c-type [%s]) into c-type [%s]", rt.Name(), ct.Name(), enc.cval.Type().Name())
	}
	return enc.encode_value(rv)
}

func (enc *Encoder) encode_value(v reflect.Value) (err error) {
	enc.cval.set_value(v)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("ffi.Encoder: %v", r)
		}
	}()
	return err
}

// EOF
