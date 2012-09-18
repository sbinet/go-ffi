package ffi

import (
	"fmt"
	"reflect"
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
	Array Kind = 255 + iota
	Slice
	String
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
	case Slice:
		return "Slice"
	case String:
		return "String"
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

	// GoType returns the reflect.Type this ffi.Type is mirroring
	// It returns nil if there is no such equivalent go type.
	GoType() reflect.Type

	// set_gotype sets the reflect.Type associated with this ffi.Type
	set_gotype(t reflect.Type)
}

type cffi_type struct {
	n  string
	c  *C.ffi_type
	rt reflect.Type
}

func (t *cffi_type) cptr() *C.ffi_type {
	return t.c
}

func (t *cffi_type) Name() string {
	return t.n
}

func (t *cffi_type) Size() uintptr {
	return uintptr(t.c.size)
}

func (t *cffi_type) String() string {
	// fixme:
	return t.n
}

func (t *cffi_type) Kind() Kind {
	return Kind(C._go_ffi_type_get_type(t.c))
}

func (t *cffi_type) Align() int {
	return int(t.c.alignment)
}

func (t *cffi_type) Len() int {
	if t.Kind() != Array {
		panic("ffi: Len of non-array type")
	}
	tt := (*cffi_array)(unsafe.Pointer(&t))
	return tt.Len()
}

func (t *cffi_type) Elem() Type {
	switch t.Kind() {
	case Array:
		tt := (*cffi_array)(unsafe.Pointer(&t))
		return tt.Elem()
	case Ptr:
		tt := (*cffi_ptr)(unsafe.Pointer(&t))
		return tt.Elem()
	case Slice:
		tt := (*cffi_slice)(unsafe.Pointer(&t))
		return tt.Elem()
	}
	panic("ffi: Elem of invalid type")
}

func (t *cffi_type) NumField() int {
	if t.Kind() != Struct {
		panic("ffi: NumField of non-struct type")
	}
	tt := (*cffi_struct)(unsafe.Pointer(&t))
	return tt.NumField()
}

func (t *cffi_type) Field(i int) StructField {
	if t.Kind() != Struct {
		panic("ffi: Field of non-struct type")
	}
	tt := (*cffi_struct)(unsafe.Pointer(&t))
	return tt.Field(i)
}

func (t *cffi_type) GoType() reflect.Type {
	return t.rt
}

func (t *cffi_type) set_gotype(rt reflect.Type) {
	t.rt = rt
}

