package ffi

// #include <stdlib.h>
// #include "ffi.h"
// typedef void (*_go_ffi_fctptr_t)(void);
// static void _go_ffi_type_set_type(ffi_type *t, unsigned short type)
// {
//   t->type = type;
// }
// static unsigned short _go_ffi_type_get_type(ffi_type *t)
// {
//   return t->type;
// }
// static void _go_ffi_type_set_elements(ffi_type *t, void *elmts)
// {
//   t->elements = (ffi_type**)(elmts);
// }
// static void *_go_ffi_type_get_offset(void *data, unsigned n, ffi_type **types)
// {
//   size_t ofs = 0;
//   unsigned i;
//   unsigned short a;
//   for (i = 0; i < n && types[i]; i++) {
//     a = ofs % types[i]->alignment;
//     if (a != 0) ofs += types[i]->alignment-a;
//     ofs += types[i]->size;
//   }
//   if (i < n || !types[i])
//     return 0;
//   a = ofs % types[i]->alignment;
//   if (a != 0) ofs += types[i]->alignment-a;
//   return data+ofs;
// }
// static int _go_ffi_type_get_offsetof(ffi_type *t, int i)
// {
//   void *v;
//   void *data = NULL + 2; // make a non-null pointer
//   if (t->type != FFI_TYPE_STRUCT) return 0;
//   v = _go_ffi_type_get_offset(data, i, t->elements);
//   if (v) {
//     return (int)(v - data);
//   } else {
//     return 0;
//   }
//   return 0;
// }
import "C"

import (
	"fmt"
	"reflect"
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

type Kind uint

const (
	Void       Kind = C.FFI_TYPE_VOID
	Int             = C.FFI_TYPE_INT
	Float           = C.FFI_TYPE_FLOAT
	Double          = C.FFI_TYPE_DOUBLE
	LongDouble      = C.FFI_TYPE_LONGDOUBLE
	Uint8           = C.FFI_TYPE_UINT8
	Int8            = C.FFI_TYPE_SINT8
	Uint16          = C.FFI_TYPE_UINT16
	Int16           = C.FFI_TYPE_SINT16
	Uint32          = C.FFI_TYPE_UINT32
	Int32           = C.FFI_TYPE_SINT32
	Uint64          = C.FFI_TYPE_UINT64
	Int64           = C.FFI_TYPE_SINT64
	Struct          = C.FFI_TYPE_STRUCT
	Pointer         = C.FFI_TYPE_POINTER
)

func (k Kind) String() string {
	switch k {
	case Void:
		return "Void"
	case Int:
        return "Int"
	case Float:
        return "Float"
	case Double:
        return "Double"
	case LongDouble:
		return "LongDouble"
	case Uint8:
        return "Uint8"
	case Int8:
        return "Int8"
	case Uint16:
        return "Uint16"
	case Int16:
        return "Int16"
	case Uint32:
        return "Uint32"
	case Int32:
        return "Int32"
	case Uint64:
        return "Uint64"
	case Int64:
        return "Int64"
	case Struct:
        return "Struct"
	case Pointer:
        return "Pointer"
	}
	panic("unreachable")
}

// Type is a FFI type, describing functions' type arguments
type Type interface {
	cptr() *C.ffi_type

	// Name returns the type's name.
	Name() string

	// Size returns the number of bytes needed to store
	// a value of the given type.
	Size() uintptr

	// String returns a string representation of the type.
	String() string

	// Kind returns the specific kind of this type
	Kind() Kind

	// Align returns the alignment in bytes of a value of this type.
	Align() int

	// Field returns a struct type's i'th field.
	// It panics if the type's Kind is not Struct.
	// It panics if i is not in the range [0, NumField()).
	Field(i int) StructField

	// NumField returns a struct type's field count.
	// It panics if the type's Kind is not Struct.
	NumField() int
}

type cffi_type struct {
	n string
	c *C.ffi_type
}

func (t cffi_type) cptr() *C.ffi_type {
	return t.c
}

func (t cffi_type) Name() string {
	return t.n
}

func (t cffi_type) Size() uintptr {
	return uintptr(t.c.size)
}

func (t cffi_type) String() string {
	// fixme:
	return t.n
}

func (t cffi_type) Kind() Kind {
	return Kind(C._go_ffi_type_get_type(t.c))
}

func (t cffi_type) Align() int {
	return int(t.c.alignment)
}

func (t cffi_type) NumField() int {
	if t.Kind() != Struct {
		panic("ffi: NumField of non-struct type")
	}
	tt := (*cffi_struct)(unsafe.Pointer(&t))
	return tt.NumField()
}

func (t cffi_type) Field(i int) StructField {
	if t.Kind() != Struct {
		panic("ffi: Field of non-struct type")
	}
	tt := (*cffi_struct)(unsafe.Pointer(&t))
	return tt.Field(i)
}

var (
	C_void       Type = cffi_type{"void", &C.ffi_type_void}
	C_uchar           = cffi_type{"unsigned char", &C.ffi_type_uchar}
	C_char            = cffi_type{"char", &C.ffi_type_schar}
	C_ushort          = cffi_type{"unsigned short", &C.ffi_type_ushort}
	C_short           = cffi_type{"short", &C.ffi_type_sshort}
	C_uint            = cffi_type{"unsigned int", &C.ffi_type_uint}
	C_int             = cffi_type{"int", &C.ffi_type_sint}
	C_ulong           = cffi_type{"unsigned long", &C.ffi_type_ulong}
	C_long            = cffi_type{"long", &C.ffi_type_slong}
	C_uint8           = cffi_type{"uint8", &C.ffi_type_uint8}
	C_int8            = cffi_type{"int8", &C.ffi_type_sint8}
	C_uint16          = cffi_type{"uint16", &C.ffi_type_uint16}
	C_int16           = cffi_type{"int16", &C.ffi_type_sint16}
	C_uint32          = cffi_type{"uint32", &C.ffi_type_uint32}
	C_int32           = cffi_type{"int32", &C.ffi_type_sint32}
	C_uint64          = cffi_type{"uint64", &C.ffi_type_uint64}
	C_int64           = cffi_type{"int64", &C.ffi_type_sint64}
	C_float           = cffi_type{"float", &C.ffi_type_float}
	C_double          = cffi_type{"double", &C.ffi_type_double}
	C_longdouble      = cffi_type{"long double", &C.ffi_type_longdouble}
	C_pointer         = cffi_type{"pointer", &C.ffi_type_pointer}
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
			cargs[i] = args[i].cptr()
		}
		c_args = &cargs[0]
	}
	sc := C.ffi_prep_cif(&cif.c, C.ffi_abi(abi), c_nargs, rtype.cptr(), c_args)
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
	out := reflect.New(rtype_from_ffi(cif.c.rtype))
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

