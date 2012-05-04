package dl_test

import (
	"testing"

	"bitbucket.org/binet/go-ffi/pkg/dl"
)

func TestDlOpenLibc(t *testing.T) {
	lib, err := dl.Open(libc_name, dl.Now)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = lib.Close()
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestDlSymLibc(t *testing.T) {
	lib, err := dl.Open(libc_name, dl.Now)

	if err != nil {
		t.Errorf("%v", err)
	}
	
	_, err = lib.Symbol("puts")
	if err != nil {
		t.Errorf("%v", err)
	}

	err = lib.Close()
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestDlOpenLibm(t *testing.T) {
	lib, err := dl.Open(libm_name, dl.Now)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = lib.Close()
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestDlSymLibm(t *testing.T) {
	lib, err := dl.Open(libm_name, dl.Now)

	if err != nil {
		t.Errorf("%v", err)
	}
	
	_, err = lib.Symbol("fabs")
	if err != nil {
		t.Errorf("%v", err)
	}

	err = lib.Close()
	if err != nil {
		t.Errorf("%v", err)
	}
}

// EOF
