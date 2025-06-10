package optimizing

import (
	"fmt"
	"testing"
)

func TestData(t *testing.T) {
	PrintInts(nums)
	PrintStrs(strs)
}

func TestBPTree_Add(t *testing.T) {
	tree, err := NewBPTree()
	if err != nil {
		t.Errorf("Error creating BPT: %v", err)
	}

	// add data
	tree.Add(keyType{2}, valType{2})
	tree.Add(keyType{1}, valType{1})
	tree.Add(keyType{6}, valType{6})
	tree.Add(keyType{3}, valType{3})
	tree.Add(keyType{4}, valType{4})
	tree.Add(keyType{5}, valType{5})
	tree.Add(keyType{7}, valType{7})
	tree.Add(keyType{13}, valType{13})
	tree.Add(keyType{8}, valType{8})
	tree.Add(keyType{0}, valType{0})
	tree.Add(keyType{11}, valType{11})
	tree.Add(keyType{9}, valType{9})
	tree.Add(keyType{12}, valType{12})

	// print data
	minK, minV := tree.Min()
	fmt.Printf("min: %v %v\n", minK, minV)
	maxK, maxV := tree.Max()
	fmt.Printf("max: %v %v\n", maxK, maxV)

	// range
	tree.Range(func(k keyType, v valType) bool {
		fmt.Printf("%v %v\n", k, v)
		return true
	})

	// close tree
	tree.Close()
}
