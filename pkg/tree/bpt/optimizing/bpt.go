package optimizing

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"strings"
	"unsafe"
)

func MakeStrKey(k string) [16]byte {
	var key [16]byte
	copy(key[:], k)
	return key
}

func MakeIntKey(k uint64) [16]byte {
	var key [16]byte
	binary.LittleEndian.PutUint64(key[:8], k)
	return key
}

func compare(k1, k2 keyType) int {
	return bytes.Compare(k1.data[:16], k2.data[:16])
	//return strings.Compare(k1.data, k2.data)
	//if k1.data < k2.data {
	//	return -1
	//}
	//if k1.data > k2.data {
	//	return 1
	//}
	//return 0
}

type keyType struct {
	data [16]byte
}

type valType struct {
	data uint32 // implement
}

// record represents a record pointed to by a leaf node
type record struct {
	Key   keyType
	Value valType
}

func (r *record) Size() int64 {
	return int64(unsafe.Sizeof(r.Key.data) + unsafe.Sizeof(r.Value.data))
}

const M = 64 // 128

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
		//ss += fmt.Sprintf("%.2d", n.keys[i].data)
		ss += fmt.Sprintf("%v", n.keys[i].data)
		ss += fmt.Sprintf(",")
	}
	//ss += fmt.Sprintf("%.2d]", n.keys[n.numKeys-1].data)
	ss += fmt.Sprintf("%v]", n.keys[n.numKeys-1].data)
	return ss
}

// NewBPTree initializes a new tree
func NewBPTree() (*BPTree, error) {
	bpt := new(BPTree)
	return bpt, nil
}

// Has returns a boolean indicating weather or not
// the provided key and associated record exists.
func (t *BPTree) Has(k keyType) bool {
	return t.findEntry(k) != nil
}

// Add inserts a new record using the provided key. It
// only inserts an record if the key does not already exist.
func (t *BPTree) Add(k keyType, v valType) {
	// master insertUnique method only inserts if the key
	// does not currently exist in the tree
	t.insertUnique(k, v)
}

// Put is mainly used when you wish to upsert as it assumes the
// data to already be contained the tree. It will  overwrite
// duplicate keys, as it does not check to see if the key exists
func (t *BPTree) Put(k keyType, v valType) bool {
	// master insert method treats insertion much like
	// "setting" in a hashmap (an upsert) by default
	return t.insert(k, v)
}

// Get returns the record for a given key if it exists
func (t *BPTree) Get(k keyType) (keyType, valType) {
	e := t.findEntry(k)
	if e == nil {
		return *new(keyType), *new(valType)
	}
	return e.Key, e.Value
}

// Del removes the record for the supplied key and attempts
// to return the previous key and value
func (t *BPTree) Del(k keyType) (keyType, valType) {
	e := t.delete(k)
	if e == nil {
		return *new(keyType), *new(valType)
	}
	return e.Key, e.Value
}

// Range provides a simple iteration function for the tree
func (t *BPTree) Range(iter func(k keyType, v valType) bool) {
	c := findFirstLeaf(t.root)
	if c == nil {
		return
	}
	var e *record
	for {
		for i := 0; i < c.numKeys; i++ {
			e = (*record)(c.ptrs[i])
			if e != nil && !iter(e.Key, e.Value) {
				continue
			}
		}
		if c.ptrs[order-1] != nil {
			c = (*node)(c.ptrs[order-1])
		} else {
			break
		}
	}
}

// Min returns the minimum (lowest) key and value pair in the tree
func (t *BPTree) Min() (keyType, valType) {
	c := findFirstLeaf(t.root)
	if c == nil {
		return *new(keyType), *new(valType)
	}
	e := (*record)(c.ptrs[0])
	return e.Key, e.Value
}

// Max returns the maximum (highest) key and value pair in the tree
func (t *BPTree) Max() (keyType, valType) {
	c := findLastLeaf(t.root)
	if c == nil {
		return *new(keyType), *new(valType)
	}
	e := (*record)(c.ptrs[c.numKeys-1])
	return e.Key, e.Value
}

