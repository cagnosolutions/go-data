package bpt

import (
	"log"
	"unsafe"
)

// delete functions as the master delete method. it returns the previous record upon success
func (t *BPTree) delete(k keyType) *record {
	var old *record
	keyLeaf, keyEntry := t.find(k)
	if keyEntry != nil && keyLeaf != nil {
		t.root = deleteEntry(t.root, keyLeaf, k, unsafe.Pointer(keyEntry))
		old = keyEntry
		keyEntry = nil
	}
	return old // return the old record we just deleted
}

// deleteEntry removes the record and its key and pointer from the leaf, and then makes all
// appropriate changes to preserve the tree's properties
func deleteEntry(root, n *node, k keyType, pointer unsafe.Pointer) *node {

	// initialize temporary variables
	var minKeys, kPrimeIndex, capacity int

	// remove the key and value from the current node
	n = removeEntryFromNode(n, k, pointer)

	// if the node is the room node, make sure to adjust
	if n == root {
		return adjustRoot(root)
	}

	// otherwise, we are deleting with an internal or leaf node, so we must determine
	// the minimum allowable size of the node to be preserved after deletion to remain
	// true to the tree's properties. a leaf and an internal node will have different cut points
	if n.isLeaf {
		minKeys = cut(order - 1)
	} else {
		minKeys = cut(order) - 1
	}

	// and if the node is above (or at) the minimum order simply return (the deletion is done)
	if n.numKeys >= minKeys {
		return root
	}

	// otherwise, the node falls below the minimum order, so we must determine if we
	// must coalesce or redistribute the nodes. first we will find the appropriate
	// neighbor node with which to coalesce along with the key (kPrime) in the parent
	// between the pointer to node n and the pointer to the neighboring node
	neighborIndex := getNeighborIndex(n)
	if neighborIndex == -1 {
		kPrimeIndex = 0
	} else {
		kPrimeIndex = neighborIndex
	}

	kPrime := n.parent.keys[kPrimeIndex]

	var neighbor *node
	if neighborIndex == -1 {
		neighbor = (*node)(n.parent.ptrs[1])
	} else {
		neighbor = (*node)(n.parent.ptrs[neighborIndex])
	}

	if n.isLeaf {
		capacity = order
	} else {
		capacity = order - 1
	}

	// coalesce (underflow) the nodes
	if neighbor.numKeys+n.numKeys < capacity {
		return coalesceNodes(root, n, neighbor, neighborIndex, kPrime)
	}

	// redistribute the nodes
	return redistributeNodes(root, n, neighbor, neighborIndex, kPrimeIndex, kPrime)
}

// removeEntryFromNode does just that
func removeEntryFromNode(n *node, k keyType, pointer unsafe.Pointer) *node {

	// remove the key and shift the other keys accordingly
	var i, numPointers int
	for n.keys[i].data != k.data {
		i++
	}
	for i++; i < n.numKeys; i++ { // was for i+=1;
		n.keys[i-1] = n.keys[i]
	}

	// then, remove the pointer and shift the other pointers accordingly
	if n.isLeaf {
		numPointers = n.numKeys
	} else {
		numPointers = n.numKeys + 1
	}

	i = 0
	for n.ptrs[i] != pointer {
		i++
	}
	for i++; i < numPointers; i++ { // was for i+=1;
		n.ptrs[i-1] = n.ptrs[i]
	}

	// make sure we decrement, because now we are one key fewer
	n.numKeys--

	// set the other pointers to nil for tidiness and remember that a leaf uses
	// the last pointer to point to the next leaf
	if n.isLeaf {
		for i = n.numKeys; i < order-1; i++ {
			n.ptrs[i] = nil
		}
	} else {
		for i = n.numKeys + 1; i < order; i++ {
			n.ptrs[i] = nil
		}
	}
	return n
}

// adjustRoot does some magic in the root node (not really)
func adjustRoot(root *node) *node {

	// in the case of a non-empty root, the key and the pointer for the
	// entry have already been removed so there is nothing else to do
	if root.numKeys > 0 {
		return root
	}

	// otherwise, the root node is empty, so it must have at least one child. we must
	// promote the first child as the new root node (the tree must always have a root)
	var newRoot *node
	if !root.isLeaf {
		newRoot = (*node)(root.ptrs[0])
		newRoot.parent = nil
	} else {
		// and if it is a leaf node (has no children) then the whole tree is in fact empty
		newRoot = nil // free
	}
	root = nil
	return newRoot
}

// getNeighborIndex is a utility function for deletion. it gets the index of
// a node's nearest sibling (that exists) to the left and if it cannot find one
// then the node is already the leftmost child and (in such a case the node)
// will return -1
func getNeighborIndex(n *node) int {
	var i int
	for i = 0; i <= n.parent.numKeys; i++ {
		if (*node)(n.parent.ptrs[i]) == n {
			return i - 1
		}
	}
	log.Panicf("getNeighborIndex: Search for nonexistent pointer to node in parent.\nNode: %#v\n", n)
	return i
}

