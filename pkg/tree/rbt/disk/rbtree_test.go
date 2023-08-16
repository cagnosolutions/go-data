package disk

import (
	"fmt"
	"testing"
)

func TestRbTree_Scan(t *testing.T) {
	tree := newRBTree()
	for i := 0; i < 32; i++ {
		tree.Add(fmt.Sprintf("entry-%.3d", i), i)
	}
	fmt.Println(tree.Len())
	tree.Scan(
		func(key string, val any) bool {
			fmt.Println(key, val)
			return true
		},
	)
	tree = nil
}

func TestRbTree_Iter(t *testing.T) {
	tree := newRBTree()
	for i := 0; i < 32; i++ {
		tree.Add(fmt.Sprintf("entry-%.3d", i), i)
	}
	it := tree.Iter()
	for e := it.First(); e != nil; e = it.Next() {
		fmt.Println(e)
	}
	tree = nil
}

func BenchmarkRbTree_Scan(b *testing.B) {
	tree := newRBTree()
	for i := 0; i < 250; i++ {
		tree.Add(fmt.Sprintf("entry-%.3d", i), i)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tree.Scan(
			func(key string, val any) bool {
				if key == "" {
					b.Error("got a nil entry")
				}
				fmt.Printf("key=%q, val=%v\n", key, val)
				return key != ""
			},
		)
	}
	tree = nil
}

func BenchmarkRbTree_Iter(b *testing.B) {
	tree := newRBTree()
	for i := 0; i < 250; i++ {
		tree.Add(fmt.Sprintf("entry-%.3d", i), i)
	}
	it := tree.Iter()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for e := it.First(); e != nil; e = it.Next() {
			if e == nil {
				b.Error("go a nil entry")
			}
		}
	}
	tree = nil
}
