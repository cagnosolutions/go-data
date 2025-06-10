package optimizing

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"
	"time"
)

//func convert(v int) [8]byte {
//	return [8]byte{
//		byte(v),
//		byte(v >> 8),
//		byte(v >> 16),
//		byte(v >> 24),
//		byte(v >> 32),
//		byte(v >> 40),
//		byte(v >> 48),
//		byte(v >> 56),
//	}
//}

//func BenchmarkBPTree_Byts_Add(b *testing.B) {
//
//	tree, err := NewBPTree()
//	if err != nil {
//		b.Errorf("Error creating BPT: %v", err)
//	}
//
//	//data := byts
//
//	b.ReportAllocs()
//	for i := 0; i < b.N; i++ {
//		// add data
//		for j := 0; j < 500; j++ {
//			tree.Add(keyType{convert(j)}, valType{nums[j]})
//		}
//		// find data
//		for j := 0; j < 500; j++ {
//			_, val := tree.Get(keyType{convert(j)})
//			if val.data != uint32(j) {
//				b.Errorf("expected %d, got %d", j, val)
//			}
//		}
//	}
//
//	// print tree data size in bytes
//	//fmt.Println(tree.Size())
//
//	//print_tree_v2(tree.root)
//
//	tree.Close()
//}

//func BenchmarkBPTree_Byts2_Add(b *testing.B) {
//
//	tree, err := NewBPTree()
//	if err != nil {
//		b.Errorf("Error creating BPT: %v", err)
//	}
//
//	data := byts
//
//	b.ReportAllocs()
//	for i := 0; i < b.N; i++ {
//		// add data
//		for j := 0; j < len(data); j++ {
//			tree.Add(keyType{data[j]}, valType{nums[j]})
//		}
//		// find data
//		for j := 0; j < len(data); j++ {
//			_, val := tree.Get(keyType{data[j]})
//			if val.data != uint32(j) {
//				b.Errorf("expected %d, got %d", j, val)
//			}
//		}
//	}
//
//	// print tree data size in bytes
//	//fmt.Println(tree.Size())
//
//	//print_tree_v2(tree.root)
//
//	tree.Close()
//}

//func BenchmarkBPTree_Ints_Add(b *testing.B) {
//
//	tree, err := NewBPTree()
//	if err != nil {
//		b.Errorf("Error creating BPT: %v", err)
//	}
//
//	data := nums
//
//	b.ReportAllocs()
//	for i := 0; i < b.N; i++ {
//		// add data
//		for j := 0; j < len(data); j++ {
//			tree.Add(keyType{data[j]}, valType{data[j]})
//		}
//		// find data
//		for j := 0; j < len(data); j++ {
//			_, val := tree.Get(keyType{uint32(j)})
//			if val.data != uint32(j) {
//				b.Errorf("expected %d, got %d", 123, val)
//			}
//		}
//	}
//
//	// print tree data size in bytes
//	//fmt.Println(tree.Size())
//
//	//print_tree_v2(tree.root)
//
//	tree.Close()
//}

//func BenchmarkBPTree_Ints2_Add(b *testing.B) {
//
//	tree, err := NewBPTree()
//	if err != nil {
//		b.Errorf("Error creating BPT: %v", err)
//	}
//
//	data := nums2
//
//	b.ReportAllocs()
//	for i := 0; i < b.N; i++ {
//		// add data
//		for j := 0; j < len(data); j++ {
//			tree.Add(keyType{data[j]}, valType{nums[j]})
//		}
//		// find data
//		for j := 0; j < len(data); j++ {
//			_, val := tree.Get(keyType{uint64(j)})
//			if val.data != uint32(j) {
//				b.Errorf("expected %d, got %d", 123, val)
//			}
//		}
//	}
//
//	// print tree data size in bytes
//	//fmt.Println(tree.Size())
//
//	//print_tree_v2(tree.root)
//
//	tree.Close()
//}

