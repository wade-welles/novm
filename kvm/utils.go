package kvm

import (
	"unsafe"
)

func normalize(val uint64, size uint) uint64 {
	switch size {
	case 1:
		return val & 0xff
	case 2:
		return val & 0xffff
	case 4:
		return val & 0xffffffff
	}
	return val
}

func AlignBytes(data []byte) []byte {
	if uintptr(unsafe.Pointer(&data[0]))%PageSize != 0 {
		originalData := data
		data = make([]byte, len(originalData), len(originalData))
		copy(data, originalData)
	}
	return data
}
