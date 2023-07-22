package generic

import (
	"fmt"
	"strings"
	"testing"
)

type entry struct {
	data string
}

func (e entry) Compare(that Entry) int {
	return strings.Compare(e.data, that.(entry).data)
}

func (e entry) String() string {
	return fmt.Sprintf("%#v", e)
}

func TestRbTree_Scan(t *testing.T) {
	tree := NewTree[entry]()
	for i := 0; i < 32; i++ {
		tree.Add(entry{fmt.Sprintf("entry-%.3d", i)})
	}
	tree.Scan(
		func(e entry) bool {
			fmt.Println(e)
			return true
		},
	)
	tree = nil
}

func TestRbTree_Iter(t *testing.T) {
	tree := NewTree[entry]()
	for i := 0; i < 32; i++ {
		tree.Add(entry{fmt.Sprintf("entry-%.3d", i)})
	}
	var empty entry
	it := tree.Iter()
	for e := it.First(); e != empty; e = it.Next() {
		fmt.Println(e)
	}
	tree = nil
}

func BenchmarkRbTree_Scan(b *testing.B) {
	tree := NewTree[entry]()
	for i := 0; i < 250; i++ {
		tree.Add(entry{fmt.Sprintf("entry-%.3d", i)})
	}
	b.ReportAllocs()
	var empty entry
	for i := 0; i < b.N; i++ {
		tree.Scan(
			func(e entry) bool {
				if e == empty {
					b.Error("got a nil entry")
				}
				return e != empty
			},
		)
	}
	tree = nil
}

func BenchmarkRbTree_Iter(b *testing.B) {
	tree := NewTree[entry]()
	for i := 0; i < 250; i++ {
		tree.Add(entry{fmt.Sprintf("entry-%.3d", i)})
	}
	it := tree.Iter()
	b.ReportAllocs()
	var empty entry
	for i := 0; i < b.N; i++ {
		for e := it.First(); e != empty; e = it.Next() {
			if e == empty {
				b.Error("go a nil entry")
			}
		}
	}
	tree = nil
}
