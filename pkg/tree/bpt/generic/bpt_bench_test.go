package generic

import (
	"testing"
)

func BenchmarkBPTree_Ints_Add(b *testing.B) {

	comp := func(a, b uint32) int {
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	}
	tree, err := NewTree[uint32, uint32](comp)
	if err != nil {
		b.Errorf("Error creating BPT: %v", err)
	}

	data := nums

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// add data
		for j := 0; j < len(data); j++ {
			tree.Add(data[j], data[j])
		}
		// find data
		for j := 0; j < len(data); j++ {
			_, val := tree.Get(data[j])
			if val != data[j] {
				b.Errorf("expected %d, got %d", data[j], val)
			}
		}
	}

	// print tree data size in bytes
	//fmt.Println(tree.Size())

	tree.Close()
}
