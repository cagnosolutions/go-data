package generic

import (
	"fmt"
	"testing"
)

func TestRbTree_Scan(t *testing.T) {
	tree := NewTree[string, int]()
	for i := 0; i < 32; i++ {
		tree.Add(fmt.Sprintf("entry-%.3d", i), i)
	}
	tree.Scan(
		func(key string, val int) bool {
			fmt.Println(key, val)
			return true
		},
	)
	tree = nil
}

func TestRbTree_Iter(t *testing.T) {
	tree := NewTree[string, int]()
	for i := 1; i < 32; i++ {
		tree.Add(fmt.Sprintf("entry-%.3d", i), i)
	}
	var empty int
	it := tree.Iter()
	for v := it.First(); v != empty; v = it.Next() {
		fmt.Println(v)
	}
	tree = nil
}

func BenchmarkRbTree_Scan(b *testing.B) {
	tree := NewTree[string, int]()
	for i := 1; i < 250; i++ {
		tree.Add(fmt.Sprintf("entry-%.3d", i), i)
	}
	b.ReportAllocs()
	var empty int
	for i := 0; i < b.N; i++ {
		tree.Scan(
			func(key string, val int) bool {
				if val == empty {
					b.Error("got a nil entry")
				}
				return val != empty
			},
		)
	}
	tree = nil
}

func BenchmarkRbTree_Iter(b *testing.B) {
	tree := NewTree[string, int]()
	for i := 1; i < 250; i++ {
		tree.Add(fmt.Sprintf("entry-%.3d", i), i)
	}
	it := tree.Iter()
	b.ReportAllocs()
	var empty int
	for i := 0; i < b.N; i++ {
		for v := it.First(); v != empty; v = it.Next() {
			if v == empty {
				b.Error("go a nil entry")
			}
		}
	}
	tree = nil
}
