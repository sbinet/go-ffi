package ffi_test

import (
	"reflect"
	"testing"

	"github.com/sbinet/go-ffi/pkg/ffi"
)

func TestNewBuiltinValue(t *testing.T) {

	for _, tt := range []struct {
		n  string
		t  ffi.Type
		rt reflect.Type
	}{
		{"int8", ffi.C_int8, reflect.TypeOf(int8(0))},
	} {
		cval := ffi.New(tt.t)
		//gval := reflect.New(tt.rt)
		//println("gval:",tt.rt.Name(), gval.Name())
		eq(t, tt.rt.Name(), cval.Type().Name())
	}
}

// EOF
