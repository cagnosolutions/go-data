package compare

import (
	"golang.org/x/sys/cpu"
)

//go:noescape
func CompareASM(a, b [16]byte) int

// Compare16 dynamically picks the best available method
func Compare16(a, b [16]byte) bool {
	if cpu.X86.HasAVX2 {
		return Compare16AVX2(&a, &b)
	}
	if cpu.X86.HasSSE42 {
		return Compare16SSE42(&a, &b)
	}
	return Compare16Fallback(a, b)
}

// Optimized Go fallback
func Compare16Fallback(a, b [16]byte) bool {
	for i := 0; i < 16; i++ {
		if a[i] < b[i] {
			return true
		} else if a[i] > b[i] {
			return false
		}
	}
	return false
}

// Assembly functions
func Compare16AVX2(a, b *[16]byte) bool
func Compare16SSE42(a, b *[16]byte) bool
