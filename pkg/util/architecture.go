package util

import (
	"fmt"
	"unsafe"
)

func sign() {
	var arr = [2]byte{0x01, 0x00}
	var x = *(*uint16)(unsafe.Pointer(&arr[0]))
	if x == 1 {
		fmt.Println("architecture is little endian")
		return
	}
	fmt.Println("architecture is big endian")
}
