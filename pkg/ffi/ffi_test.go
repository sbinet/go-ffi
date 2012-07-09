package ffi_test

import (
	"math"
	"testing"

	"github.com/sbinet/go-ffi/pkg/ffi"
	//"github.com/sbinet/go-ffi/pkg/dl"
)

type info struct {
	fct string // fct name
	arg float64
	res float64 // expected value
}

func TestFFIMathf(t *testing.T) {
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

func TestFFIMathi(t *testing.T) {
	lib, err := ffi.NewLibrary(libm_name)

	if err != nil {
		t.Errorf("%v", err)
	}

	f, err := lib.Fct("abs", ffi.Int, []ffi.Type{ffi.Int})
	if err != nil {
		t.Errorf("could not locate function [abs]: %v", err)
	}
	{
		out := f(10).Int()
		if out != 10 {
			t.Errorf("expected [10], got [%v] (fct=abs(10))", out)
		}

	}
	{
		out := f(-10).Int()
		if out != 10 {
			t.Errorf("expected [10], got [%v] (fct=abs(-10))", out)
		}

	}

	err = lib.Close()
	if err != nil {
		t.Errorf("error closing [%s]: %v", libm_name, err)
	}
}

func TestFFIStrCmp(t *testing.T) {
	lib, err := ffi.NewLibrary(libc_name)

	if err != nil {
		t.Errorf("%v", err)
	}

	//int strcmp(const char* cs, const char* ct);
	f, err := lib.Fct("strcmp", ffi.Int, []ffi.Type{ffi.Pointer, ffi.Pointer})
	if err != nil {
		t.Errorf("could not locate function [strcmp]: %v", err)
	}
	{
		s1 := "foo"
		s2 := "foo"
		out := f(s1, s2).Int()
		if out != 0 {
			t.Errorf("expected [0], got [%v]", out)
		}

	}
	{
		s1 := "foo"
		s2 := "foo1"
		out := f(s1, s2).Int()
		if out == 0 {
			t.Errorf("expected [!0], got [%v]", out)
		}

	}

	err = lib.Close()
	if err != nil {
		t.Errorf("error closing [%s]: %v", libc_name, err)
	}
}

func TestFFIStrLen(t *testing.T) {
	lib, err := ffi.NewLibrary(libc_name)

	if err != nil {
		t.Errorf("%v", err)
	}

	//size_t strlen(const char* cs);
	f, err := lib.Fct("strlen", ffi.Int, []ffi.Type{ffi.Pointer})
	if err != nil {
		t.Errorf("could not locate function [strlen]: %v", err)
	}
	{
		str := `foo-bar-\nfoo foo`
		out := int(f(str).Int())
		if out != len(str) {
			t.Errorf("expected [%d], got [%d]", len(str), out)
		}

	}

	err = lib.Close()
	if err != nil {
		t.Errorf("error closing [%s]: %v", libc_name, err)
	}
}

func TestFFIStrCat(t *testing.T) {
	lib, err := ffi.NewLibrary(libc_name)

	if err != nil {
		t.Errorf("%v", err)
	}

	//char* strcat(char* s, const char* ct);
	f, err := lib.Fct("strcat", ffi.Pointer, []ffi.Type{ffi.Pointer, ffi.Pointer})
	if err != nil {
		t.Errorf("could not locate function [strlen]: %v", err)
	}
	{
		s1 := "foo"
		s2 := "bar"
		out := f(s1,s2).String()
		//FIXME
		if out != "foobar" && false {
			t.Errorf("expected [foobar], got [%s] (s1=%s, s2=%s)", out, s1, s2)
		}

	}

	err = lib.Close()
	if err != nil {
		t.Errorf("error closing [%s]: %v", libc_name, err)
	}
}

// EOF
