package ffi

import (
	"fmt"
	"unsafe"
)

// #include "ffi.h"
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

type Kind uint

const (
	Void       Kind = C.FFI_TYPE_VOID
	Int        Kind = C.FFI_TYPE_INT
	Float      Kind = C.FFI_TYPE_FLOAT
	Double     Kind = C.FFI_TYPE_DOUBLE
	LongDouble Kind = C.FFI_TYPE_LONGDOUBLE
	Uint8      Kind = C.FFI_TYPE_UINT8
	Int8       Kind = C.FFI_TYPE_SINT8
	Uint16     Kind = C.FFI_TYPE_UINT16
	Int16      Kind = C.FFI_TYPE_SINT16
	Uint32     Kind = C.FFI_TYPE_UINT32
	Int32      Kind = C.FFI_TYPE_SINT32
	Uint64     Kind = C.FFI_TYPE_UINT64
	Int64      Kind = C.FFI_TYPE_SINT64
	Struct     Kind = C.FFI_TYPE_STRUCT
	Ptr        Kind = C.FFI_TYPE_POINTER
	//FIXME
	Array Kind = 255
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
	case Ptr:
		return "Ptr"
	case Array:
		return "Array"
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

	// Len returns an array type's length
	// It panics if the type's Kind is not Array.
	Len() int

	// Elem returns a type's element type.
	// It panics if the type's Kind is not Array or Ptr
	Elem() Type

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

func (t cffi_type) Len() int {
	if t.Kind() != Array {
		panic("ffi: Len of non-array type")
	}
	tt := (*cffi_array)(unsafe.Pointer(&t))
	return tt.Len()
}

func (t cffi_type) Elem() Type {
	switch t.Kind() {
	case Array:
		tt := (*cffi_array)(unsafe.Pointer(&t))
		return tt.Elem()
	case Ptr:
		tt := (*cffi_ptr)(unsafe.Pointer(&t))
		return tt.Elem()
	}
	panic("ffi: Elem of invalid type")
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
	C_pointer         = cffi_type{"*", &C.ffi_type_pointer}
)

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
	Name string // Name is the field name
	Type Type   // field type
}

// NewStructType creates a new ffi_type describing a C-struct
func NewStructType(name string, fields []Field) (Type, error) {
	if t := TypeByName(name); t != nil {
		return t, nil
	}
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
	register_type(t)
	return t, nil
}

type cffi_array struct {
	cffi_type
	len  int
	elem Type
}

func (t cffi_array) Kind() Kind {
	// FIXME: ffi has no concept of array (as they decay to pointers in C)
	//return Kind(C._go_ffi_type_get_type(t.c))
	return Array
}

func (t cffi_array) Len() int {
	return t.len
}

func (t cffi_array) Elem() Type {
	return t.elem
}

// NewArrayType creates a new ffi_type with the given size and element type.
func NewArrayType(sz int, elmt Type) (Type, error) {
	n := fmt.Sprintf("%s[%d]", elmt.Name(), sz)
	if t := TypeByName(n); t != nil {
		return t, nil
	}
	c := C.ffi_type{}
	t := cffi_array{
		cffi_type: cffi_type{n: n, c: &c},
		len:       sz,
		elem:      elmt,
	}
	t.cffi_type.c.size = C.size_t(sz * int(elmt.Size()))
	t.cffi_type.c.alignment = C_pointer.c.alignment
	var c_fields **C.ffi_type = nil
	C._go_ffi_type_set_elements(t.cptr(), unsafe.Pointer(c_fields))
	C._go_ffi_type_set_type(t.cptr(), C.FFI_TYPE_POINTER)

	// initialize type (computes alignment and size)
	_, err := NewCif(DefaultAbi, t, nil)
	if err != nil {
		return nil, err
	}

	register_type(t)
	return t, nil
}

type cffi_ptr struct {
	cffi_type
	elem Type
}

func (t cffi_ptr) Elem() Type {
	return t.elem
}

// NewPointerType creates a new ffi_type with the given element type
func NewPointerType(elmt Type) (Type, error) {
	n := elmt.Name() + "*"
	if t := TypeByName(n); t != nil {
		return t, nil
	}
	c := C.ffi_type{}
	t := cffi_ptr{
		cffi_type: cffi_type{n: n, c: &c},
		elem:      elmt,
	}
	t.cffi_type.c.size = C_pointer.c.size
	t.cffi_type.c.alignment = C_pointer.c.alignment
	var c_fields **C.ffi_type = nil
	C._go_ffi_type_set_elements(t.cptr(), unsafe.Pointer(c_fields))
	C._go_ffi_type_set_type(t.cptr(), C.FFI_TYPE_POINTER)

	// initialize type (computes alignment and size)
	_, err := NewCif(DefaultAbi, t, nil)
	if err != nil {
		return nil, err
	}

	register_type(t)
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

func register_type(t Type) {
	g_types[t.Name()] = t
}

/*
func ctype_from_ffi(t *C.ffi_type) Type {
	switch t {
	case &C.ffi_type_void:
		return C_void
	case &C.ffi_type_pointer:
		return C_pointer
	case &C.ffi_type_uint:
		return C_uint
	case &C.ffi_type_sint:
		return C_int
	case &C.ffi_type_uint8:
		return C_uint8
	case &C.ffi_type_sint8:
		return C_int8
	case &C.ffi_type_uint16:
		return C_uint16
	case &C.ffi_type_sint16:
		return C_int16
	case &C.ffi_type_uint32:
		return C_uint32
	case &C.ffi_type_sint32:
		return C_int32
	case &C.ffi_type_uint64:
		return C_uint64
	case &C.ffi_type_sint64:
		return C_int64
	case &C.ffi_type_ulong:
		return C_ulong
	case &C.ffi_type_slong:
		return C_long
	case &C.ffi_type_float:
		return C_float
	case &C.ffi_type_double:
		return C_double
	case &C.ffi_type_longdouble:
		return C_longdouble
	}

	if C._go_ffi_type_get_type(t) == C.FFI_TYPE_STRUCT {

	}
	panic("unreachable")
}
*/

// PtrTo returns the pointer type with element t.
// For example, if t represents type Foo, PtrTo(t) represents *Foo.
func PtrTo(t Type) Type {
	typ, err := NewPointerType(t)
	if err != nil {
		return nil
	}
	return typ
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
