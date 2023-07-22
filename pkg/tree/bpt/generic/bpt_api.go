package generic

// NewTree initializes a new tree
func NewTree[K Key, V any]() (*BPTree[K, V], error) {
	bpt := new(BPTree[K, V])
	return bpt, nil
}

// Has returns a boolean indicating weather or not
// the provided Key and associated record exists.
func (t *BPTree[K, V]) Has(k K) bool {
	return t.findEntry(k) != nil
}

// Add inserts a new record using the provided Key. It
// only inserts an record if the Key does not already exist.
func (t *BPTree[K, V]) Add(k K, v V) {
	// master insertUnique method only inserts if the Key
	// does not currently exist in the tree
	t.insertUnique(k, v)
}

// Put is mainly used when you wish to upsert as it assumes the
// data to already be contained the tree. It will  overwrite
// duplicate keys, as it does not check to see if the Key exists
func (t *BPTree[K, V]) Put(k K, v V) bool {
	// master insert method treats insertion much like
	// "setting" in a hashmap (an upsert) by default
	return t.insert(k, v)
}

// Get returns the record for a given Key if it exists
func (t *BPTree[K, V]) Get(k K) (key K, val V) {
	e := t.findEntry(k)
	if e == nil {
		return key, val
	}
	return e.Key, e.Value
}

// Del removes the record for the supplied Key and attempts
// to return the previous Key and value
func (t *BPTree[K, V]) Del(k K) (key K, val V) {
	e := t.delete(k)
	if e == nil {
		return key, val
	}
	return e.Key, e.Value
}

// Range provides a simple iteration function for the tree
func (t *BPTree[K, V]) Range(iter func(k K, v V) bool) {
	c := findFirstLeaf[K, V](t.root)
	if c == nil {
		return
	}
	var e *record[K, V]
	for {
		for i := 0; i < c.numKeys; i++ {
			e = (*record[K, V])(c.ptrs[i])
			if e != nil && !iter(e.Key, e.Value) {
				continue
			}
		}
		if c.ptrs[order-1] != nil {
			c = (*node[K, V])(c.ptrs[order-1])
		} else {
			break
		}
	}
}

// Min returns the minimum (lowest) Key and value pair in the tree
func (t *BPTree[K, V]) Min() (key K, val V) {
	c := findFirstLeaf[K, V](t.root)
	if c == nil {
		return key, val
	}
	e := (*record[K, V])(c.ptrs[0])
	return e.Key, e.Value
}

// Max returns the maximum (highest) Key and value pair in the tree
func (t *BPTree[K, V]) Max() (key K, val V) {
	c := findLastLeaf[K, V](t.root)
	if c == nil {
		return key, val
	}
	e := (*record[K, V])(c.ptrs[c.numKeys-1])
	return e.Key, e.Value
}

// GetClosest attempts to return the closest match in the tree
// if an explicit match cannot be found
func (t *BPTree[K, V]) GetClosest(k K) (key K, val V) {
	l := findLeaf[K, V](t.root, k)
	if l == nil {
		return key, val
	}
	e, ok := l.closest(k)
	if !ok {
		return key, val
	}
	return e.Key, e.Value
}

// Len returns the a count of the number of items in the tree
func (t *BPTree[K, V]) Len() int {
	var count int
	for n := findFirstLeaf[K, V](t.root); n != nil; n = n.nextLeaf() {
		count += n.numKeys
	}
	return count
}

// Close closes the tree
func (t *BPTree[K, V]) Close() {
	// t.destroyTree()
	t.root = nil
}
