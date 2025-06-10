package compare

import (
	"bytes"
	"math/rand"
	"testing"
)

// Sample keys for benchmarking
var keyA = randKey()
var keyB = randKey()

func randKey() [16]byte {
	var key [16]byte
	rand.Read(key[:])
	return key
}

// manual compare loop
func compareLoop(a, b [16]byte) int {
	for i := 0; i < 16; i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

func Benchmark_CompareAll(b *testing.B) {
	b.Run("Compare16_Manual", Benchmark_Compare16_Manual)
	b.Run("Compare16_BytesCompare", Benchmark_Compare16_BytesCompare)
	b.Run("Compare16_OptimizedASM", Benchmark_Compare16_OptimizedASM)
}

func Benchmark_Compare16_Manual(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = compareLoop(keyA, keyB)
	}
}

func Benchmark_Compare16_BytesCompare(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = bytes.Compare(keyA[:], keyB[:])
	}
}

func Benchmark_Compare16_OptimizedASM(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CompareASM(keyA, keyB)
	}
}
