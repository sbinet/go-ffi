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
	FirstAbi   Abi = C.FFI_FIRST_ABI
	DefaultAbi Abi = C.FFI_DEFAULT_ABI
	LastAbi    Abi = C.FFI_LAST_ABI
)

const (
	TrampolineSize = C.FFI_TRAMPOLINE_SIZE
	NativeRawApi   = C.FFI_NATIVE_RAW_API

	//Closures          = C.FFI_CLOSURES
	//TypeSmallStruct1B = C.FFI_TYPE_SMALL_STRUCT_1B
	//TypeSmallStruct2B = C.FFI_TYPE_SMALL_STRUCT_2B
	//TypeSmallStruct4B = C.FFI_TYPE_SMALL_STRUCT_4B
)

// Type is a FFI type, describing functions' type arguments
type Type struct {
	c *C.ffi_type
}

var (
	Void       = Type{&C.ffi_type_void}
	Uchar      = Type{&C.ffi_type_uchar}
	Schar      = Type{&C.ffi_type_schar}
	Ushort     = Type{&C.ffi_type_ushort}
	Sshort     = Type{&C.ffi_type_sshort}
	Uint       = Type{&C.ffi_type_uint}
	Sint       = Type{&C.ffi_type_sint}
	Ulong      = Type{&C.ffi_type_ulong}
	Slong      = Type{&C.ffi_type_slong}
	Uint8      = Type{&C.ffi_type_uint8}
	Sint8      = Type{&C.ffi_type_sint8}
	Uint16     = Type{&C.ffi_type_uint16}
	Sint16     = Type{&C.ffi_type_sint16}
	Uint32     = Type{&C.ffi_type_uint32}
	Sint32     = Type{&C.ffi_type_sint32}
	Uint64     = Type{&C.ffi_type_uint64}
	Sint64     = Type{&C.ffi_type_sint64}
	Float      = Type{&C.ffi_type_float}
	Double     = Type{&C.ffi_type_double}
	LongDouble = Type{&C.ffi_type_longdouble}
	Pointer    = Type{&C.ffi_type_pointer}
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

// // Arg is a ffi argument
// type Arg struct {
// 	c C.ffi_arg
// }

// // SArg is a ffi argument
// type SArg struct {
// 	c C.ffi_sarg
// }

// Cif is the ffi call interface
type Cif struct {
	c C.ffi_cif
}

type FctPtr struct {
	c C._go_ffi_fctptr_t
}

// NewCif creates a new ffi call interface object
func NewCif(abi Abi, rtype Type, args []Type) (*Cif, error) {
	cif := &Cif{}
	c_nargs := C.uint(len(args))
	var c_args **C.ffi_type = nil
	if len(args) > 0 {
		var cargs = make([]*C.ffi_type, len(args))
		for i, _ := range args {
			cargs[i] = args[i].c
		}
		c_args = &cargs[0]
	}
	sc := C.ffi_prep_cif(&cif.c, C.ffi_abi(abi), c_nargs, rtype.c, c_args)
	if sc != C.FFI_OK {
		return nil, fmt.Errorf("error while preparing cif (%s)",
			Status(sc))
	}
	return cif, nil
}

// Call invokes the cif with the provided function pointer and arguments
func (cif *Cif) Call(fct FctPtr, args ...interface{}) (reflect.Value, error) {
	nargs := len(args)
	if nargs != int(cif.c.nargs) {
		return reflect.New(reflect.TypeOf(0)), fmt.Errorf("ffi: invalid number of arguments. expected '%d', got '%s'.",
			int(cif.c.nargs), nargs)
	}
	var c_args *unsafe.Pointer = nil
	if nargs > 0 {
		cargs := make([]unsafe.Pointer, nargs)
		for i, _ := range args {
			//fmt.Printf("[%d]: (%v)\n", i, args[i])
			cargs[i] = to_voidptr(args[i])
		}
		c_args = &cargs[0]
	}
	out := reflect.New(rtype_from_ffi(cif.c.rtype))
	var c_out unsafe.Pointer = unsafe.Pointer(out.Elem().UnsafeAddr())
	//println("...ffi_call...")
	C.ffi_call(&cif.c, fct.c, c_out, c_args)
	//fmt.Printf("...ffi_call...[done] [%v]\n",out.Elem())
	return out.Elem(), nil
}

func rtype_from_ffi(t *C.ffi_type) reflect.Type {
	switch t {
	case &C.ffi_type_uint:
		return reflect.TypeOf(uint(0))
	case &C.ffi_type_sint:
		return reflect.TypeOf(int(0))
	case &C.ffi_type_float:
		return reflect.TypeOf(float32(0))
	case &C.ffi_type_double:
		return reflect.TypeOf(float64(0))
	}
	panic("unreachable")
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
	t := reflect.TypeOf(v)
	rv := reflect.ValueOf(v)
	//println("-->",t.Kind().String())
	switch t.Kind() {
	case reflect.String:
		//fixme: memleak
		cstr := C.CString(rv.String())
		return unsafe.Pointer(&cstr)
	case reflect.Ptr:
		return unsafe.Pointer(rv.Elem().UnsafeAddr())
	case reflect.Float64:
		vv := rv.Float()
		rv = reflect.ValueOf(&vv)
		return unsafe.Pointer(rv.Elem().UnsafeAddr())
	}
	panic("to-voidptr: unreachable [" + t.Kind().String() + "]")
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
type Function func(args ...interface{}) reflect.Value

type cfct struct {
	addr unsafe.Pointer
}

var nil_fct Function = func(args ...interface{}) reflect.Value {
	panic("ffi: nil_fct called")
}

/*
func (lib Library) Fct(fctname string) (Function, error) {
	println("Fct(",fctname,")...")
	sym, err := lib.handle.Symbol(fctname)
	if err != nil {
		return nil_fct, err
	}

	addr := (C._go_ffi_fctptr_t)(unsafe.Pointer(sym))
	cif, err := NewCif(DefaultAbi, Double, []Type{Double})
	if err != nil {
		return nil_fct, err
	}

	fct := func(args ...interface{}) reflect.Value {
		println("...call.cif...")
		out, err := cif.Call(FctPtr{addr}, args...)
		if err != nil {
			panic(err)
		}
		println("...call.cif...[done]")
		return out
	}
	return Function(fct), nil
}
*/

func (lib Library) Fct(fctname string, rtype Type, argtypes []Type) (Function, error) {
	//println("Fct(",fctname,")...")
	sym, err := lib.handle.Symbol(fctname)
	if err != nil {
		return nil_fct, err
	}

	addr := (C._go_ffi_fctptr_t)(unsafe.Pointer(sym))
	cif, err := NewCif(DefaultAbi, rtype, argtypes)
	if err != nil {
		return nil_fct, err
	}

	fct := func(args ...interface{}) reflect.Value {
		//println("...call.cif...")
		out, err := cif.Call(FctPtr{addr}, args...)
		if err != nil {
			panic(err)
		}
		//println("...call.cif...[done]")
		return out
	}
	return Function(fct), nil
}

// EOF
