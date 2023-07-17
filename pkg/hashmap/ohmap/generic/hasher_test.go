package generic

import (
	"fmt"
	"hash/crc64"
	"hash/fnv"
	"log"
	"strings"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/hash/maphash"
)

func hashmap() {

	// initialize new map
	m := NewMap[string, any](32)

	// add key and value pairs
	m.Set("foo", 1)
	m.Set("bar", 2)
	m.Set("baz", 3)

	// find a value
	val, found := m.Get("foo")
	if !found || val != 1 {
		log.Printf("did not find value, or value not in map")
	}

	// remove a value
	oldval, removed := m.Del("foo")
	if !removed || oldval == 1 {
		log.Printf("did not find value, or value not in map")
	}

	// get all the map keys
	mapKeys := m.Keys()
	_ = mapKeys

	// get all the map values
	mapVals := m.Vals()
	_ = mapVals

	// get all the map entries that match the filter
	vals := m.Filter(
		func(key string) bool {
			// get all values from keys that start with "ba"
			return strings.HasPrefix(key, "ba")
		},
	)
	if len(vals) != 2 {
		log.Printf("did not find the correct values")
	}

	// range the map
	m.Range(
		func(key string, val any) bool {
			fmt.Printf("key=%q, val=%v\n", key, val)
			return true
		},
	)

}

func BenchmarkHasher64VsNoWrapper(b *testing.B) {

	benches := []struct {
		name string
		fn   func(b *testing.B)
	}{
		{
			"hasher (wrapped fnv64a)",
			func(b *testing.B) {
				h := NewHasher64[string](fnv.New64a())
				b.ResetTimer()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					h.HashKey2("foo")
				}
			},
		},
		{
			"hasher (wrapped maphash64)",
			func(b *testing.B) {
				h := NewHasher64[string](maphash.New64())
				b.ResetTimer()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					h.HashKey2("foo")
				}
			},
		},
		{
			"hasher (wrapped crc64)",
			func(b *testing.B) {
				h := NewHasher64[string](crc64.New(crc64.MakeTable(crc64.ECMA)))
				b.ResetTimer()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					h.HashKey2("foo")
				}
			},
		},
		{
			"fnv 64a",
			func(b *testing.B) {
				h := fnv.New64a()
				b.ResetTimer()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					h.Write([]byte("foo"))
					sum := h.Sum64()
					if sum < 0 {
						b.Fatalf("bad sum")
					}
					h.Reset()
				}
			},
		},
		{
			"maphash64",
			func(b *testing.B) {
				h := maphash.New64()
				b.ResetTimer()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					h.Write([]byte("foo"))
					sum := h.Sum64()
					if sum < 0 {
						b.Fatalf("bad sum")
					}
					h.Reset()
				}
			},
		},
		{
			"crc64",
			func(b *testing.B) {
				h := crc64.New(crc64.MakeTable(crc64.ECMA))
				b.ResetTimer()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					h.Write([]byte("foo"))
					sum := h.Sum64()
					if sum < 0 {
						b.Fatalf("bad sum")
					}
					h.Reset()
				}
			},
		},
	}

	for _, bench := range benches {
		b.Run(bench.name, bench.fn)
	}
}

func BenchmarkHasher32_Hash(b *testing.B) {
	type foo struct {
		ID   int
		Name string
	}
	h1 := NewHasher32[string](fnv.New32a())
	h2 := NewHasher32[foo](fnv.New32a())
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h1.HashKey("foo")
		h2.HashKey(foo{1, "bob"})
	}
}

func BenchmarkHasher64_Hash(b *testing.B) {
	type foo struct {
		ID   int
		Name string
	}
	h1 := NewHasher64[foo](fnv.New64a())
	h2 := NewHasher64[string](fnv.New64a())
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h1.HashKey(foo{1, "bob"})
		h2.HashKey("foo")
	}
}
