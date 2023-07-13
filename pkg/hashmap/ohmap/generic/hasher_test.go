package generic

import (
	"hash/fnv"
	"testing"
)

func BenchmarkHasher32_Hash(b *testing.B) {
	h := NewHasher32[string](fnv.New32a())
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h.Hash("foo")
	}
}

func BenchmarkHasher64_Hash(b *testing.B) {
	type foo struct {
		ID   int
		Name string
	}

	h := NewHasher64[foo](fnv.New64a())
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h.Hash(foo{1, "bob"})
	}
}
