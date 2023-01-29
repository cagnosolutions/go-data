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
