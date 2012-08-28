package ffi

import (
	"bytes"
	"io"
	"reflect"
	"runtime"
	"unsafe"
)

// methodName returns the name of the calling method,
// assumed to be two stack frames above.
func methodName() string {
	pc, _, _, _ := runtime.Caller(2)
	f := runtime.FuncForPC(pc)
	if f == nil {
		return "unknown method"
	}
	return f.Name()
}

// A ValueError occurs when a Value method is invoked on
// a Value that does not support it.  Such cases are documented
// in the description of each method.
type ValueError struct {
	Method string
	Kind   Kind
}

func (e *ValueError) Error() string {
	if e.Kind == 0 {
		return "ffi: call of " + e.Method + " on zero Value"
	}
	return "ffi: call of " + e.Method + " on " + e.Kind.String() + " Value"
}

// Value is the binary representation of an instance of type Type
type Value struct {
	// typ holds the type of the value represented by the Value
	typ Type

	// val holds the 1-word representation of the value.
	// If flag's flagIndir bit is set, then val is a pointer to the data.
	// Otherwise, val is a word holding the actual data.
	// When the data is smaller than a word, it begins at
	// the first byte (in the memory address sense) of val.
	// We use unsafe.Pointer so that the garbage collector
	// knows that val could be a pointer.
	val unsafe.Pointer

	// flag holds metadata about the value.
	// The lowest bits are flag bits:
	//	- flagIndir: val holds a pointer to the data
	//	- flagAddr: v.CanAddr is true (implies flagIndir)
	// The next five bits give the Kind of the value.
	// This repeats typ.Kind() except for method values.
	// The remaining 23+ bits give a method number for method values.
	// If flag.kind() != Func, code can assume that flagMethod is unset.
	// If typ.size > ptrSize, code can assume that flagIndir is set.
	//flag
}

type flag uintptr

const (
	flagRO flag = 1 << iota
	flagIndir
	flagAddr
	flagKindShift      = iota
	flagKindWidth      = 5 // there are 16 kinds
	flagKindMask  flag = 1<<flagKindWidth - 1
)

func (f flag) kind() Kind {
	return Kind((f >> flagKindShift) & flagKindMask)
}

// New returns a Value representing a pointer to a new zero value for
// the specified type.
func New(typ Type) Value {
	if typ == nil {
		panic("ffi: New(nil)")
	}
	buf := make([]byte, int(typ.Size()))
	ptr := unsafe.Pointer(&buf[0])
	v := Value{typ: typ, val: ptr}

	return v
}

// NewAt returns a Value representing a pointer to a value of the specified
// type, using p as that pointer.
func NewAt(typ Type, p unsafe.Pointer) Value {
	if typ == nil {
		panic("ffi: NewAt(nil)")
	}
	typ, err := NewPointerType(typ)
	if err != nil {
		return Value{}
	}
	v := Value{typ, p}
	return v
}

// mustBe panics if v's kind is not expected.
func (v Value) mustBe(expected Kind) {
	k := v.typ.Kind()
	if k != expected {
		panic("ffi: call of " + methodName() + " on " + k.String() + " Value")
	}
}

// Addr returns a pointer value representing the address of v.
// It panics if CanAddr() returns false.
// Addr is typically used to obtain a pointer to a struct field.
func (v Value) Addr() Value {
	typ := PtrTo(v.typ)
	if typ == nil {
		return Value{}
	}
	ptr := unsafe.Pointer(&v.val)
	return Value{typ, ptr}
}

// Buffer returns the underlying byte storage for this value.
func (v Value) Buffer() []byte {
	buf := make([]byte, 0)
	val := reflect.ValueOf(&buf)
	slice := (*reflect.SliceHeader)(unsafe.Pointer(val.Pointer()))
	slice.Len = int(v.typ.Size())
	slice.Data = uintptr(v.val)
	return buf
}

