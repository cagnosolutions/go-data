package generic

// find, finds and returns the node and record to which a Key refers
func (t *BPTree[K, V]) find(k K) (*node[K, V], *record[K, V]) {
	leaf := findLeaf(t.root, k)
	if leaf == nil {
		return nil, nil
	}
	// if the leaf returned by findLeaf != nil then the leaf must contain a
	// value, even if it does not contain the desired Key. the leaf holds
	// the range of keys that would include the desired Key
	var i int
	for i = 0; i < leaf.numKeys; i++ {
		if leaf.keys[i].Compare(k) == 0 {
			break
		}
	}
	if i == leaf.numKeys {
		return leaf, nil
	}
	return leaf, (*record[K, V])(leaf.ptrs[i])
}

// findLeaf traces the path from the root to a leaf, searching by Key.
// findLeaf returns the leaf containing the given Key
func findLeaf[K Key, V any](root *node[K, V], k K) *node[K, V] {
	if root == nil {
		return root
	}
	i, c := 0, root
	for !c.isLeaf {
		i = 0
		for i < c.numKeys {
			if k.Compare(c.keys[i]) >= 0 {
				i++
			} else {
				break
			}
		}
		c = (*node[K, V])(c.ptrs[i])
	}
	// c is the found leaf node
	return c
}

// findEntry finds and returns the record to which a Key refers. It is for all
// practical purposes identical to find(), it just does not return the leaf
// like find does, mechanically you don't save any more time or space using this
// version. consider removing it
func (t *BPTree[K, V]) findEntry(k K) *record[K, V] {
	leaf := findLeaf(t.root, k)
	if leaf == nil {
		return nil
	}
	// if the leaf returned by findLeaf != nil then the leaf must contain a
	// value, even if it does not contain the desired Key. the leaf holds
	// the range of keys that would include the desired Key
	var i int
	for i = 0; i < leaf.numKeys; i++ {
		if leaf.keys[i].Compare(k) == 0 {
			break
		}
	}
	if i == leaf.numKeys {
		return nil
	}
	return (*record[K, V])(leaf.ptrs[i])
}

// findFirstLeaf traces the path from the root to the leftmost leaf in the tree
func findFirstLeaf[K Key, V any](root *node[K, V]) *node[K, V] {
	if root == nil {
		return root
	}
	c := root
	for !c.isLeaf {
		c = (*node[K, V])(c.ptrs[0])
	}
	return c
}

// findLastLeaf traces the path from the root to the rightmost leaf in the tree
func findLastLeaf[K Key, V any](root *node[K, V]) *node[K, V] {
	if root == nil {
		return root
	}
	c := root
	for !c.isLeaf {
		c = (*node[K, V])(c.ptrs[c.numKeys])
	}
	return c
}
