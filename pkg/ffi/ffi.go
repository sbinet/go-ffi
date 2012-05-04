package ffi

// #include "ffi.h"
// typedef void (*_go_ffi_fctptr_t)(void);
import "C"

import (
	"fmt"
	"reflect"
	"unsafe"

	"bitbucket.org/binet/go-ffi/pkg/dl"
)

// Abi is the ffi abi of the local plateform
type Abi C.ffi_abi

const (
	FirstAbi Abi   = C.FFI_FIRST_ABI
	DefaultAbi Abi = C.FFI_DEFAULT_ABI
	LastAbi Abi = C.FFI_LAST_ABI
)

const (
	TrampolineSize = C.FFI_TRAMPOLINE_SIZE
	NativeRawApi   = C.FFI_NATIVE_RAW_API

	Closures          = C.FFI_CLOSURES
	TypeSmallStruct1B = C.FFI_TYPE_SMALL_STRUCT_1B
	TypeSmallStruct2B = C.FFI_TYPE_SMALL_STRUCT_2B
	TypeSmallStruct4B = C.FFI_TYPE_SMALL_STRUCT_4B
)

// 
type Type struct {
	c C.ffi_type
}

var (
	Void       = Type{C.ffi_type_void}
	Uchar      = Type{C.ffi_type_uchar}
	Schar      = Type{C.ffi_type_schar}
	Ushort     = Type{C.ffi_type_ushort}
	Sshort     = Type{C.ffi_type_sshort}
	Uint       = Type{C.ffi_type_uint}
	Sint       = Type{C.ffi_type_sint}
	Ulong      = Type{C.ffi_type_ulong}
	Slong      = Type{C.ffi_type_slong}
	Uint8      = Type{C.ffi_type_uint8}
	Sint8      = Type{C.ffi_type_sint8}
	Uint16     = Type{C.ffi_type_uint16}
	Sint16     = Type{C.ffi_type_sint16}
	Uint32     = Type{C.ffi_type_uint32}
	Sint32     = Type{C.ffi_type_sint32}
	Uint64     = Type{C.ffi_type_uint64}
	Sint64     = Type{C.ffi_type_sint64}
	Float      = Type{C.ffi_type_float}
	Double     = Type{C.ffi_type_double}
	LongDouble = Type{C.ffi_type_longdouble}
	Pointer    = Type{C.ffi_type_pointer}
)

type Status uint32

const (
	Ok         Status = C.FFI_OK
	BadTypedef Status = C.FFI_BAD_TYPEDEF
	BadAbi     Status = C.FFI_BAD_ABI
)

func (sc Status) String() string {
	switch sc {
	case Ok:
		return "FFI_OK"
	case BadTypedef:
		return "FFI_BAD_TYPEDEF"
	case BadAbi:
		return "FFI_BAD_ABI"
	}
	panic("unreachable")
}

// Arg is a ffi argument
type Arg struct {
	c C.ffi_arg
}

// SArg is a ffi argument
type SArg struct {
	c C.ffi_sarg
}

// Cif is the ffi call interface
type Cif struct {
	c C.ffi_cif
}

type FctPtr struct {
	//c unsafe.Pointer
	c C._go_ffi_fctptr_t
}

// NewCif creates a new ffi call interface object
func NewCif(abi Abi, rtype Type, args []Type) (*Cif, error) {
	cif := &Cif{}
	c_nargs := C.uint(len(args))
	var c_args **C.ffi_type = nil
	if len(args) > 0 {
		var cargs = make([]*C.ffi_type, len(args))
		for i,_ := range args {
			cargs[i] = &args[i].c
		}
		c_args = &cargs[0]
	}
	sc := C.ffi_prep_cif(&cif.c, C.ffi_abi(abi), c_nargs, &rtype.c, c_args)
	if sc != C.FFI_OK {
		return nil, fmt.Errorf("error while preparing cif (%s)", 
			Status(sc))
	}
	return cif, nil
}

// Call invokes the cif with the provided function pointer and arguments
func (cif *Cif) Call(fct FctPtr, args ...interface{}) (interface{}, error) {
	nargs := len(args)
	if nargs != int(cif.c.nargs) {
		return nil, fmt.Errorf("ffi: invalid number of arguments. expected '%d', got '%s'.",
			int(cif.c.nargs), nargs)
	}
	var c_args *unsafe.Pointer = nil
	if nargs > 0 {
		cargs := make([]unsafe.Pointer, nargs)
		for i,_ := range args {
			cargs[i] = to_voidptr(args[i])
		}
		c_args = &cargs[0]
	}
	var out interface{} = nil
	var c_out unsafe.Pointer = nil
	C.ffi_call(&cif.c, fct.c, c_out, c_args)
	return out, nil
}

// void ffi_call(ffi_cif *cif,
// 	      void (*fn)(void),
// 	      void *rvalue,
// 	      void **avalue);

// Closure models a ffi closure
type Closure struct {
	c C.ffi_closure
}


// utils ---

// to_voidptr wraps an interface value into an unsafe.Pointer
func to_voidptr(v interface{}) unsafe.Pointer {
	return nil
}

// Library is a dl-opened library holding the corresponding dl.Handle
type Library struct {
	handle dl.Handle
}

func NewLibrary(libname string) (lib Library, err error) {
	lib.handle, err = dl.Open(libname, dl.Now)
	return
}

func (lib Library) Close() error {
	return lib.handle.Close()
}

// Function is a dl-loaded function from a dl-opened library
type Function func(args ...interface{}) Value

// func Function(args ...interface{}) retvalue
// retvalue.Int()

type Value reflect.Value
func valueOf(o interface{}) Value {
	return Value(reflect.ValueOf(o))
}

type cfct struct {
	addr unsafe.Pointer
}

var nil_fct Function = func(args ...interface{}) Value {
	panic("ffi: nil_fct called")
}

func (lib Library) Fct(fctname string) (Function, error) {
	sym, err := lib.handle.Symbol(fctname)
	if err != nil {
		return nil_fct, err
	}

	f := cfct{unsafe.Pointer(sym)}
	fct := func(args ...interface{}) Value {
		if f.addr != nil {
		}
		return valueOf(0)
	}
	return Function(fct), nil
}

// EOF
