package rbt

import (
	"fmt"
	"strings"
	"testing"
)

type entry struct {
	data string
}

func (e entry) Compare(that RBEntry) int {
	return strings.Compare(e.data, that.(entry).data)
}

func (e entry) Size() int {
	return len(e.data)
}

func (e entry) String() string {
	return fmt.Sprintf("entry.data=%q", e.data)
}

func TestRbTree_Scan(t *testing.T) {
	tree := newRBTree()
	for i := 0; i < 32; i++ {
		tree.Add(entry{fmt.Sprintf("entry-%.3d", i)})
	}
	tree.Scan(
		func(e RBEntry) bool {
			fmt.Println(e)
			return true
		},
	)
	tree = nil
}

func TestRbTree_Iter(t *testing.T) {
	tree := newRBTree()
	for i := 0; i < 32; i++ {
		tree.Add(entry{fmt.Sprintf("entry-%.3d", i)})
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
		tree.Add(entry{fmt.Sprintf("entry-%.3d", i)})
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tree.Scan(
			func(e RBEntry) bool {
				if e == nil {
					b.Error("got a nil entry")
				}
				return e != nil
			},
		)
	}
	tree = nil
}

func BenchmarkRbTree_Iter(b *testing.B) {
	tree := newRBTree()
	for i := 0; i < 250; i++ {
		tree.Add(entry{fmt.Sprintf("entry-%.3d", i)})
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
