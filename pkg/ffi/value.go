package ffi

import (
	"unsafe"
)

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
	ptr := unsafe.Pointer(&buf)
	//fl := flag(Ptr) << flagKindShift
	return Value{
		typ: typ,
		val: ptr,
	}
}

/* FIXME: we'd need some kind of runtime support...
// NewAt returns a Value representing a pointer to a value of the specified
// type, using p as that pointer.
func NewAt(typ Type, p unsafe.Pointer) Value {

}
*/

// Type returns v's type
func (v Value) Type() Type {
	return v.typ
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

// Elem returns the value that the pointer v points to.
// It panics if v's kind is not Ptr
func (v Value) Elem() Value {
	if v.typ.Kind() != Ptr {
		panic("ffi: call of ffi.Value.Elem on non-ptr Value")
	}
	// FIXME: would need to actually chase the pointer into C-world!!
	typ := v.typ.Elem()
	val := v.val
	val = *(*unsafe.Pointer)(val)
	return Value{typ: typ, val: val}
}

// EOF