// Cap returns v's capacity.
// It panics if v's Kind is not Array.
func (v Value) Cap() int {
	k := v.Kind()
	switch k {
	case Array:
		return v.typ.Len()
	case Slice:
		return (*reflect.SliceHeader)(v.val).Cap
	}
	panic(&ValueError{"ffi.Value.Cap", k})
}

// Elem returns the value that the pointer v points to.
// It panics if v's kind is not Ptr
func (v Value) Elem() Value {
	v.mustBe(Ptr)
	typ := v.typ.Elem()
	val := v.val
	val = *(*unsafe.Pointer)(val)
	return Value{typ: typ, val: val}
}

// Field returns the i'th field of the struct v.
// It panics if v's Kind is not Struct or i is out of range.
func (v Value) Field(i int) Value {
	v.mustBe(Struct)
	tt := v.typ.(*cffi_struct)
	nfields := tt.NumField()
	if i < 0 || i >= nfields {
		panic("ffi: Field index out of range")
	}
	field := &tt.fields[i]
	typ := field.Type

	var val unsafe.Pointer
	// Indirect.  Just bump pointer.
	val = unsafe.Pointer(uintptr(v.val) + field.Offset)
	return Value{typ, val}
}

// FieldByIndex returns the nested field corresponding to index.
// It panics if v's Kind is not struct.
func (v Value) FieldByIndex(index []int) Value {
	v.mustBe(Struct)
	for i, x := range index {
		if i > 0 {
			if v.Kind() == Ptr && v.Elem().Kind() == Struct {
				v = v.Elem()
			}
		}
		v = v.Field(x)
	}
	return v
}

// FieldByName returns the struct field with the given name.
// It returns the zero Value if no field was found.
// It panics if v's Kind is not struct.
func (v Value) FieldByName(name string) Value {
	v.mustBe(Struct)
	for i := 0; i < v.typ.NumField(); i++ {
		if v.typ.Field(i).Name == name {
			return v.Field(i)
		}
	}
	return Value{}
	/*
		if f, ok := v.typ.FieldByName(name); ok {
			return v.FieldByIndex(f.Index)
		}
		return Value{}
	*/
}

// Float returns v's underlying value, as a float64.
// It panics if v's Kind is not Float or Double
func (v Value) Float() float64 {
	k := v.typ.Kind()
	switch k {
	case Float:
		return float64(*(*float32)(v.val))
	case Double:
		return *(*float64)(v.val)
	}
	panic(&ValueError{"ffi.Value.Float", k})
}

// Index returns v's i'th element.
// It panics if v's Kind is not Array or Slice or i is out of range.
func (v Value) Index(i int) Value {
	k := v.typ.Kind()
	switch k {
	case Array:
		tt := v.typ.(*cffi_array)
		if i < 0 || i > int(tt.Len()) {
			panic("ffi: array index out of range")
		}
		typ := tt.Elem()
		offset := uintptr(i) * typ.Size()

		var val unsafe.Pointer = unsafe.Pointer(uintptr(v.val) + offset)
		return Value{typ, val}
	case Slice:
		s := (*reflect.SliceHeader)(v.val)
		if i < 0 || i >= s.Len {
			panic("ffi: slice index out of range")
		}
		tt := v.typ.(*cffi_slice)
		typ := tt.Elem()
		offset := uintptr(i) * typ.Size()
		val := unsafe.Pointer(s.Data + offset)
		return Value{typ, val}
	}
	panic(&ValueError{"ffi.Value.Index", k})
}

// Int returns v's underlying value, as an int64.
// It panics if v's Kind is not Int, Int8, Int16, Int32, or Int64.
func (v Value) Int() int64 {
	k := v.typ.Kind()
	var p unsafe.Pointer = v.val
	switch k {
	case Int:
		return int64(*(*int)(p))
	case Int8:
		return int64(*(*int8)(p))
	case Int16:
		return int64(*(*int16)(p))
	case Int32:
		return int64(*(*int32)(p))
	case Int64:
		return int64(*(*int64)(p))
	}
	panic(&ValueError{"ffi.Value.Int", k})
}

