package ffi

// #include <stdlib.h>
// #include "ffi.h"
// typedef void (*_go_ffi_fctptr_t)(void);
import "C"

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/sbinet/go-ffi/pkg/dl"
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
	c     C.ffi_cif
	rtype Type
	args  []Type
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
			cargs[i] = args[i].cptr()
		}
		c_args = &cargs[0]
	}
	sc := C.ffi_prep_cif(&cif.c, C.ffi_abi(abi), c_nargs, rtype.cptr(), c_args)
	if sc != C.FFI_OK {
		return nil, fmt.Errorf("error while preparing cif (%s)",
			Status(sc))
	}
	cif.rtype = rtype
	cif.args = args
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
			var carg unsafe.Pointer
			//fmt.Printf("[%d]: (%v)\n", i, args[i])
			t := reflect.TypeOf(args[i])
			rv := reflect.ValueOf(args[i])
			switch t.Kind() {
			case reflect.String:
				cstr := C.CString(args[i].(string))
				defer C.free(unsafe.Pointer(cstr))
				carg = unsafe.Pointer(&cstr)
			case reflect.Ptr:
				carg = unsafe.Pointer(rv.Elem().UnsafeAddr())
			case reflect.Float32:
				vv := args[i].(float32)
				rv = reflect.ValueOf(&vv)
				carg = unsafe.Pointer(rv.Elem().UnsafeAddr())
			case reflect.Float64:
				vv := args[i].(float64)
				rv = reflect.ValueOf(&vv)
				carg = unsafe.Pointer(rv.Elem().UnsafeAddr())
			case reflect.Int:
				vv := args[i].(int)
				rv = reflect.ValueOf(&vv)
				carg = unsafe.Pointer(rv.Elem().UnsafeAddr())
			case reflect.Int8:
				vv := args[i].(int8)
				rv = reflect.ValueOf(&vv)
				carg = unsafe.Pointer(rv.Elem().UnsafeAddr())
			case reflect.Int16:
				vv := args[i].(int16)
				rv = reflect.ValueOf(&vv)
				carg = unsafe.Pointer(rv.Elem().UnsafeAddr())
			case reflect.Int32:
				vv := args[i].(int32)
				rv = reflect.ValueOf(&vv)
				carg = unsafe.Pointer(rv.Elem().UnsafeAddr())
			case reflect.Int64:
				vv := args[i].(int64)
				rv = reflect.ValueOf(&vv)
				carg = unsafe.Pointer(rv.Elem().UnsafeAddr())
			case reflect.Uint:
				vv := args[i].(uint)
				rv = reflect.ValueOf(&vv)
				carg = unsafe.Pointer(rv.Elem().UnsafeAddr())
			case reflect.Uint8:
				vv := args[i].(uint8)
				rv = reflect.ValueOf(&vv)
				carg = unsafe.Pointer(rv.Elem().UnsafeAddr())
			case reflect.Uint16:
				vv := args[i].(uint16)
				rv = reflect.ValueOf(&vv)
				carg = unsafe.Pointer(rv.Elem().UnsafeAddr())
			case reflect.Uint32:
				vv := args[i].(uint32)
				rv = reflect.ValueOf(&vv)
				carg = unsafe.Pointer(rv.Elem().UnsafeAddr())
			case reflect.Uint64:
				vv := args[i].(uint64)
				rv = reflect.ValueOf(&vv)
				carg = unsafe.Pointer(rv.Elem().UnsafeAddr())
			}
			cargs[i] = carg
		}
		c_args = &cargs[0]
	}
	out := reflect.New(rtype_from_ffi(cif.rtype.cptr()))
	var c_out unsafe.Pointer = unsafe.Pointer(out.Elem().UnsafeAddr())
	//println("...ffi_call...")
	C.ffi_call(&cif.c, fct.c, c_out, c_args)
	//fmt.Printf("...ffi_call...[done] [%v]\n",out.Elem())
	return out.Elem(), nil
}

type go_void struct{}

func rtype_from_ffi(t *C.ffi_type) reflect.Type {
	switch t {
	case &C.ffi_type_void:
		return reflect.TypeOf(go_void{})
	case &C.ffi_type_pointer:
		return reflect.TypeOf(uintptr(0))
	case &C.ffi_type_uint:
		return reflect.TypeOf(uint(0))
	case &C.ffi_type_sint:
		return reflect.TypeOf(int(0))
	case &C.ffi_type_uint8:
		return reflect.TypeOf(uint8(0))
	case &C.ffi_type_sint8:
		return reflect.TypeOf(int8(0))
	case &C.ffi_type_uint16:
		return reflect.TypeOf(uint16(0))
	case &C.ffi_type_sint16:
		return reflect.TypeOf(int16(0))
	case &C.ffi_type_uint32:
		return reflect.TypeOf(uint32(0))
	case &C.ffi_type_sint32:
		return reflect.TypeOf(int32(0))
	case &C.ffi_type_uint64:
		return reflect.TypeOf(uint64(0))
	case &C.ffi_type_sint64:
		return reflect.TypeOf(int64(0))
	case &C.ffi_type_ulong:
		return reflect.TypeOf(uint64(0))
	case &C.ffi_type_slong:
		return reflect.TypeOf(int64(0))
	case &C.ffi_type_float:
		return reflect.TypeOf(float32(0))
	case &C.ffi_type_double:
		return reflect.TypeOf(float64(0))
	case &C.ffi_type_longdouble:
		// FIXME!!
		return reflect.TypeOf(complex128(0))
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

// Library is a dl-opened library holding the corresponding dl.Handle
type Library struct {
	handle dl.Handle
}

func get_lib_arch_name(libname string) string {
	fname := libname
	if !strings.HasPrefix(libname, g_lib_prefix) {
		fname = g_lib_prefix + libname
	}
	if !strings.HasSuffix(libname, g_lib_suffix) {
		fname = fname + g_lib_suffix
	}
	return fname
}

// NewLibrary takes the name of the library and returns a handle towards it.
// ex:
//   lib, err := ffi.NewLibrary("m") // loads libm
func NewLibrary(libname string) (lib Library, err error) {
	fname := get_lib_arch_name(libname)
	lib.handle, err = dl.Open(fname, dl.Now)
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
