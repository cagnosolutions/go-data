package bytes

import (
	"sort"
	"testing"
)

const N = 1024 * 1024

type T = int64

var xForMakeCopy = make([]T, N)
var xForAppend = make([]T, N)
var xForAppendNil = make([]T, N)
var yForMakeCopy []T
var yForAppend []T
var yForAppendNil []T
var zForMakeCopy any
var zForAppend any
var zForAppendNil any

func Benchmark_MakeAndCopy(b *testing.B) {
	b.ReportAllocs()
	var z any
	for i := 0; i < b.N; i++ {
		yForMakeCopy = make([]T, N)
		copy(yForMakeCopy, xForMakeCopy)
		z = yForMakeCopy
	}
	zForMakeCopy = z
	_ = zForMakeCopy
}

func Benchmark_Append(b *testing.B) {
	b.ReportAllocs()
	var z any
	for i := 0; i < b.N; i++ {
		yForAppend = append(xForAppend[:0:0], xForAppend...)
		z = yForAppend
	}
	zForAppend = z
	_ = zForAppend
}

func Benchmark_AppendNil(b *testing.B) {
	b.ReportAllocs()
	var z any
	for i := 0; i < b.N; i++ {
		yForAppendNil = append([]T(nil), xForAppendNil...)
		z = yForAppendNil
	}
	zForAppendNil = z
	_ = zForAppendNil
}

var s1 = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28}
var s2 = []int{1, 2, 3, 4, 5, 6, 14, 7, 8, 9, 10, 0, 11, 21, 12, 13, 15, 16, 17, 19, 18, 20, 22, 23, 24, 25, 26, 27, 28}

func Benchmark_Sorted(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if !sort.IntsAreSorted(s1) {
			b.Fail()
		}
	}
}

func Benchmark_Sort(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sort.Ints(s2)
	}
}