//func BenchmarkBPTree_Strs_Add(b *testing.B) {
//
//	tree, err := NewBPTree()
//	if err != nil {
//		b.Errorf("Error creating BPT: %v", err)
//	}
//
//	data := strs
//
//	b.ReportAllocs()
//	for i := 0; i < b.N; i++ {
//		// add data
//		for j := 0; j < len(data); j++ {
//			tree.Add(keyType{data[j]}, valType{nums[j]})
//		}
//		// find data
//		for j := 0; j < len(data); j++ {
//			_, val := tree.Get(keyType{data[j]})
//			if val.data != uint32(j) {
//				b.Errorf("expected %d, got %d", nums[j], val)
//			}
//		}
//	}
//
//	// print tree data size in bytes
//	//fmt.Println(tree.Size())
//
//	tree.Close()
//}

func PrintMemStats(mem runtime.MemStats) {
	fmt.Printf("  Alloc: %v bytes (%d KB) (%d MB)\n", mem.Alloc, mem.Alloc/1024, mem.Alloc/1024/1024)
	fmt.Printf("  TotalAlloc: %v bytes (%d KB) (%d MB)\n", mem.TotalAlloc, mem.TotalAlloc/1024, mem.TotalAlloc/1024/1024)
	fmt.Printf("  Sys: %v bytes (%d KB) (%d MB)\n", mem.Sys, mem.Sys/1024, mem.Sys/1024/1024)
	fmt.Printf("  NumGC: %v\n", mem.NumGC)
}

//func TestBTreeLoad_Ints(t *testing.T) {
//
//	tree, err := NewBPTree()
//	if err != nil {
//		t.Errorf("Error creating BPT: %v", err)
//	}
//
//	count := 10000000
//
//	var mem runtime.MemStats
//
//	// add data
//	start := time.Now()
//	for j := 0; j < count; j++ {
//		tree.Add(keyType{uint64(j)}, valType{uint32(j)})
//	}
//	fmt.Printf("Time to add %d elements: %v\n", count, time.Since(start))
//
//	time.Sleep(5 * time.Second)
//
//	runtime.ReadMemStats(&mem)
//	fmt.Printf("Added %d elements to tree\n", count)
//	PrintMemStats(mem)
//
//	// find data
//	for j := 0; j < count; j++ {
//		_, val := tree.Get(keyType{uint64(j)})
//		if val.data != uint32(j) {
//			t.Errorf("expected %d, got %d", uint32(j), val)
//		}
//	}
//
//	time.Sleep(10 * time.Second)
//
//	tree.Close()
//}

func TestBTreeLoad_Strs(t *testing.T) {

	tree, err := NewBPTree()
	if err != nil {
		t.Errorf("Error creating BPT: %v", err)
	}

	count := 100000

	var mem runtime.MemStats

	// add data
	start := time.Now()
	for j := 0; j < count; j++ {
		tree.Add(keyType{MakeStrKey(strconv.Itoa(j))}, valType{uint32(j)})
	}
	fmt.Printf("Time to add %d elements: %v\n", count, time.Since(start))

	time.Sleep(5 * time.Second)

	runtime.ReadMemStats(&mem)
	fmt.Printf("Added %d elements to tree\n", count)
	PrintMemStats(mem)

	// find data
	for j := 0; j < count; j++ {
		_, val := tree.Get(keyType{MakeStrKey(strconv.Itoa(j))})
		if val.data != uint32(j) {
			t.Errorf("expected %d, got %d", uint32(j), val)
		}
	}

	time.Sleep(10 * time.Second)

	tree.Close()
}

func TestMapLoad_Strs(t *testing.T) {

	type key struct{ data string }
	type val struct{ data uint32 }

	mdata := make(map[key]val)

	count := 10_000_000

	var mem runtime.MemStats

	// add data
	start := time.Now()
	for j := 0; j < count; j++ {
		mdata[key{strconv.Itoa(j)}] = val{uint32(j)}
	}
	fmt.Printf("Time to add %d elements: %v\n", count, time.Since(start))

	time.Sleep(5 * time.Second)

	runtime.ReadMemStats(&mem)
	fmt.Printf("Added %d elements to tree\n", count)
	PrintMemStats(mem)

	// find data
	for j := 0; j < count; j++ {
		val, _ := mdata[key{strconv.Itoa(j)}]
		if val.data != uint32(j) {
			t.Errorf("expected %d, got %d", uint32(j), val)
		}
	}

	time.Sleep(10 * time.Second)
	mdata = nil
}
