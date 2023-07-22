package generic

import (
	"fmt"
	"unsafe"
)

type Key interface {
	Compare(that Key) int
}

// record represents a record pointed to by a leaf node
type record[K Key, V any] struct {
	Key   K
	Value V
}

func (r *record[K, V]) Size() int64 {
	return int64(unsafe.Sizeof(r.Key) + unsafe.Sizeof(r.Value))
}

const M = 5 // 128

// order is the tree's order
const order = M // 128

// node represents a node of the BPTree
type node[K Key, V any] struct {
	numKeys int
	keys    [order - 1]K
	ptrs    [order]unsafe.Pointer
	parent  *node[K, V]
	isLeaf  bool
}

func (n *node[K, V]) equals(that unsafe.Pointer) bool {
	return unsafe.Pointer(n) == that
}

// String is node's stringer method
func (n *node[K, V]) String() string {
	ss := fmt.Sprintf("\tr%dn%d[", height(n), pathToRoot(n.parent, n))
	for i := 0; i < n.numKeys-1; i++ {
		ss += fmt.Sprintf("%v", n.keys[i])
		ss += fmt.Sprintf(",")
	}
	ss += fmt.Sprintf("%v]", n.keys[n.numKeys-1])
	return ss
}

// BPTree represents the root of a b+tree
// the only thing needed to start a new tree
// is to simply call bpt := new(BPTree)
type BPTree[K Key, V any] struct {
	root    *node[K, V]
	isEmpty bool
}

// cut finds the appropriate place to split a node that is
// too big. it is used both during insertion and deletion
func cut(length int) int {
	if length%2 == 0 {
		return length / 2
	}
	return length/2 + 1
}

// nextLeaf returns the next non-nil leaf in the chain (to the right) of the current leaf
func (n *node[K, V]) nextLeaf() *node[K, V] {
	if p := (*node[K, V])(n.ptrs[order-1]); p != nil && p.isLeaf {
		return p
	}
	return nil
}

// destroyTree is a helper for "destroying" the tree
func (t *BPTree[K, V]) destroyTree() {
	destroyTreeNodes[K, V](t.root)
}

// destroyTreeNodes is called recursively by destroyTree
func destroyTreeNodes[K Key, V any](n *node[K, V]) {
	if n == nil {
		return
	}
	if n.isLeaf {
		for i := 0; i < n.numKeys; i++ {
			n.ptrs[i] = nil
		}
	} else {
		for i := 0; i < n.numKeys+1; i++ {
			destroyTreeNodes((*node[K, V])(n.ptrs[i]))
		}
	}
	n = nil
}

// Size attempts to return the tree size in bytes
func (t *BPTree[K, V]) Size() int64 {
	c := findFirstLeaf(t.root)
	if c == nil {
		return 0
	}
	var s int64
	var r *record[K, V]
	for {
		for i := 0; i < c.numKeys; i++ {
			r = (*record[K, V])(c.ptrs[i])
			if r != nil {
				s += int64(r.Size())
			}
		}
		if c.ptrs[order-1] != nil {
			c = (*node[K, V])(c.ptrs[order-1])
		} else {
			break
		}
	}
	return s
}

// hasKey reports whether this leaf node contains the provided Key
func (n *node[K, V]) hasKey(k K) bool {
	if n.isLeaf {
		for i := 0; i < n.numKeys; i++ {
			if k.Compare(n.keys[i]) == 0 {
				return true
			}
		}
	}
	return false
}

// closest returns the closest matching record for the provided Key
func (n *node[K, V]) closest(k K) (*record[K, V], bool) {
	if n.isLeaf {
		i := 0
		for ; i < n.numKeys; i++ {
			if k.Compare(n.keys[i]) == -1 {
				break
			}
		}
		if i > 0 {
			i--
		}
		return (*record[K, V])(n.ptrs[i]), true
	}
	return nil, false
}

// record returns the matching record for the provided Key
func (n *node[K, V]) record(k K) (*record[K, V], bool) {
	if n.isLeaf {
		for i := 0; i < n.numKeys; i++ {
			if k.Compare(n.keys[i]) == 0 {
				return (*record[K, V])(n.ptrs[i]), true
			}
		}
	}
	return nil, false
}