var (
	C_void       Type = &cffi_type{"void", &C.ffi_type_void, nil}
	C_uchar           = &cffi_type{"unsigned char", &C.ffi_type_uchar, reflect.TypeOf(uint8(0))}
	C_char            = &cffi_type{"char", &C.ffi_type_schar, reflect.TypeOf(int8(0))}
	C_ushort          = &cffi_type{"unsigned short", &C.ffi_type_ushort, reflect.TypeOf(uint16(0))}
	C_short           = &cffi_type{"short", &C.ffi_type_sshort, reflect.TypeOf(int16(0))}
	C_uint            = &cffi_type{"unsigned int", &C.ffi_type_uint, reflect.TypeOf(uint(0))}
	C_int             = &cffi_type{"int", &C.ffi_type_sint, reflect.TypeOf(int(0))}
	C_ulong           = &cffi_type{"unsigned long", &C.ffi_type_ulong, reflect.TypeOf(uint64(0))}
	C_long            = &cffi_type{"long", &C.ffi_type_slong, reflect.TypeOf(int64(0))}
	C_uint8           = &cffi_type{"uint8", &C.ffi_type_uint8, reflect.TypeOf(uint8(0))}
	C_int8            = &cffi_type{"int8", &C.ffi_type_sint8, reflect.TypeOf(int8(0))}
	C_uint16          = &cffi_type{"uint16", &C.ffi_type_uint16, reflect.TypeOf(uint16(0))}
	C_int16           = &cffi_type{"int16", &C.ffi_type_sint16, reflect.TypeOf(int16(0))}
	C_uint32          = &cffi_type{"uint32", &C.ffi_type_uint32, reflect.TypeOf(uint32(0))}
	C_int32           = &cffi_type{"int32", &C.ffi_type_sint32, reflect.TypeOf(int32(0))}
	C_uint64          = &cffi_type{"uint64", &C.ffi_type_uint64, reflect.TypeOf(uint64(0))}
	C_int64           = &cffi_type{"int64", &C.ffi_type_sint64, reflect.TypeOf(int64(0))}
	C_float           = &cffi_type{"float", &C.ffi_type_float, reflect.TypeOf(float32(0.))}
	C_double          = &cffi_type{"double", &C.ffi_type_double, reflect.TypeOf(float64(0.))}
	C_longdouble      = &cffi_type{"long double", &C.ffi_type_longdouble, nil}
	C_pointer         = &cffi_type{"*", &C.ffi_type_pointer, reflect.TypeOf(nil)}
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

func (t *cffi_struct) NumField() int {
	return len(t.fields)
}

func (t *cffi_struct) Field(i int) StructField {
	if i < 0 || i >= len(t.fields) {
		panic("ffi: field index out of range")
	}
	return t.fields[i]
}

func (t *cffi_struct) set_gotype(rt reflect.Type) {
	t.cffi_type.rt = rt
}

type Field struct {
	Name string // Name is the field name
	Type Type   // field type
}

var g_id_ch chan int

// NewStructType creates a new ffi_type describing a C-struct
func NewStructType(name string, fields []Field) (Type, error) {
	if name == "" {
		// anonymous type...
		// generate some id.
		name = fmt.Sprintf("_ffi_anon_type_%d", <-g_id_ch)
	}
	if t := TypeByName(name); t != nil {
		// check the definitions are the same
		if t.NumField() != len(fields) {
			return nil, fmt.Errorf("ffi.NewStructType: inconsistent re-declaration of [%s]", name)
		}
		for i := range fields {
			if fields[i].Name != t.Field(i).Name {
				return nil, fmt.Errorf("ffi.NewStructType: inconsistent re-declaration of [%s] (field #%d name mismatch)", name, i)

			}
			if fields[i].Type != t.Field(i).Type {
				return nil, fmt.Errorf("ffi.NewStructType: inconsistent re-declaration of [%s] (field #%d type mismatch)", name, i)

			}
		}
		return t, nil
	}
	c := C.ffi_type{}
	t := &cffi_struct{
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

func (t *cffi_array) Kind() Kind {
	// FIXME: ffi has no concept of array (as they decay to pointers in C)
	//return Kind(C._go_ffi_type_get_type(t.c))
	return Array
}

func (t *cffi_array) Len() int {
	return t.len
}

func (t *cffi_array) Elem() Type {
	return t.elem
}

// NewArrayType creates a new ffi_type with the given size and element type.
func NewArrayType(sz int, elmt Type) (Type, error) {
	n := fmt.Sprintf("%s[%d]", elmt.Name(), sz)
	if t := TypeByName(n); t != nil {
		return t, nil
	}
	c := C.ffi_type{}
	t := &cffi_array{
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

func (t *cffi_ptr) Elem() Type {
	return t.elem
}

// NewPointerType creates a new ffi_type with the given element type
func NewPointerType(elmt Type) (Type, error) {
	n := elmt.Name() + "*"
	if t := TypeByName(n); t != nil {
		return t, nil
	}
	c := C.ffi_type{}
	t := &cffi_ptr{
		cffi_type: cffi_type{n: n, c: &c},
		elem:      elmt,
	}
	if elmt.GoType() != nil {
		t.cffi_type.rt = reflect.PtrTo(elmt.GoType())
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

type cffi_slice struct {
	cffi_type
	elem Type
}

func (t *cffi_slice) Kind() Kind {
	// FIXME: ffi has no concept of array (as they decay to pointers in C)
	//return Kind(C._go_ffi_type_get_type(t.c))
	return Slice
}

func (t *cffi_slice) Elem() Type {
	return t.elem
}

// NewSliceType creates a new ffi_type slice with the given element type
func NewSliceType(elmt Type) (Type, error) {
	n := elmt.Name() + "[]"
	if t := TypeByName(n); t != nil {
		return t, nil
	}
	c := C.ffi_type{}
	t := &cffi_slice{
		cffi_type: cffi_type{n: n, c: &c},
		elem:      elmt,
	}
	t.cffi_type.c.size = 0
	t.cffi_type.c.alignment = 0
	C._go_ffi_type_set_type(t.cptr(), C.FFI_TYPE_STRUCT)

	var c_fields **C.ffi_type = nil
	var cargs = make([]*C.ffi_type, 3+1)
	cargs[0] = C_pointer.cptr()    // ptr to C-array
	cargs[1] = C_int.cptr() // len -- FIXME: use int64 when go-1.1
	cargs[2] = C_int.cptr() // cap -- FIXME: use int64 when go-1.1
	cargs[3] = nil
	c_fields = &cargs[0]
	C._go_ffi_type_set_elements(t.cptr(), unsafe.Pointer(c_fields))

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

func ctype_from_gotype(rt reflect.Type) Type {
	var t Type

	switch rt.Kind() {
	case reflect.Int:
		t = C_int

	case reflect.Int8:
		t = C_int8

	case reflect.Int16:
		t = C_int16

	case reflect.Int32:
		t = C_int32

	case reflect.Int64:
		t = C_int64

	case reflect.Uint:
		t = C_uint

	case reflect.Uint8:
		t = C_uint8

	case reflect.Uint16:
		t = C_uint16

	case reflect.Uint32:
		t = C_uint32

	case reflect.Uint64:
		t = C_uint64

	case reflect.Float32:
		t = C_float

	case reflect.Float64:
		t = C_double

	case reflect.Array:
		et := ctype_from_gotype(rt.Elem())
		ct, err := NewArrayType(rt.Len(), et)
		if err != nil {
			panic("ffi: " + err.Error())
		}
		ct.set_gotype(rt)
		t = ct

	case reflect.Ptr:
		et := ctype_from_gotype(rt.Elem())
		ct, err := NewPointerType(et)
		if err != nil {
			panic("ffi: " + err.Error())
		}
		t = ct

	case reflect.Slice:
		et := ctype_from_gotype(rt.Elem())
		ct, err := NewSliceType(et)
		if err != nil {
			panic("ffi: " + err.Error())
		}
		ct.set_gotype(rt)
		t = ct

	case reflect.Struct:
		fields := make([]Field, rt.NumField())
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			fields[i] = Field{
				Name: field.Name,
				Type: ctype_from_gotype(field.Type),
			}
		}
		ct, err := NewStructType(rt.Name(), fields)
		if err != nil {
			panic("ffi: " + err.Error())
		}
		ct.set_gotype(rt)
		t = ct

	case reflect.String:
		panic("unimplemented")
	default:
		panic("unhandled kind [" + rt.Kind().String() + "]")
	}

	return t
}

// Associate creates a link b/w a ffi.Type and a reflect.Type to allow
// automatic conversions b/w these types.
func Associate(ct Type, rt reflect.Type) error {
	crt := ct.GoType()
	if crt != nil {
		if crt != rt {
			return fmt.Errorf("ffi.Associate: ffi.Type [%s] already associated to reflect.Type [%s]", ct.Name(), crt.Name())
		}
		return nil
	}

	ct.set_gotype(rt)
	if ct.GoType() != rt {
		panic("ffi.Associate: internal error")
	}
	return nil
}

// PtrTo returns the pointer type with element t.
// For example, if t represents type Foo, PtrTo(t) represents *Foo.
func PtrTo(t Type) Type {
	typ, err := NewPointerType(t)
	if err != nil {
		return nil
	}
	return typ
}

// TypeOf returns the ffi Type of the value in the interface{}.
// TypeOf(nil) returns nil
// TypeOf(reflect.Type) returns the ffi Type corresponding to the reflected value
func TypeOf(i interface{}) Type {
	switch typ := i.(type) {
	case reflect.Type:
		return ctype_from_gotype(typ)
	case reflect.Value:
		return ctype_from_gotype(typ.Type())
	default:
		rt := reflect.TypeOf(i)
		return ctype_from_gotype(rt)
	}
	panic("unreachable")
}

// is_compatible returns whether two ffi Types are binary compatible
func is_compatible(t1, t2 Type) bool {
	if t1.Kind() != t2.Kind() {
		//FIXME: test if it is int/intX and uint/uintX
		return false
	}
	switch t1.Kind() {
	case Struct:
		for i := 0; i < t1.NumField(); i++ {
			f1 := t1.Field(i)
			f2 := t2.Field(i)
			if !is_compatible(f1.Type, f2.Type) {
				return false
			}
		}
	case Array:
		if t1.Len() != t2.Len() {
			return false
		}
		et1 := t1.Elem()
		et2 := t2.Elem()
		if !is_compatible(et1, et2) {
			return false
		}
		return true

	case Ptr:
		et1 := t1.Elem()
		et2 := t2.Elem()
		if !is_compatible(et1, et2) {
			return false
		}
		return true

	case Slice:
		et1 := t1.Elem()
		et2 := t2.Elem()
		if !is_compatible(et1, et2) {
			return false
		}
		return true

	case String:
		panic("unimplemented: ffi.String")
	}
	return true
}

func init() {
	// init out id counter channel
	g_id_ch = make(chan int, 1)
	go func() {
		i := 0
		for {
			g_id_ch <- i
			i++
		}
	}()

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

// make sure ffi_types satisfy ffi.Type interface
var _ Type = (*cffi_type)(nil)
var _ Type = (*cffi_array)(nil)
var _ Type = (*cffi_ptr)(nil)
var _ Type = (*cffi_slice)(nil)
var _ Type = (*cffi_struct)(nil)

// EOF
