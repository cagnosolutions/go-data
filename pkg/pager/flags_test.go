package pager

import (
	"fmt"
	"testing"
)

func runGenUint8Flags() {
	flags := genUint8Flags()
	fmt.Printf("Generated %d uint8 flags:\n", len(flags))
	var max uint8
	for i, flag := range flags {
		max += flag
		fmt.Printf("%.3d\t\t0x%.2x\t%.8b\t(max=%.3d @ bit %d)\n", flag, flag, flag, max, i+1)
	}
}

func runGenUint16Flags() {
	flags := genUint16Flags()
	fmt.Printf("Generated %d uint16 flags:\n", len(flags))
	var max uint16
	for i, flag := range flags {
		max += flag
		fmt.Printf("%.5d\t\t0x%.4x\t%.16b\t(max=%.5d @ bit %d)\n", flag, flag, flag, max, i+1)
	}
}

func runGenUint32Flags() {
	flags := genUint32Flags()
	fmt.Printf("Generated %d uint32 flags:\n", len(flags))
	var max uint32
	for i, flag := range flags {
		max += flag
		fmt.Printf("%.10d\t\t0x%.8x\t%.32b\t(max=%.10d @ bit %d)\n", flag, flag, flag, max, i+1)
	}
}

func runGenUint64Flags() {
	flags := genUint64Flags()
	fmt.Printf("Generated %d uint64 flags:\n", len(flags))
	var max uint64
	for i, flag := range flags {
		max += flag
		fmt.Printf("%.20d\t\t0x%.16x\t%.64b\t(max=%.20d @ bit %d)\n", flag, flag, flag, max, i+1)
	}
}

func Test_GenUint8_Flags(t *testing.T) {
	runGenUint8Flags()
}

func Test_GenUint16_Flags(t *testing.T) {
	runGenUint16Flags()
}

func Test_GenUint32_Flags(t *testing.T) {
	runGenUint32Flags()
}

func Test_GenUint64_Flags(t *testing.T) {
	runGenUint64Flags()
}

const (
	// unique flags

	// uint8 set
	flU8x01 = 0x01
	flU8x02 = 0x02
	flU8x03 = 0x04
	flU8x04 = 0x08
	flU8x05 = 0x10
	flU8x06 = 0x20
	flU8x07 = 0x40
	flU8x08 = 0x80

	// uint16 set
	flU16x01 = 0x0001
	flU16x02 = 0x0002
	flU16x03 = 0x0004
	flU16x04 = 0x0008
	flU16x05 = 0x0010
	flU16x06 = 0x0020
	flU16x07 = 0x0040
	flU16x08 = 0x0080
	flU16x09 = 0x0100
	flU16x10 = 0x0200
	flU16x11 = 0x0400
	flU16x12 = 0x0800
	flU16x13 = 0x1000
	flU16x14 = 0x2000
	flU16x15 = 0x4000
	flU16x16 = 0x8000

	// uint32 set
	flU32x01 = 0x00000001
	flU32x02 = 0x00000002
	flU32x03 = 0x00000004
	flU32x04 = 0x00000008
	flU32x05 = 0x00000010
	flU32x06 = 0x00000020
	flU32x07 = 0x00000040
	flU32x08 = 0x00000080
	flU32x09 = 0x00000100
	flU32x10 = 0x00000200
	flU32x11 = 0x00000400
	flU32x12 = 0x00000800
	flU32x13 = 0x00001000
	flU32x14 = 0x00002000
	flU32x15 = 0x00004000
	flU32x16 = 0x00008000
	flU32x17 = 0x00010000
	flU32x18 = 0x00020000
	flU32x19 = 0x00040000
	flU32x20 = 0x00080000
	flU32x21 = 0x00100000
	flU32x22 = 0x00200000
	flU32x23 = 0x00400000
	flU32x24 = 0x00800000
	flU32x25 = 0x01000000
	flU32x26 = 0x02000000
	flU32x27 = 0x04000000
	flU32x28 = 0x08000000
	flU32x29 = 0x10000000
	flU32x30 = 0x20000000
	flU32x31 = 0x40000000
	flU32x32 = 0x80000000

	// uint64 set
	flU64x01 = 0x0000000000000001
	flU64x02 = 0x0000000000000002
	flU64x03 = 0x0000000000000004
	flU64x04 = 0x0000000000000008
	flU64x05 = 0x0000000000000010
	flU64x06 = 0x0000000000000020
	flU64x07 = 0x0000000000000040
	flU64x08 = 0x0000000000000080
	flU64x09 = 0x0000000000000100
	flU64x10 = 0x0000000000000200
	flU64x11 = 0x0000000000000400
	flU64x12 = 0x0000000000000800
	flU64x13 = 0x0000000000001000
	flU64x14 = 0x0000000000002000
	flU64x15 = 0x0000000000004000
	flU64x16 = 0x0000000000008000
	flU64x17 = 0x0000000000010000
	flU64x18 = 0x0000000000020000
	flU64x19 = 0x0000000000040000
	flU64x20 = 0x0000000000080000
	flU64x21 = 0x0000000000100000
	flU64x22 = 0x0000000000200000
	flU64x23 = 0x0000000000400000
	flU64x24 = 0x0000000000800000
	flU64x25 = 0x0000000001000000
	flU64x26 = 0x0000000002000000
	flU64x27 = 0x0000000004000000
	flU64x28 = 0x0000000008000000
	flU64x29 = 0x0000000010000000
	flU64x30 = 0x0000000020000000
	flU64x31 = 0x0000000040000000
	flU64x32 = 0x0000000080000000
	flU64x33 = 0x0000000100000000
	flU64x34 = 0x0000000200000000
	flU64x35 = 0x0000000400000000
	flU64x36 = 0x0000000800000000
	flU64x37 = 0x0000001000000000
	flU64x38 = 0x0000002000000000
	flU64x39 = 0x0000004000000000
	flU64x40 = 0x0000008000000000
	flU64x41 = 0x0000010000000000
	flU64x42 = 0x0000020000000000
	flU64x43 = 0x0000040000000000
	flU64x44 = 0x0000080000000000
	flU64x45 = 0x0000100000000000
	flU64x46 = 0x0000200000000000
	flU64x47 = 0x0000400000000000
	flU64x48 = 0x0000800000000000
	flU64x49 = 0x0001000000000000
	flU64x50 = 0x0002000000000000
	flU64x51 = 0x0004000000000000
	flU64x52 = 0x0008000000000000
	flU64x53 = 0x0010000000000000
	flU64x54 = 0x0020000000000000
	flU64x55 = 0x0040000000000000
	flU64x56 = 0x0080000000000000
	flU64x57 = 0x0100000000000000
	flU64x58 = 0x0200000000000000
	flU64x59 = 0x0400000000000000
	flU64x60 = 0x0800000000000000
	flU64x61 = 0x1000000000000000
	flU64x62 = 0x2000000000000000
	flU64x63 = 0x4000000000000000
	flU64x64 = 0x8000000000000000
)
