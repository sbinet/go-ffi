package ffi_test

import (
	"math"
	"testing"

	"bitbucket.org/binet/go-ffi/pkg/ffi"
	//"bitbucket.org/binet/go-ffi/pkg/dl"
)

type info struct {
	fct string // fct name
	arg float64
	res float64 // expected value
}

func TestFFIMath(t *testing.T) {
	lib, err := ffi.NewLibrary(libm_name)

	if err != nil {
		t.Errorf("%v", err)
	}

	tests := []info{
		{"cos", 0., math.Cos(0.)},
		{"cos", math.Pi / 2., math.Cos(math.Pi / 2.)},
		{"sin", 0., math.Sin(0.)},
		{"sin", math.Pi / 2., math.Sin(math.Pi / 2.)},
	}

	for _, info := range tests {
		f, err := lib.Fct(info.fct, ffi.Double, []ffi.Type{ffi.Double})
		if err != nil {
			t.Errorf("could not locate function [%s]: %v", info.fct, err)
		}
		out := f(info.arg).Float()
		if math.Abs(out-info.res) > 1e-16 {
			t.Errorf("expected [%v], got [%v] (fct=%v(%v))", info.res, out, info.fct, info.arg)
		}

	}

	err = lib.Close()
	if err != nil {
		t.Errorf("error closing [%s]: %v", libm_name, err)
	}
}

// EOF
