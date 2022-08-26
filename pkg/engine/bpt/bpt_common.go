package bpt

import (
	"fmt"
	"unsafe"
)

type keyType struct {
	data uint32
}

type valType struct {
	data []byte // implement
}

// record represents a record pointed to by a leaf node
type record struct {
	Key   keyType
	Value valType
}

func (r *record) Size() int64 {
	return int64(unsafe.Sizeof(r.Key.data) + unsafe.Sizeof(r.Value.data))
}

const M = 5 // 128

// order is the tree's order
const order = M // 128

// node represents a node of the BPTree
type node struct {
	numKeys int
	keys    [order - 1]keyType
	ptrs    [order]unsafe.Pointer
	parent  *node
	isLeaf  bool
}

// String is node's stringer method
func (n *node) String() string {
	ss := fmt.Sprintf("\tr%dn%d[", height(n), pathToRoot(n.parent, n))
	for i := 0; i < n.numKeys-1; i++ {
		ss += fmt.Sprintf("%.2d", n.keys[i].data)
		ss += fmt.Sprintf(",")
	}
	ss += fmt.Sprintf("%.2d]", n.keys[n.numKeys-1].data)
	return ss
}

// BPTree represents the root of a b+tree
// the only thing needed to start a new tree
// is to simply call bpt := new(BPTree)
type BPTree struct {
	root *node
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
func (n *node) nextLeaf() *node {
	if p := (*node)(n.ptrs[order-1]); p != nil && p.isLeaf {
		return p
	}
	return nil
}

// destroyTree is a helper for "destroying" the tree
func (t *BPTree) destroyTree() {
	destroyTreeNodes(t.root)
}

// destroyTreeNodes is called recursively by destroyTree
func destroyTreeNodes(n *node) {
	if n == nil {
		return
	}
	if n.isLeaf {
		for i := 0; i < n.numKeys; i++ {
			n.ptrs[i] = nil
		}
	} else {
		for i := 0; i < n.numKeys+1; i++ {
			destroyTreeNodes((*node)(n.ptrs[i]))
		}
	}
	n = nil
}

// Size attempts to return the tree size in bytes
func (t *BPTree) Size() int64 {
	c := findFirstLeaf(t.root)
	if c == nil {
		return 0
	}
	var s int64
	var r *record
	for {
		for i := 0; i < c.numKeys; i++ {
			r = (*record)(c.ptrs[i])
			if r != nil {
				s += int64(r.Size())
			}
		}
		if c.ptrs[order-1] != nil {
			c = (*node)(c.ptrs[order-1])
		} else {
			break
		}
	}
	return s
}

// hasKey reports whether this leaf node contains the provided key
func (n *node) hasKey(k keyType) bool {
	if n.isLeaf {
		for i := 0; i < n.numKeys; i++ {
			if k.data == n.keys[i].data {
				return true
			}
		}
	}
	return false
}

// closest returns the closest matching record for the provided key
func (n *node) closest(k keyType) (*record, bool) {
	if n.isLeaf {
		i := 0
		for ; i < n.numKeys; i++ {
			if k.data < n.keys[i].data {
				break
			}
		}
		if i > 0 {
			i--
		}
		return (*record)(n.ptrs[i]), true
	}
	return nil, false
}

// record returns the matching record for the provided key
func (n *node) record(k keyType) (*record, bool) {
	if n.isLeaf {
		for i := 0; i < n.numKeys; i++ {
			if k.data == n.keys[i].data {
				return (*record)(n.ptrs[i]), true
			}
		}
	}
	return nil, false
}
