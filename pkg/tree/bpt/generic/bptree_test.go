package generic

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

const (
	thousand = 1000
	n        = 1
)

type KEY uint32

func (k KEY) Compare(that Key) int {
	if k < that.(KEY) {
		return -1
	}
	if k > that.(KEY) {
		return +1
	}
	return 0
}

func TestBPTree_Print(t *testing.T) {

	tree := new(BPTree[KEY, string])

	for i := 0; i < 32; i++ {
		existing := tree.Put(makeKey(i), makeVal(i))
		if existing { // existing=updated
			t.Errorf("putting: %v", existing)
		}
	}

	print_tree[KEY, string](tree.root)
	// print_leaves(tree.root)

	tree.Close()
}

func TestBPTree_PrintV2(t *testing.T) {

	tree := new(BPTree[KEY, string])

	for i := 0; i < 64; i++ {
		existing := tree.Put(makeKey(i), makeVal(i))
		if existing { // existing=updated
			t.Errorf("putting: %v", existing)
		}
	}

	printTree(tree.root)

	print_tree_v2(tree.root)
	// print_leaves(tree.root)

	tree.Close()
}

func TestBPTree_PrintMarkdownTree(t *testing.T) {

	tree := new(BPTree[KEY, string])

	for i := 0; i < 32; i++ {
		existing := tree.Put(makeKey(i), makeVal(i))
		if existing { // existing=updated
			t.Errorf("putting: %v", existing)
		}
	}

	print_markdown_tree(tree.root)
	// print_leaves(tree.root)

	tree.Close()
}

func TestNewBPTree(t *testing.T) {
	var tree *BPTree[KEY, string]
	tree = new(BPTree[KEY, string])
	AssertNotNil(t, tree)
	tree.Close()
}

func TestDelFromNewBPTree(t *testing.T) {
	var tree *BPTree[KEY, string]
	tree = new(BPTree[KEY, string])
	AssertNotNil(t, tree)
	tree.Del(KEY(4))
	tree.Close()
}

func TestBPTree_Has(t *testing.T) {
	tree := new(BPTree[KEY, string])
	AssertLen(t, 0, tree.Len())
	for i := 0; i < n*thousand; i++ {
		tree.Put(makeKey(i), makeVal(i))
	}
	for i := 0; i < n*thousand; i++ {
		ok := tree.Has(makeKey(i))
		if !ok { // existing=updated
			t.Errorf("has: %v", ok)
		}
	}
	AssertLen(t, n*thousand, tree.Len())
	tree.Close()
}

func TestBPTree_Put(t *testing.T) {
	tree := new(BPTree[KEY, string])
	AssertLen(t, 0, tree.Len())
	for i := 0; i < n*thousand; i++ {
		existing := tree.Put(makeKey(i), makeVal(i))
		if existing { // existing=updated
			t.Errorf("putting: %v", existing)
		}
	}
	AssertLen(t, n*thousand, tree.Len())
	tree.Close()
}

func TestBPTree_Get(t *testing.T) {
	tree := new(BPTree[KEY, string])
	for i := 0; i < n*thousand; i++ {
		tree.Put(makeKey(i), makeVal(i))
	}
	AssertLen(t, n*thousand, tree.Len())
	for i := 0; i < n*thousand; i++ {
		_, v := tree.Get(makeKey(i))
		if v == "" {
			t.Errorf("getting: %v", v)
		}
		AssertEqual(t, makeVal(i), v)
	}
	tree.Close()
}

func TestBPTree_Del(t *testing.T) {
	tree := new(BPTree[KEY, string])
	for i := 0; i < n*thousand; i++ {
		tree.Put(makeKey(i), makeVal(i))
	}
	AssertLen(t, n*thousand, tree.Len())
	for i := 0; i < n*thousand; i++ {
		_, v := tree.Del(makeKey(i))
		if v == "" {
			t.Errorf("delete: %v", v)
		}
	}
	AssertLen(t, 0, tree.Len())
	tree.Close()
}

func TestBPTree_Len(t *testing.T) {
	tree := new(BPTree[KEY, string])
	for i := 0; i < n*thousand; i++ {
		tree.Put(makeKey(i), makeVal(i))
	}
	AssertLen(t, n*thousand, tree.Len())
	tree.Close()
}

func TestBPTree_Min(t *testing.T) {
	tree := new(BPTree[KEY, string])
	for i := 0; i < n*thousand; i++ {
		tree.Put(makeKey(i), makeVal(i))
	}
	AssertLen(t, n*thousand, tree.Len())
	k, v := tree.Min()
	if v == "" {
		t.Errorf("min: %v", tree)
	}
	AssertEqual(t, makeKey(0), k)
	tree.Close()
}

func TestBPTree_Max(t *testing.T) {
	tree := new(BPTree[KEY, string])
	for i := 0; i < n*thousand; i++ {
		tree.Put(makeKey(i), makeVal(i))
	}
	AssertLen(t, n*thousand, tree.Len())
	k, v := tree.Max()
	if v == "" {
		t.Errorf("min: %v", tree)
	}
	AssertEqual(t, makeKey(n*thousand-1), k)
	tree.Close()
}

func TestBPTree_Range(t *testing.T) {
	tree := new(BPTree[KEY, string])
	for i := 0; i < n*thousand; i++ {
		tree.Put(makeKey(i), makeVal(i))
	}
	AssertLen(t, n*thousand, tree.Len())

	printInfo := true

	// do scan front
	tree.Range(
		func(k KEY, v string) bool {
			if k < 0 {
				t.Errorf("scan front, issue with Key: %v", k)
				return false
			}
			if printInfo {
				log.Printf("Key: %v\n", k)
			}
			return true
		},
	)

	tree.Close()
}

func TestBPTree_Close(t *testing.T) {
	var tree *BPTree[KEY, string]
	tree = new(BPTree[KEY, string])
	tree.Close()
}

func makeKey(i int) KEY {
	return KEY(i)
}

func makeVal(i int) string {
	return fmt.Sprintf("{\"id\":%.6d,\"Key\":\"Key-%.6d\",\"value\":\"val-%.6d\"}", i, i, i)
}

func AssertExpected(t *testing.T, expected, got interface{}) bool {
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("error, expected: %v, got: %v\n", expected, got)
		return false
	}
	return true
}

func AssertLen(t *testing.T, expected, got interface{}) bool {
	return AssertExpected(t, expected, got)
}

func AssertEqual(t *testing.T, expected, got interface{}) bool {
	return AssertExpected(t, expected, got)
}

func AssertTrue(t *testing.T, got interface{}) bool {
	return AssertExpected(t, true, got)
}

func AssertError(t *testing.T, got interface{}) bool {
	return AssertExpected(t, got, got)
}

func AssertNoError(t *testing.T, got interface{}) bool {
	return AssertExpected(t, nil, got)
}

func AssertNil(t *testing.T, got interface{}) bool {
	return AssertExpected(t, nil, got)
}

func AssertNotNil(t *testing.T, got interface{}) bool {
	return got != nil
}