// IsNil returns true if v is a nil value.
// It panics if v's Kind is Ptr.
func (v Value) IsNil() bool {
	v.mustBe(Ptr)
	ptr := v.val
	ptr = *(*unsafe.Pointer)(ptr)
	return ptr == nil
}

// IsValid returns true if v represents a value.
// It returns false if v is the zero Value.
// If IsValid returns false, all other methods except String panic.
// Most functions and methods never return an invalid value.
// If one does, its documentation states the conditions explicitly.
func (v Value) IsValid() bool {
	return v.val != nil
}

// Kind returns v's Kind.
func (v Value) Kind() Kind {
	return v.typ.Kind()
}

// Len returns v's length.
// It panics if v's Kind is not Array
func (v Value) Len() int {
	switch k := v.Kind(); k {
	case Array:
		tt := v.typ.(*cffi_array)
		return int(tt.Len())
	case Slice:
		//FIXME: make more robust
		//NOTE: we assume the layout of our "slice header" is the same than
		//      reflect.SliceHeader's...
		return (*reflect.SliceHeader)(v.val).Len
	default:
		panic(&ValueError{"ffi.Value.Len", k})
	}
	panic("unreachable")
}

// NumField returns the number of fields in the struct v.
// It panics if v's Kind is not Struct.
func (v Value) NumField() int {
	v.mustBe(Struct)
	return v.typ.NumField()
}

// SetFloat sets v's underlying value to x.
// It panics if v's Kind is not Float or Double, or if CanSet() is false.
func (v Value) SetFloat(x float64) {
	switch k := v.typ.Kind(); k {
	default:
		panic(&ValueError{"ffi.Value.SetFloat", k})
	case Float:
		*(*float32)(v.val) = float32(x)
	case Double:
		*(*float64)(v.val) = x
	}
}

// SetInt sets v's underlying value to x.
// It panics if v's Kind is not Int, Int8, Int16, Int32, or Int64, or if CanSet() is false.
func (v Value) SetInt(x int64) {
	//v.mustBeAssignable()
	switch k := v.typ.Kind(); k {
	default:
		panic(&ValueError{"ffi.Value.SetInt", k})
	case Int:
		*(*int)(v.val) = int(x)
	case Int8:
		*(*int8)(v.val) = int8(x)
	case Int16:
		*(*int16)(v.val) = int16(x)
	case Int32:
		*(*int32)(v.val) = int32(x)
	case Int64:
		*(*int64)(v.val) = x
	}
}

// SetPointer sets the unsafe.Pointer value v to x.
func (v Value) SetPointer(x unsafe.Pointer) {
	v.mustBe(Ptr)
	*(*unsafe.Pointer)(v.val) = x
}

// SetUint sets v's underlying value to x.
// It panics if v's Kind is not Int, Int8, Int16, Int32, or Int64, or if CanSet() is false.
func (v Value) SetUint(x uint64) {
	//v.mustBeAssignable()
	switch k := v.typ.Kind(); k {
	default:
		panic(&ValueError{"ffi.Value.SetUint", k})
	// case Uint:
	// 	*(*uint)(v.val) = uint(x)
	case Uint8:
		*(*uint8)(v.val) = uint8(x)
	case Uint16:
		*(*uint16)(v.val) = uint16(x)
	case Uint32:
		*(*uint32)(v.val) = uint32(x)
	case Uint64:
		*(*uint64)(v.val) = x
	}
}