// coalesceNodes coalesces a node (that has become too small after deletion) along with
// a neighboring node that has room to accept the additional entries without exceeding
// the maximum order of the tree
func coalesceNodes(root, n, neighbor *node, neighborIndex int, kPrime keyType) *node {

	// initialize temp variables
	var tmp *node

	// swap neighbor with node if node is on the extreme left and neighbor is to its right
	if neighborIndex == -1 {
		tmp = n
		n = neighbor
		neighbor = tmp
	}

	// starting point in the neighbor for copying keys and pointers from node n. recall
	// that n and neighbor have swapped places the in special case of n being a leftmost child
	neighborInsertionIndex := neighbor.numKeys
	var i, j, nEnd int

	// and if the node is an internal (non leaf) node, we append the kPrime key and the
	// following keys and pointers from the neighbor
	if !n.isLeaf {
		// append kPrime
		neighbor.keys[neighborInsertionIndex] = kPrime
		neighbor.numKeys++
		nEnd = n.numKeys

		for i, j = neighborInsertionIndex+1, 0; j < nEnd; i, j = i+1, j+1 {
			neighbor.keys[i] = n.keys[j]
			neighbor.ptrs[i] = n.ptrs[j]
			neighbor.numKeys++
			n.numKeys--
		}

		// the number of pointers is always one more than the number of keys
		neighbor.ptrs[i] = n.ptrs[j]

		// all children must now point up to the same parent
		for i = 0; i < neighbor.numKeys+1; i++ {
			tmp = (*node)(neighbor.ptrs[i])
			tmp.parent = neighbor
		}
	} else {
		// otherwise, the node is a leaf node, append the keys and pointers of n to
		// the neighbor and because it's a leaf node, we must set the neighbor's last
		// pointer to point to what ha been n's rightmost neighbor
		for i, j = neighborInsertionIndex, 0; j < n.numKeys; i, j = i+1, j+1 {
			neighbor.keys[i] = n.keys[j]
			neighbor.ptrs[i] = n.ptrs[j]
			neighbor.numKeys++
		}
		neighbor.ptrs[order-1] = n.ptrs[order-1]
	}
	root = deleteEntry(root, n.parent, kPrime, unsafe.Pointer(n))
	n = nil // free
	return root
}

// redistributeNodes redistributes entries between two nodes when one has become too
// small after deletion but its neighbor is too big to append the small node's entries
// without exceeding the maximum
func redistributeNodes(root, n, neighbor *node, neighborIndex, kPrimeIndex int, kPrime keyType) *node {

	// initialize temporary variables
	var i int
	var tmp *node

	// in the case where n has a neighbor to the left, pull the neighbor's last
	// key-pointer pair over from the neighbor's right end to n's left end
	if neighborIndex != -1 {
		if !n.isLeaf {
			n.ptrs[n.numKeys+1] = n.ptrs[n.numKeys]
		}
		for i = n.numKeys; i > 0; i-- {
			n.keys[i] = n.keys[i-1]
			n.ptrs[i] = n.ptrs[i-1]
		}
		if !n.isLeaf {
			n.ptrs[0] = neighbor.ptrs[neighbor.numKeys]
			tmp = (*node)(n.ptrs[0])
			tmp.parent = n
			neighbor.ptrs[neighbor.numKeys] = nil
			n.keys[0] = kPrime
			n.parent.keys[kPrimeIndex] = neighbor.keys[neighbor.numKeys-1]
		} else {
			n.ptrs[0] = neighbor.ptrs[neighbor.numKeys-1]
			neighbor.ptrs[neighbor.numKeys-1] = nil
			n.keys[0] = neighbor.keys[neighbor.numKeys-1]
			n.parent.keys[kPrimeIndex] = n.keys[0]
		}
	} else {
		// in the case where n is the leftmost child, take a key-pointer pair from
		// the neighbor to the right, then move the neighbor's leftmost key-pointer
		// pair to n's rightmost position
		if n.isLeaf {
			n.keys[n.numKeys] = neighbor.keys[0]
			n.ptrs[n.numKeys] = neighbor.ptrs[0]
			n.parent.keys[kPrimeIndex] = neighbor.keys[1]
		} else {
			n.keys[n.numKeys] = kPrime
			n.ptrs[n.numKeys+1] = neighbor.ptrs[0]
			tmp = (*node)(n.ptrs[n.numKeys+1])
			tmp.parent = n
			n.parent.keys[kPrimeIndex] = neighbor.keys[0]
		}
		for i = 0; i < neighbor.numKeys-1; i++ {
			neighbor.keys[i] = neighbor.keys[i+1]
			neighbor.ptrs[i] = neighbor.ptrs[i+1]
		}
		if !n.isLeaf {
			neighbor.ptrs[i] = neighbor.ptrs[i+1]
		}
	}

	// now, n has one more key and one more pointer and the neighbor has one fewer
	// of each, so don't forget to properly increment and decrement each accordingly
	n.numKeys++
	neighbor.numKeys--
	return root
}
