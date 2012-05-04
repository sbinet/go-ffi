package ffi_test

import (
	"math"
	"testing"

	"bitbucket.org/binet/go-ffi/pkg/ffi"
	//"bitbucket.org/binet/go-ffi/pkg/dl"
)

type info struct {
	fct string // fct name
	arg float32
	res float32 // expected value
}

func TestFFIMath(t *testing.T) {
	lib, err := ffi.NewLibrary(libm_name)

	if err != nil {
		t.Errorf("%v", err)
	}
	
	tests := []info{
		{"cos", 0., 1.},
		{"cos", math.Pi/2., 0.},
	}

	for _,info := range tests {
		f, err := lib.Fct(info.fct)
		if err != nil {
			t.Errorf("could not locate function [%s]: %v", info.fct, err)
		}
		f(info.arg)
		
	}

	err = lib.Close()
	if err != nil {
		t.Errorf("error closing [%s]: %v", libm_name, err)
	}
}


// EOF
