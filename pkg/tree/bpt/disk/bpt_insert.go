package disk

import "unsafe"

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
	for insertionPoint < leaf.numKeys && KeysCompare(leaf.keys[insertionPoint], k) < 0 {
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
	for insertionIndex < order-1 && KeysCompare(leaf.keys[insertionIndex], k) < 0 {
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
