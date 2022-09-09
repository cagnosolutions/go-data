package iters

import (
	"math/rand"
	"testing"
)

//go:run test

const NumItems int = 1000000

var struct_data = make(Collection[Data], NumItems)

type Data struct {
	foo int
	bar *Data
}

func InitData() {
	for i := 0; i < NumItems; i++ {
		struct_data[i] = Data{foo: rand.Int(), bar: nil}
	}
}

func BenchmarkCollection_CallbackIter(b *testing.B) {
	InitData()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var sum int
		struct_data.CallbackIter(
			func(val Data) bool {
				sum += val.foo
				return true
			},
		)
	}
}

func BenchmarkCollection_ClosureIter(b *testing.B) {
	InitData()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var sum int
		var val Data
		for it, more := struct_data.ClosureIter(); more; val, more = it() {
			sum += val.foo
		}
	}
}

func BenchmarkCollection_ChannelIter(b *testing.B) {
	InitData()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var sum int
		for val := range struct_data.ChannelIter(0) {
			sum += val.foo
		}
	}
}

func BenchmarkCollection_ChannelIterWithBuffer(b *testing.B) {
	InitData()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var sum int
		for val := range struct_data.ChannelIter(100) {
			sum += val.foo
		}
	}
}

func BenchmarkCollection_NewStatefulIter(b *testing.B) {
	InitData()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var sum int
		it := NewStatefulIter(struct_data)
		for it.Next() {
			sum += it.Value().foo
		}
	}
}