type StructField struct {
	Name   string  // Name is the field name
	Type   Type    // field type
	Offset uintptr // offset within struct, in bytes
}

type cffi_struct struct {
	cffi_type
	fields []StructField
}

func (t cffi_struct) NumField() int {
	return len(t.fields)
}

func (t cffi_struct) Field(i int) StructField {
	if i < 0 || i >= len(t.fields) {
		panic("ffi: field index out of range")
	}
	return t.fields[i]
}

type Field struct {
	Name   string  // Name is the field name
	Type   Type    // field type
}

// NewType creates a new ffi_type describing a C-struct
func NewType(name string, fields []Field) (Type, error) {
	c := C.ffi_type{}
	t := cffi_struct{
		cffi_type: cffi_type{n: name, c: &c},
		fields:    make([]StructField, len(fields)),
	}
	t.cffi_type.c.size = 0
	t.cffi_type.c.alignment = 0
	C._go_ffi_type_set_type(t.cptr(), C.FFI_TYPE_STRUCT)

	var c_fields **C.ffi_type = nil
	if len(fields) > 0 {
		var cargs = make([]*C.ffi_type, len(fields)+1)
		for i, f := range fields {
			cargs[i] = f.Type.cptr()
		}
		cargs[len(fields)] = nil
		c_fields = &cargs[0]
	}
	C._go_ffi_type_set_elements(t.cptr(), unsafe.Pointer(c_fields))

	// initialize type (computes alignment and size)
	_, err := NewCif(DefaultAbi, t, nil)
	if err != nil {
		return nil, err
	}
	
	for i := 0; i < len(fields); i++ {
		//cft := C._go_ffi_type_get_element(t.cptr(), C.int(i))
		ff := fields[i]
		t.fields[i] = StructField{
			ff.Name, 
			TypeByName(ff.Type.Name()),
			uintptr(C._go_ffi_type_get_offsetof(t.cptr(), C.int(i))),
		}
	}
	return t, nil
}

// the global map of types
var g_types map[string]Type

// TypeByName returns a ffi.Type by name.
// Returns nil if no such type exists
func TypeByName(n string) Type {
	t, ok := g_types[n]
	if ok {
		return t
	}
	return nil
}

func init() {
	g_types = make(map[string]Type)

	// initialize all builtin types
	init_type := func(t Type) {
		n := t.Name()
		//fmt.Printf("ctype [%s] - size: %v...\n", n, t.Size())
		if _, ok := g_types[n]; ok {
			//fmt.Printf("ctypes [%s] already registered\n", n)
			return
		}
		//NewCif(DefaultAbi, t, nil)
		//fmt.Printf("ctype [%s] - size: %v\n", n, t.Size())
		g_types[n] = t
	}

	init_type(C_void)
	init_type(C_uchar)
	init_type(C_char)
	init_type(C_ushort)
	init_type(C_short)
	init_type(C_uint)
	init_type(C_int)
	init_type(C_ulong)
	init_type(C_long)
	init_type(C_uint8)
	init_type(C_int8)
	init_type(C_uint16)
	init_type(C_int16)
	init_type(C_uint32)
	init_type(C_int32)
	init_type(C_uint64)
	init_type(C_int64)
	init_type(C_float)
	init_type(C_double)
	init_type(C_longdouble)
	init_type(C_pointer)

}

// EOF