// Slice returns a slice of v.
// It panics if v's Kind is not Array or Slice.
func (v Value) Slice(beg, end int) Value {
	var (
		cap  int
		typ  Type
		base unsafe.Pointer
	)
	switch k := v.Kind(); k {
	default:
		panic(&ValueError{"ffi.Value.Slice", k})
	case Array:
		tt := v.typ.(*cffi_array)
		cap = int(tt.Len())
		var err error
		typ, err = NewSliceType(tt.Elem())
		if err != nil {
			panic("ffi.Value.Slice: "+err.Error())
		}
		base = v.val
	case Slice:
		typ = v.typ.(*cffi_slice)
		s := (*reflect.SliceHeader)(v.val)
		base = unsafe.Pointer(s.Data)
		cap = s.Cap

	}
	if beg < 0 || end < beg || end > cap {
		panic("ffi.Value.Slice: slice index out of bounds")
	}

	// Declare slice so that gc can see the base pointer in it.
	var x []byte

	// Reinterpret as *SliceHeader to edit.
	s := (*reflect.SliceHeader)(unsafe.Pointer(&x))
	s.Data = uintptr(base) + uintptr(beg)*typ.Elem().Size()
	s.Len = end - beg
	s.Cap = cap - beg

	return Value{typ, unsafe.Pointer(&x)}
}

// Type returns v's type
func (v Value) Type() Type {
	return v.typ
}

// Uint returns v's underlying value, as a uint64.
// It panics if v's Kind is not Uint, Uintptr, Uint8, Uint16, Uint32, or Uint64.
func (v Value) Uint() uint64 {
	k := v.typ.Kind()
	var p unsafe.Pointer = v.val
	switch k {
	// case Uint:
	// 	return uint64(*(*uint)(p))
	case Uint8:
		return uint64(*(*uint8)(p))
	case Uint16:
		return uint64(*(*uint16)(p))
	case Uint32:
		return uint64(*(*uint32)(p))
	case Uint64:
		return uint64(*(*uint64)(p))
		// case Uintptr:
		// 	return uint64(*(*uintptr)(p))
	}
	panic(&ValueError{"ffi.Value.Uint", k})
}

// UnsafeAddr returns a pointer to v's data.
// It is for advanced clients that also import the "unsafe" package.
func (v Value) UnsafeAddr() uintptr {
	if v.typ == nil {
		panic("ffi: call of ffi.Value.UnsafeAddr on an invalid Value")
	}
	// FIXME: use flagAddr ??
	return uintptr(v.val)
}

// Indirect returns the value that v points to.
// If v is a nil pointer, Indirect returns a zero Value.
// If v is not a pointer, Indirect returns v.
func Indirect(v Value) Value {
	if v.typ.Kind() != Ptr {
		return v
	}
	return v.Elem()
}

// ValueOf returns a new Value initialized to the concrete value stored in
// the interface i.
// ValueOf(nil) returns the zero Value
func ValueOf(i interface{}) Value {
	if i == nil {
		return Value{}
	}
	v := Value{}
	rv := reflect.ValueOf(i)
	rt := rv.Type()
	switch rt.Kind() {
	case reflect.Int:
		v = New(C_int)
		v.SetInt(rv.Int())

	case reflect.Int8:
		v = New(C_int8)
		v.SetInt(rv.Int())

	case reflect.Int16:
		v = New(C_int16)
		v.SetInt(rv.Int())

	case reflect.Int32:
		v = New(C_int32)
		v.SetInt(rv.Int())

	case reflect.Int64:
		v = New(C_int64)
		v.SetInt(rv.Int())

	case reflect.Uint:
		v = New(C_uint)
		v.SetUint(rv.Uint())

	case reflect.Uint8:
		v = New(C_uint8)
		v.SetUint(rv.Uint())

	case reflect.Uint16:
		v = New(C_uint16)
		v.SetUint(rv.Uint())

	case reflect.Uint32:
		v = New(C_uint32)
		v.SetUint(rv.Uint())

	case reflect.Uint64:
		v = New(C_uint64)
		v.SetUint(rv.Uint())

	case reflect.Float32:
		v = New(C_float)
		v.SetFloat(rv.Float())

	case reflect.Float64:
		v = New(C_double)
		v.SetFloat(rv.Float())

	case reflect.Array:
		ct := ctype_from_gotype(rt)
		v = New(ct)
		enc := NewEncoder(v)
		err := enc.Encode(rv.Interface())
		if err != nil {
			panic("ffi: " + err.Error())
		}

	case reflect.Ptr:
		ct := ctype_from_gotype(rt)
		v = New(ct)
		enc := NewEncoder(v)
		err := enc.Encode(rv.Interface())
		if err != nil {
			panic("ffi: " + err.Error())
		}

	case reflect.Struct:
		ct := ctype_from_gotype(rt)
		v = New(ct)
		for i := 0; i < rt.NumField(); i++ {
			cfield := v.Field(i)
			goval := rv.Field(i)
			enc := NewEncoder(cfield)
			err := enc.Encode(goval.Interface())
			if err != nil {
				panic("ffi: " + err.Error())
			}
		}

	case reflect.String:
		panic("ffi.ValueOf: String unimplemented")

	case reflect.Slice:
		ct := ctype_from_gotype(rt)
		v = MakeSlice(ct, rv.Len(), rv.Cap())
		enc := NewEncoder(v)
		err := enc.Encode(rv.Interface())
		if err != nil {
			panic("ffi: " + err.Error())
		}

	default:
		panic("unhandled kind [" + rt.Kind().String() + "]")
	}

	return v
}