// GetClosest attempts to return the closest match in the tree
// if an explicit match cannot be found
func (t *BPTree) GetClosest(k keyType) (keyType, valType) {
	l := findLeaf(t.root, k)
	if l == nil {
		return *new(keyType), *new(valType)
	}
	e, ok := l.closest(k)
	if !ok {
		return *new(keyType), *new(valType)
	}
	return e.Key, e.Value
}

// Len returns the a count of the number of items in the tree
func (t *BPTree) Len() int {
	var count int
	for n := findFirstLeaf(t.root); n != nil; n = n.nextLeaf() {
		count += n.numKeys
	}
	return count
}

// Close closes the tree
func (t *BPTree) Close() {
	// t.destroyTree()
	t.root = nil
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
			if compare(k, n.keys[i]) == 0 {
				//if k.data == n.keys[i].data {
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
			//if k.data < n.keys[i].data {
			//	break
			//}
			if compare(k, n.keys[i]) < 0 {
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
			if compare(k, n.keys[i]) == 0 {
				//if k.data == n.keys[i].data {
				return (*record)(n.ptrs[i]), true
			}
		}
	}
	return nil, false
}

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

	for compare(n.keys[i], k) != 0 {
		//for n.keys[i].data != k.data {
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

// find, finds and returns the node and record to which a key refers
func (t *BPTree) find(k keyType) (*node, *record) {
	leaf := findLeaf(t.root, k)
	if leaf == nil {
		return nil, nil
	}
	// if the leaf returned by findLeaf != nil then the leaf must contain a
	// value, even if it does not contain the desired key. the leaf holds
	// the range of keys that would include the desired key
	var i int
	for i = 0; i < leaf.numKeys; i++ {
		if compare(leaf.keys[i], k) == 0 {
			//if leaf.keys[i].data == k.data {
			break
		}
	}
	if i == leaf.numKeys {
		return leaf, nil
	}
	return leaf, (*record)(leaf.ptrs[i])
}

// findLeaf traces the path from the root to a leaf, searching by key.
// findLeaf returns the leaf containing the given key
func findLeaf(root *node, k keyType) *node {
	if root == nil {
		return root
	}
	i, c := 0, root
	for !c.isLeaf {
		i = 0
		for i < c.numKeys {
			//if k.data >= c.keys[i].data {
			//	i++
			if compare(k, c.keys[i]) >= 0 {
				i++
			} else {
				break
			}
		}
		c = (*node)(c.ptrs[i])
	}
	// c is the found leaf node
	return c
}

// findEntry finds and returns the record to which a key refers. It is for all
// practical purposes identical to find(), it just does not return the leaf
// like find does, mechanically you don't save any more time or space using this
// version. consider removing it
func (t *BPTree) findEntry(k keyType) *record {
	leaf := findLeaf(t.root, k)
	if leaf == nil {
		return nil
	}
	// if the leaf returned by findLeaf != nil then the leaf must contain a
	// value, even if it does not contain the desired key. the leaf holds
	// the range of keys that would include the desired key
	var i int
	for i = 0; i < leaf.numKeys; i++ {
		if compare(leaf.keys[i], k) == 0 {
			//if leaf.keys[i].data == k.data {
			break
		}
	}
	if i == leaf.numKeys {
		return nil
	}
	return (*record)(leaf.ptrs[i])
}

// findFirstLeaf traces the path from the root to the leftmost leaf in the tree
func findFirstLeaf(root *node) *node {
	if root == nil {
		return root
	}
	c := root
	for !c.isLeaf {
		c = (*node)(c.ptrs[0])
	}
	return c
}

// findLastLeaf traces the path from the root to the rightmost leaf in the tree
func findLastLeaf(root *node) *node {
	if root == nil {
		return root
	}
	c := root
	for !c.isLeaf {
		c = (*node)(c.ptrs[c.numKeys])
	}
	return c
}

// insert is the "master" insertion function. it inserts a key and an associated
// value into the tree causing the tree to be adjusted however necessary to
// maintain the tree's properties
func (t *BPTree) insert(k keyType, v valType) bool {
	// if the root is nil, then the tree does not exist yet, start a new tree
	if t.root == nil {
		t.root = startNewTree(k, &record{k, v})
		return false
	}
	// the current implementation ignores duplicates (will treat it kind of
	// like a map's set operation), use insertUnique() if you wish to support
	// an add only type of action.
	leaf, recordPointer := t.find(k)
	if recordPointer != nil {
		// If the key already exists in this tree then we can simply proceed
		// to just update the value of the record pointer that was returned
		recordPointer.Value = v
		return true
	}
	// if we are here, then no existing record has been found in the tree. now
	// we must create a new record. it is worth mentioning that normally t.find
	// would not return a record pointer in which case we would need to do
	// something like the following:
	// recordPointer = makeRecord(v)

	// check to see if the leaf (that the record should go into) has room, and
	// if it does, simply insert into the leaf and return
	if leaf.numKeys < order-1 {
		insertIntoLeaf(leaf, k, &record{k, v})
		return false
	}

	// otherwise, leaf does not have enough room and needs to be split
	t.root = insertIntoLeafAfterSplitting(t.root, leaf, k, &record{k, v})
	return false
}

// insertUnique inserts a new record using the provided key. it only inserts
// a record if the key does not already exist
func (t *BPTree) insertUnique(k keyType, v valType) {
	// if the root is nil, then the tree does not exist yet, start a new tree
	if t.root == nil {
		t.root = startNewTree(k, &record{k, v})
		return
	}
	// see what we get when we try to find the correct leaf
	leaf := findLeaf(t.root, k)
	// check to ensure the leaf node does already contain the key
	if leaf.hasKey(k) {
		// if this is true, then they key already exists, so we
		// should just return
		return
	}

	// looks like it is not already in the tree, so now we must check
	// to see if the leaf (that the record should go into) has room,
	// and if it does, simply insert into the leaf and return
	if leaf.numKeys < order-1 {
		insertIntoLeaf(leaf, k, &record{k, v})
		return
	}

	// otherwise, leaf does not have enough room and needs to be split
	t.root = insertIntoLeafAfterSplitting(t.root, leaf, k, &record{k, v})
}

// startNewTree first insertion case: starts a new tree
func startNewTree(k keyType, ptr *record) *node {
	root := &node{isLeaf: true}
	root.keys[0] = k
	root.ptrs[0] = unsafe.Pointer(ptr)
	root.ptrs[order-1] = nil
	root.parent = nil
	root.numKeys++
	return root
}

// insertIntoLeaf inserts a new pointer to a Record and its
// corresponding key into a leaf.
func insertIntoLeaf(leaf *node, k keyType, ptr *record) /* *node */ {
	var i, insertionPoint int
	for insertionPoint < leaf.numKeys && compare(leaf.keys[insertionPoint], k) < 0 {
		//for insertionPoint < leaf.numKeys && leaf.keys[insertionPoint].data < k.data {
		insertionPoint++
	}
	for i = leaf.numKeys; i > insertionPoint; i-- {
		leaf.keys[i] = leaf.keys[i-1]
		leaf.ptrs[i] = leaf.ptrs[i-1]
	}
	leaf.keys[insertionPoint] = k
	leaf.ptrs[insertionPoint] = unsafe.Pointer(ptr)
	leaf.numKeys++
	// return leaf // might not need to return this leaf
}

// insertIntoLeafAfterSplitting is specifically called to insert a key and value when
// the leaf node is full (aka, exceeds the order of the tree) and the leaf must be split
// in half, and then re-balance upward toward the root
func insertIntoLeafAfterSplitting(root, leaf *node, k keyType, pointer *record) *node {

	// perform linear search to find index to insert new record
	var insertionIndex int

	for insertionIndex < order-1 && compare(leaf.keys[insertionIndex], k) < 0 {
		//for insertionIndex < order-1 && leaf.keys[insertionIndex].data < k.data {
		insertionIndex++
	}

	// initialize temporary variables
	var i, j int
	var tempKeys [order]keyType
	var tempPointers [order]unsafe.Pointer

	// copy leaf keys and ptrs to temp sets
	// reserve space at insertion index for new record
	for i, j = 0, 0; i < leaf.numKeys; i, j = i+1, j+1 {
		if j == insertionIndex {
			j++
		}
		tempKeys[j] = leaf.keys[i]
		tempPointers[j] = leaf.ptrs[i]
	}

	tempKeys[insertionIndex] = k
	tempPointers[insertionIndex] = unsafe.Pointer(pointer)

	leaf.numKeys = 0

	// find pivot index where to split leaf
	split := cut(order - 1)

	// overwrite original leaf up to the split point
	for i = 0; i < split; i++ {
		leaf.keys[i] = tempKeys[i]
		leaf.ptrs[i] = tempPointers[i]
		leaf.numKeys++
	}

	// create new leaf
	newLeaf := &node{isLeaf: true} // makeLeaf()

	// writing to new leaf from split point to end of original leaf pre-split
	for i, j = split, 0; i < order; i, j = i+1, j+1 {
		newLeaf.keys[j] = tempKeys[i]
		newLeaf.ptrs[j] = tempPointers[i]
		newLeaf.numKeys++
	}

	// free temps
	for i = 0; i < order; i++ {
		tempKeys[i] = *new(keyType) // zero Value
		tempPointers[i] = nil       // zero Value
	}

	newLeaf.ptrs[order-1] = leaf.ptrs[order-1]
	leaf.ptrs[order-1] = unsafe.Pointer(newLeaf)

	for i = leaf.numKeys; i < order-1; i++ {
		leaf.ptrs[i] = nil
	}
	for i = newLeaf.numKeys; i < order-1; i++ {
		newLeaf.ptrs[i] = nil
	}

	newLeaf.parent = leaf.parent
	newKey := newLeaf.keys[0]

	// call insertIntoParent to ensure the tree gets balanced back
	// up to the root
	return insertIntoParent(root, leaf, newKey, newLeaf)
}

// insertIntoParent inserts a new node (leaf or internal node) into the tree and returns the root
// of the tree after insertion is complete
func insertIntoParent(root *node, left *node, k keyType, right *node) *node {

	// this is the case if the left parent is the root
	if left.parent == nil {
		return insertIntoNewRoot(left, k, right)
	}

	// otherwise, we are not dealing with the parent node being the root, so we must find the
	// parent's pointer to the left node
	leftIndex := getLeftIndex(left.parent, left)

	// check to see if the new key fits into the left parent
	if left.parent.numKeys < order-1 {
		return insertIntoNode(root, left.parent, leftIndex, k, right)
	}

	// otherwise, it doesn't fit, so we need to split upward
	return insertIntoNodeAfterSplitting(root, left.parent, leftIndex, k, right)
}

// insertIntoNewRoot creates a new root for two subtrees and inserts the appropriate key into the new root
func insertIntoNewRoot(left *node, k keyType, right *node) *node {
	root := &node{} // makeNode()
	root.keys[0] = k
	root.ptrs[0] = unsafe.Pointer(left)
	root.ptrs[1] = unsafe.Pointer(right)
	root.numKeys++
	root.parent = nil
	left.parent = root
	right.parent = root
	return root
}

// getLeftIndex helper function used in insertIntoParent to find the index of the parent's pointer to the
// node to the left of the key to be inserted
func getLeftIndex(parent, left *node) int {
	var leftIndex int
	for leftIndex <= parent.numKeys && (*node)(parent.ptrs[leftIndex]) != left {
		leftIndex++
	}
	return leftIndex
}

// insertIntoNode inserts a new key and pointer to a node into a node into which these can fit
// without violating the tree's properties
func insertIntoNode(root, n *node, leftIndex int, k keyType, right *node) *node {
	// Consider using copy, it might be better
	copy(n.ptrs[leftIndex+2:], n.ptrs[leftIndex+1:])
	copy(n.keys[leftIndex+1:], n.keys[leftIndex:])

	// this for loop is the original implementation, for what it's worth
	// for i := n.numKeys; i > leftIndex; i-- {
	//	n.ptrs[i+1] = n.ptrs[i]
	//	n.keys[i] = n.keys[i-1]
	// }

	n.ptrs[leftIndex+1] = unsafe.Pointer(right)
	n.keys[leftIndex] = k
	n.numKeys++
	return root
}

// insertIntoNodeAfterSplitting inserts a new key and pointer to a node into a node, causing
// the nodes size to exceed the tree's order, and causing the node to split
func insertIntoNodeAfterSplitting(root, oldNode *node, leftIndex int, k keyType, right *node) *node {
	// first create a temp set of keys and ptrs to hold everything, including the new key and
	// pointer inserted in their correct places--then create a new node and copy half of the
	// keys and ptrs to the old node and the other half to the new

	// initialize temporary variables
	var i, j int
	var tempKeys [order]keyType
	var tempPointers [order + 1]unsafe.Pointer

	// load up the pointers into the temporary set
	for i, j = 0, 0; i < oldNode.numKeys+1; i, j = i+1, j+1 {
		if j == leftIndex+1 {
			j++
		}
		tempPointers[j] = oldNode.ptrs[i]
	}

	// load up the keys into the temporary set
	for i, j = 0, 0; i < oldNode.numKeys; i, j = i+1, j+1 {
		if j == leftIndex {
			j++
		}
		tempKeys[j] = oldNode.keys[i]
	}

	// set the temporary index pointers to their new location
	tempPointers[leftIndex+1] = unsafe.Pointer(right)
	tempKeys[leftIndex] = k

	// get the split index
	split := cut(order)

	// "reset" the "old node"
	oldNode.numKeys = 0

	// put half (left/first half) of the temporary pointers and keys into the old node
	for i = 0; i < split-1; i++ {
		oldNode.ptrs[i] = tempPointers[i]
		oldNode.keys[i] = tempKeys[i]
		oldNode.numKeys++
	}
	oldNode.ptrs[i] = tempPointers[i]
	kPrime := tempKeys[split-1]

	// create a new node which will become the right child node
	newNode := &node{} // makeNode()

	// ...and copy the other half (right/last half) of the temporary keys and
	// pointers into the new node
	for i, j = i+1, 0; i < order; i, j = i+1, j+1 {
		newNode.ptrs[j] = tempPointers[i]
		newNode.keys[j] = tempKeys[i]
		newNode.numKeys++
	}
	newNode.ptrs[j] = tempPointers[i]
	newNode.parent = oldNode.parent

	// finally, free up the temporary keys and pointers, we're done with them
	for i = 0; i < order; i++ {
		tempKeys[i] = *new(keyType) // zero values
		tempPointers[i] = nil       // zero values
	}
	tempPointers[order] = nil

	// create a child that will contain the value pointers of the newly split
	// new node, and make the child node's parent, the new node (not sure i
	// remember how this part actually works)
	var child *node
	for i = 0; i <= newNode.numKeys; i++ {
		child = (*node)(newNode.ptrs[i])
		child.parent = newNode
	}

	// and then finally, insert new key into the parent of the two nodes resulting
	// from the split with the old node to the left, and the new node to the right
	return insertIntoParent(root, oldNode, kPrime, newNode)
}

var queue *print = nil

type print struct {
	node *node
	next *print
}

// String is node's stringer method
func (n *print) String() string {
	ss := fmt.Sprintf("\tr%dn%d[", height(n.node), pathToRoot(n.node.parent, n.node))
	for i := 0; i < n.node.numKeys-1; i++ {
		//ss += fmt.Sprintf("%.2d", n.node.keys[i].data)
		ss += fmt.Sprintf("%v", n.node.keys[i].data)
		ss += fmt.Sprintf(",")
	}
	//ss += fmt.Sprintf("%.2d]", n.node.keys[n.node.numKeys-1].data)
	ss += fmt.Sprintf("%v]", n.node.keys[n.node.numKeys-1].data)
	return ss
}

func nodeID(n *node) string {
	ss := fmt.Sprintf("h%.4xk", height(n))
	for i := 0; i < n.numKeys-1; i++ {
		ss += fmt.Sprintf("%.4x", n.keys[i].data)
	}
	ss += fmt.Sprintf("%.4x", n.keys[n.numKeys-1].data)
	return ss
}

func printNodeMarkdown(n *node) {
	ss := fmt.Sprintf("\t%s[", nodeID(n))
	for i := 0; i < n.numKeys-1; i++ {
		//ss += fmt.Sprintf("%.2d", n.keys[i].data)
		ss += fmt.Sprintf("%v", n.keys[i].data)
		ss += fmt.Sprintf(",")
	}
	//ss += fmt.Sprintf("%.2d]", n.keys[n.numKeys-1].data)
	ss += fmt.Sprintf("%v]", n.keys[n.numKeys-1].data)
	if !n.isLeaf {
		cc := make([]string, n.numKeys)
		for i := 0; i <= n.numKeys; i++ {
			child := (*node)(n.ptrs[i])
			cc = append(cc, fmt.Sprintf("%s --- %s", ss, nodeID(child)))
		}
		ss = strings.Join(cc, "\n")
	}
	fmt.Println(ss)
}

func (n *node) _String() string {
	ss := fmt.Sprintf("[")
	for i := 0; i < n.numKeys-1; i++ {
		//ss += fmt.Sprintf("%d|", n.keys[i].data)
		ss += fmt.Sprintf("%v|", n.keys[i].data)
	}
	//ss += fmt.Sprintf("%d]", n.keys[n.numKeys-1].data)
	ss += fmt.Sprintf("%v]", n.keys[n.numKeys-1].data)
	return ss
}

func newPrint(n *node) *print {
	return &print{
		node: n,
		next: nil,
	}
}

func enqueue(newNode *node) {
	var c *print
	if queue == nil {
		queue = newPrint(newNode)
		queue.next = nil
	} else {
		c = queue
		for c.next != nil {
			c = c.next
		}
		c.next = newPrint(newNode)
		// newNode.next = nil
	}
}

func dequeue() *print {
	var n *print
	n = queue
	queue = queue.next
	n.next = nil
	return n
}

func printLeaves(root *node) {
	if root == nil {
		fmt.Println("empty tree")
		return
	}
	var c *node
	c = root
	for !c.isLeaf {
		c = (*node)(c.ptrs[0])
	}
	for {
		for i := 0; i < c.numKeys; i++ {
			//fmt.Printf("%d ", c.keys[i])
			fmt.Printf("%v ", c.keys[i])
			fmt.Printf("%p ", c.ptrs[i])
		}
		fmt.Printf("%p ", c.ptrs[order-1])
		if c.ptrs[order-1] != nil {
			fmt.Printf(" | ")
			c = (*node)(c.ptrs[order-1])
		} else {
			break
		}
		fmt.Printf("\n")
	}
}

// height is a utility function to give the height of the tree, which
// length in number of edges of the path from the root to any leaf
func height(root *node) int {
	h := 0
	var c *node
	c = root
	for !c.isLeaf {
		c = (*node)(c.ptrs[0])
		h++
	}
	return h
}

// pathToRoot is a utility function to give the length in edges of
// the path from any node to the root
func pathToRoot(root *node, child *node) int {
	length := 0
	var c *node
	c = child
	for c != root {
		c = c.parent
		length++
	}
	return length
}
func printTree(root *node) {
	var rank, newRank int
	if root == nil {
		fmt.Println("empty tree")
		return
	}
	queue = nil
	enqueue(root)
	fmt.Println("graph TD")
	fmt.Printf("\ttitle{B+Tree of order %d}\n", M)
	for queue != nil {
		n := dequeue()
		if n.node.parent != nil && n.node == (*node)(n.node.parent.ptrs[0]) {
			newRank = pathToRoot(root, n.node)
			if newRank != rank {
				rank = newRank
			}
		}
		printNodeMarkdown(n.node)
		if !n.node.isLeaf {
			for i := 0; i <= n.node.numKeys; i++ {
				enqueue((*node)(n.node.ptrs[i]))
			}
		}
	}
}

func _printTree(root *node) {
	var rank, newRank int
	if root == nil {
		fmt.Println("empty tree")
		return
	}
	queue = nil
	// put the root node in the current node
	enqueue(root)
	var nn int
	for queue != nil {
		// get the current node out of the queue
		n := dequeue()
		if n.node.parent != nil && n.node == (*node)(n.node.parent.ptrs[0]) {
			newRank = pathToRoot(root, n.node)
			if newRank != rank {
				// fmt.Printf(" (level=%d)", rank)
				rank = newRank
				nn = 0
				fmt.Printf("\n")
			}
		}
		fmt.Printf("\tr%dn%d[", rank, nn)
		for i := 0; i < n.node.numKeys-1; i++ {
			//fmt.Printf("%.2d", n.node.keys[i].data)
			fmt.Printf("%v", n.node.keys[i].data)
			fmt.Printf(",")
		}
		//fmt.Printf("%.2d]", n.node.keys[n.node.numKeys-1].data)
		fmt.Printf("%v]", n.node.keys[n.node.numKeys-1].data)
		fmt.Printf(" --> ")
		fmt.Printf("\n")
		nn++
		// if not a leaf, queue up the child pointers
		if !n.node.isLeaf {
			for i := 0; i <= n.node.numKeys; i++ {
				child := (*node)(n.node.ptrs[i])
				enqueue(child)
			}
		}
		// if it is a leaf, print the values
		// if n.node.isLeaf {
		// 	fmt.Printf("%s", (*node)(n.node.ptrs[order-1]))
		// } else {
		// 	fmt.Printf("%s", (*node)(n.node.ptrs[n.node.numKeys]))
		// }
		// fmt.Printf(" | ")
	}
	fmt.Printf("\n")
}

var ident = map[int]string{
	0: "\r\t\t\t\t\t\t\t\t\t\t\t\t",
	1: "\r\t\t\t\t\t\t\t\t\t\t\t",
	2: "\r\t\t\t\t\t\t\t\t",
	3: "\r",
	4: "\r",
	5: "\r",
}

func print_tree(root *node) {
	fmt.Println("Printing Tree...")
	var i, rank, new_rank int
	if root == nil {
		fmt.Printf("Empty tree.\n")
		return
	}
	queue = nil
	enqueue(root)
	for queue != nil {
		prt := dequeue()
		if prt.node.parent != nil && prt.node == (*node)(prt.node.parent.ptrs[0]) {
			new_rank = pathToRoot(root, prt.node)
			if new_rank != rank {
				rank = new_rank
				fmt.Printf("\n%s", ident[rank])
			}
		}
		if rank == 0 {
			fmt.Printf("%s", ident[rank])
		}
		fmt.Printf("[")
		for i = 0; i < prt.node.numKeys-1; i++ {
			//fmt.Printf("%d|", prt.node.keys[i].data)
			fmt.Printf("%v|", prt.node.keys[i].data)
		}
		//fmt.Printf("%d]", prt.node.keys[prt.node.numKeys-1].data)
		fmt.Printf("%v]", prt.node.keys[prt.node.numKeys-1].data)
		if !prt.node.isLeaf {
			for i = 0; i <= prt.node.numKeys; i++ {
				enqueue((*node)(prt.node.ptrs[i]))
			}
		}
		fmt.Printf("  ")
	}
	fmt.Printf("\n\n")
}

func print_tree_v2(root *node) {
	fmt.Println("Printing Tree...")
	var i, rank, new_rank int
	if root == nil {
		fmt.Printf("Empty tree.\n")
		return
	}
	queue = nil
	enqueue(root)
	for queue != nil {
		prt := dequeue()
		if prt.node.parent != nil && prt.node == (*node)(prt.node.parent.ptrs[0]) {
			new_rank = pathToRoot(root, prt.node)
			if new_rank != rank {
				rank = new_rank
				fmt.Printf("\n%s", ident[rank])
			}
		}
		if rank == 0 {
			fmt.Printf("%s", ident[rank])
		}
		fmt.Printf("[")
		for i = 0; i < prt.node.numKeys-1; i++ {
			//fmt.Printf("%d|", prt.node.keys[i].data)
			fmt.Printf("%v|", prt.node.keys[i].data)
		}
		//fmt.Printf("%d]", prt.node.keys[prt.node.numKeys-1].data)
		fmt.Printf("%v]", prt.node.keys[prt.node.numKeys-1].data)
		if !prt.node.isLeaf {
			for i = 0; i <= prt.node.numKeys; i++ {
				enqueue((*node)(prt.node.ptrs[i]))
			}
		}
		fmt.Printf("  ")
	}
	fmt.Printf("\n\n")
}

func print_markdown_tree(root *node) {
	var sss [][]string
	var i, rank, new_rank int
	if root == nil {
		sss = append(sss, []string{"root[ ]"})
		return
	}
	queue = nil
	enqueue(root)
	for queue != nil {
		var ss []string
		prt := dequeue()
		if prt.node.parent != nil && prt.node == (*node)(prt.node.parent.ptrs[0]) {
			new_rank = pathToRoot(root, prt.node)
			if new_rank != rank {
				rank = new_rank
				// fmt.Printf("\n%s (rank=%d)", ident[rank], rank)
				// ss = append(ss, fmt.Sprintf("r%dn", rank))
			}
		}
		if rank == 0 {
			// fmt.Printf("%s", ident[rank])
			//ss = append(ss, fmt.Sprintf("r%dn[%.2d]", rank, prt.node.keys[i].data))
			ss = append(ss, fmt.Sprintf("r%dn[%v]", rank, prt.node.keys[i].data))
		}
		// fmt.Printf("[")
		for i = 0; i < prt.node.numKeys-1; i++ {
			// fmt.Printf("%d|", prt.node.keys[i].data)
			//ss = append(ss, fmt.Sprintf("r%dn[%.2d]", rank, prt.node.keys[i].data))
			ss = append(ss, fmt.Sprintf("r%dn[%v]", rank, prt.node.keys[i].data))
		}
		// fmt.Printf("%d]", prt.node.keys[prt.node.numKeys-1].data)
		//ss = append(ss, fmt.Sprintf("r%dn[%.2d]**", rank, prt.node.keys[prt.node.numKeys-1].data))
		ss = append(ss, fmt.Sprintf("r%dn[%v]**", rank, prt.node.keys[prt.node.numKeys-1].data))
		if !prt.node.isLeaf {
			for i = 0; i <= prt.node.numKeys; i++ {
				enqueue((*node)(prt.node.ptrs[i]))
			}
		}
		sss = append(sss, ss)
		// fmt.Printf("  ")
	}
	// fmt.Printf("\n\n")

	for i := range sss {
		for j := range sss[i] {
			fmt.Printf("[%s] ", sss[i][j])
		}
		fmt.Printf("\n")
	}
}

func print_leaves(root *node) {
	fmt.Println("Printing Leaves...")
	var i int
	var c *node = root
	if root == nil {
		fmt.Printf("Empty tree.\n")
		return
	}
	for !c.isLeaf {
		c = (*node)(c.ptrs[0])
	}
	for {
		/*
			for i = 0; i < M-1; i++ {
				if c.keys[i] == nil {
					fmt.Printf("___, ")
					continue
				}
				//fmt.Printf("%s, ", c.keys[i])
				// extract record / value instead
				rec := (*record)(unsafe.Pointer(c.ptrs[i]))
				fmt.Printf("%s, ", rec.val)
			}
			if c.ptrs[M-1] != nil {
				fmt.Printf(" | ")
				c = (*node)(unsafe.Pointer(c.ptrs[M-1]))
			} else {
				break
			}
		*/
		for i = 0; i < M-1; i++ {
			if r := (*record)(c.ptrs[i]); r == nil {
				fmt.Printf("___, ")
				continue
			} else {
				fmt.Printf("%v ", r.Value)
			}
		}
		if c.ptrs[M-1] != nil {
			fmt.Printf(" || ")
			c = (*node)(c.ptrs[M-1])
		} else {
			break
		}
	}
	fmt.Printf("\n\n")
}
