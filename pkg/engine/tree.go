package engine

// node represents a b plus tree node.
type node = page

// ptr represents a pointer to another node.
type ptr = uint32

// numKeys returns the number of keys currently in this node.
func (n *node) numKeys() uint16 {
	return n.getNumCells() - n.getNumFree()
}

// parent returns the parent of the current node along with a
// boolean indicating true if the node has a parent and false
// if it does not.
func (n *node) parent() (ptr, bool) {

}

func (n *node) min() uint32 {
	if n.n.getPrev()
}

// pageTree is a b plus tree wrapping the page cache.
type pageTree struct {
	*pageCache
	root PageID
}

// newPageTree initializes and returns a new pageTree instance.
func newPageTree(path string) (*pageTree, error) {
	pc, err := openPageCache(path, 5)
	if err != nil {
		return nil, err
	}
	return &pageTree{
		pageCache: pc,
	}, nil
}

func encodeNode(n *node) {
	newPtrRecord()
}

// insert takes a record and inserts it into the tree causing the
// tree to be adjusted according to the value of the key, in order
// to maintain the tree's properties. If the record key already
// exists within the tree, it will perform an upsert.
func (pt *pageTree) insert(r *record) error {
	// If the root of the tree is
}

// findLeaf traces the path from the root node down to a leaf node,
// searching according to the record key. It returns the leaf node
// containing the given key.
func (pt *pageTree) findLeaf(k []byte) uint32 {

}