// MakeSlice creates a new zero-initialized slice value
// for the specified slice type, length, and capacity.
func MakeSlice(typ Type, len, cap int) Value {
	if typ.Kind() != Slice {
		panic("ffi.MakeSlice of non-slice type")
	}
	if len < 0 {
		panic("ffi.MakeSlice: negative len")
	}
	if cap < 0 {
		panic("ffi.MakeSlice: negative cap")
	}
	if len > cap {
		panic("ffi.MakeSlice: len > cap")
	}

	// Declare slice so that gc can see the base pointer in it.
	slice_len := uintptr(len) * typ.Elem().Size()
	slice_cap := uintptr(cap) * typ.Elem().Size()
	x := make([]byte, slice_len, slice_cap)

	// Reinterpret as *SliceHeader to edit.
	s := (*reflect.SliceHeader)(unsafe.Pointer(&x))
	arr_typ, err := NewArrayType(cap, typ.Elem())
	if err != nil {
		panic("ffi: " + err.Error())
	}
	arr_val := New(arr_typ)
	s.Data = uintptr(arr_val.UnsafeAddr())
	s.Len = len
	s.Cap = cap

	return Value{typ, unsafe.Pointer(&x)}
}

// grow_slice grows the slice s so that it can hold extra more values, 
// allocating more capacity if needed.
// It also returns the old and new slice lengths.
func grow_slice(s Value, extra int) (Value, int, int) {
	s.mustBe(Slice)

	i0 := s.Len()
	i1 := i0 + extra
	if i1 < i0 {
		panic("ffi.Append: slice overflow")
	}
	m := s.Cap()
	if i1 <= m {
		return s.Slice(0, i1), i0, i1
	}
	if m == 0 {
		m = extra
	} else {
		for m < i1 {
			if i0 < 1024 {
				m += m
			} else {
				m += m / 4
			}
		}
	}
	t := MakeSlice(s.Type(), i1, m)
	tx := (*[]byte)(unsafe.Pointer(t.val))
	sx := (*[]byte)(unsafe.Pointer(s.val))
	_ = copy(*tx, *sx)
	return t, i0, i1
}

// NewReader returns an io.Reader from a value, reading from its binary storage
func NewReader(v Value) io.Reader {
	return bytes.NewReader(v.Buffer())
}

type wbuffer struct {
	buf []byte
	idx int
}

func (w *wbuffer) Write(p []byte) (n int, err error) {
	n = copy(w.buf[w.idx:], p)
	w.idx += n
	return
}

// NewWriter returns an io.Writer from a value, writing into its binary storage
func NewWriter(v Value) io.Writer {
	buf := v.Buffer()
	return &wbuffer{buf: buf, idx: 0}
}

// EOF
