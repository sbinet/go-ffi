package ffi_test

import (
	"encoding/binary"
)

// TODO: determine this at compile-time...
// in C, at runtime:
// int a = 0x12345678;
// unsigned char *c = (unsigned char*)(&a);
// if (*c == 0x78)
//    printf("little-endian\n");
// else
//    printf("big-endian\n");

// determine native endianness
var g_native_endian = binary.LittleEndian

// EOF
