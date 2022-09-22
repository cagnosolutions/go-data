package bpt

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
		if leaf.keys[i].data == k.data {
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
			if k.data >= c.keys[i].data {
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
		if leaf.keys[i].data == k.data {
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
