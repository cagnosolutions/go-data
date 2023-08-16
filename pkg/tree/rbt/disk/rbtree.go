package disk

import (
	"encoding/binary"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
	"unsafe"
)

const (
	RED   = 0
	BLACK = 1
)

type Entry struct {
	Key string
	Val []byte
}

func (e *Entry) ReadAt(p []byte, off int64) (n int, err error) {
	// TODO implement me
	panic("implement me")
}

func (e *Entry) WriteAt(p []byte, off int64) (n int, err error) {
	// write key length
	binary.BigEndian.PutUint16(p[off+0:off+2], uint16(len(e.Key)))
	// write val length
	binary.BigEndian.PutUint16(p[off+2:off+4], uint16(len(e.Val)))
	n += 4
	// write key
	n += copy(p[int(off)+n:int(off)+n+len(e.Key)], e.Key)
	// write value
	n += copy(p[int(off)+n:int(off)+n+len(e.Key)+len(e.Val)], e.Val)
	return n, err
}

func (e *Entry) isNil() bool {
	return e.Key == "" && e.Val == nil
}

func (e *Entry) Size() int {
	return len(e.Key) + int(unsafe.Sizeof(e.Val))
}

func (e *Entry) String() string {
	return fmt.Sprintf("{k:%q,v:%v}", e.Key, e.Val)
}

var empty = new(Entry)

func compare(a, b *Entry) int {
	return strings.Compare(a.Key, b.Key)
}

type rbNode struct {
	left   unsafe.Pointer
	right  unsafe.Pointer
	parent unsafe.Pointer
	color  uint8
	entry  *Entry
}

func (n *rbNode) ReadAt(p []byte, off int64) (c int, err error) {
	// TODO implement me
	panic("implement me")
}

func (n *rbNode) WriteAt(p []byte, off int64) (c int, err error) {
	// write left pointer
	binary.BigEndian.PutUint64(p[off+0:off+8], *(*uint64)(n.left))
	// write right pointer
	binary.BigEndian.PutUint64(p[off+8:off+16], *(*uint64)(n.right))
	// write parent pointer
	binary.BigEndian.PutUint64(p[off+16:off+24], *(*uint64)(n.parent))
	// write color
	p[off+24] = n.color
	// write entry
	c, err = n.entry.WriteAt(p, off+24)
	return
}

func newRBNode(w io.OffsetWriter) *rbNode {
	return nil
}

func (n *rbNode) getLeftPtr() unsafe.Pointer {
	return n.left
}

func (n *rbNode) getRightPtr() unsafe.Pointer {
	return n.right
}

func (n *rbNode) getParent() unsafe.Pointer {
	return n.parent
}

func (n *rbNode) assnLeftPtr(left unsafe.Pointer) {
	n.left = unsafe.Pointer(left)
}

func (n *rbNode) assnRightPtr(right unsafe.Pointer) {
	n.right = unsafe.Pointer(right)
}

func (n *rbNode) assnParentPtr(parent unsafe.Pointer) {
	n.parent = unsafe.Pointer(parent)
}

func (n *rbNode) getGrandparent() unsafe.Pointer {
	return (*rbNode)(n.getParent()).getParent()
}

func (n *rbNode) getLeftUncle() unsafe.Pointer {
	return (*rbNode)(n.getParent()).getLeftPtr()
}

func (n *rbNode) getRightUncle() unsafe.Pointer {
	return (*rbNode)(n.getParent()).getRightPtr()
}

type RBTree = rbTree

// rbTree is a struct representing a rbTree
type rbTree struct {
	lock  sync.RWMutex
	NIL   *rbNode
	root  *rbNode
	count int
	size  int64
}

func NewRBTree() *rbTree {
	return newRBTree()
}

// NewTree creates and returns a new rbTree
func newRBTree() *rbTree {
	n := &rbNode{
		left:   nil,
		right:  nil,
		parent: nil,
		color:  BLACK,
		entry:  empty,
	}
	return &rbTree{
		NIL:   n,
		root:  n,
		count: 0,
	}
}

func (t *rbTree) GetClone() *rbTree {
	t.Lock()
	defer t.Unlock()
	clone := newRBTree()
	t.cloneEntries(clone)
	return clone
}

func (t *rbTree) Lock() {
	t.lock.Lock()
}

func (t *rbTree) Unlock() {
	t.lock.Unlock()
}

func (t *rbTree) RLock() {
	t.lock.RLock()
}

func (t *rbTree) RUnlock() {
	t.lock.RUnlock()
}

// Has tests and returns a boolean value if the
// provided key exists in the tree
func (t *rbTree) Has(key string) bool {
	_, ok := t.getInternal(key)
	return ok
}

// Add adds the provided key and value only if it does not
// already exist in the tree. It returns false if the key and
// value was not able to be added, and true if it was added
// successfully
func (t *rbTree) Add(key string, val []byte) bool {
	_, ok := t.getInternal(key)
	if ok {
		// key already exists, so we are not adding
		return false
	}
	t.putInternal(key, val)
	return true
}

func (t *rbTree) Put(key string, val []byte) (any, bool) {
	return t.putInternal(key, val)
}

func (t *rbTree) putInternal(key string, val []byte) (*Entry, bool) {
	if key == "" {
		return nil, false
	}
	// insert returns the node inserted
	// and if the node returned already
	// existed and/or was updated
	ret, ok := t.insert(
		&rbNode{
			left:   unsafe.Pointer(t.NIL),
			right:  unsafe.Pointer(t.NIL),
			parent: unsafe.Pointer(t.NIL),
			color:  RED,
			entry:  &Entry{key, val},
		},
	)
	return ret.entry, ok
}

func (t *rbTree) Get(key string) (*Entry, bool) {
	return t.getInternal(key)
}

// GetNearMin performs an approximate search for the specified key
// and returns the closest key that is less than (the predecessor)
// to the searched key as well as a boolean reporting true if an
// exact match was found for the key, and false if it is unknown
// or and exact match was not found
func (t *rbTree) GetNearMin(key string) (*Entry, bool) {
	if key == "" {
		return nil, false
	}
	entry := &Entry{key, nil}
	ret := t.searchApprox(
		&rbNode{
			left:   unsafe.Pointer(t.NIL),
			right:  unsafe.Pointer(t.NIL),
			parent: unsafe.Pointer(t.NIL),
			color:  RED,
			entry:  entry,
		},
	)
	prev := t.predecessor(ret).entry
	if prev == nil {
		prev, _ = t.Min()
	}
	return prev, compare(ret.entry, entry) == 0
}

// GetNearMax performs an approximate search for the specified key
// and returns the closest key that is greater than (the successor)
// to the searched key as well as a boolean reporting true if an
// exact match was found for the key, and false if it is unknown or
// and exact match was not found
func (t *rbTree) GetNearMax(key string) (*Entry, bool) {
	if key == "" {
		return nil, false
	}
	entry := &Entry{key, nil}
	ret := t.searchApprox(
		&rbNode{
			left:   unsafe.Pointer(t.NIL),
			right:  unsafe.Pointer(t.NIL),
			parent: unsafe.Pointer(t.NIL),
			color:  RED,
			entry:  entry,
		},
	)
	return t.successor(ret).entry, compare(ret.entry, entry) == 0
}

// GetApproxPrevNext performs an approximate search for the specified key
// and returns the searched key, the predecessor, and the successor and a
// boolean reporting true if an exact match was found for the key, and false
// if it is unknown or and exact match was not found
func (t *rbTree) GetApproxPrevNext(key string) (*Entry, *Entry, *Entry, bool) {
	if key == "" {
		return nil, nil, nil, false
	}
	entry := &Entry{key, nil}
	ret := t.searchApprox(
		&rbNode{
			left:   unsafe.Pointer(t.NIL),
			right:  unsafe.Pointer(t.NIL),
			parent: unsafe.Pointer(t.NIL),
			color:  RED,
			entry:  entry,
		},
	)
	return ret.entry, t.predecessor(ret).entry, t.successor(ret).entry, compare(ret.entry, entry) == 0
}

func (t *rbTree) getInternal(key string) (*Entry, bool) {
	if key == "" {
		return nil, false
	}
	entry := &Entry{key, nil}
	ret := t.search(
		&rbNode{
			left:   unsafe.Pointer(t.NIL),
			right:  unsafe.Pointer(t.NIL),
			parent: unsafe.Pointer(t.NIL),
			color:  RED,
			entry:  entry,
		},
	)
	return ret.entry, !ret.entry.isNil()
}

func (t *rbTree) Del(key string) (*Entry, bool) {
	return t.delInternal(key)
}

func (t *rbTree) delInternal(key string) (*Entry, bool) {
	if key == "" {
		return nil, false
	}
	cnt := t.count
	entry := &Entry{key, nil}
	ret := t.delete(
		&rbNode{
			left:   unsafe.Pointer(t.NIL),
			right:  unsafe.Pointer(t.NIL),
			parent: unsafe.Pointer(t.NIL),
			color:  RED,
			entry:  entry,
		},
	)
	return ret.entry, cnt == t.count+1
}

func (t *rbTree) Len() int {
	return t.count
}

// Size returns the size in bytes
func (t *rbTree) Size() int64 {
	return t.size
}

func (t *rbTree) Min() (*Entry, bool) {
	x := t.min(t.root)
	if x == t.NIL {
		return nil, false
	}
	return x.entry, true
}

func (t *rbTree) Max() (*Entry, bool) {
	x := t.max(t.root)
	if x == t.NIL {
		return nil, false
	}
	return x.entry, true
}

// helper function for clone
func (t *rbTree) cloneEntries(t2 *rbTree) {
	t.ascend(
		t.root, t.min(t.root).entry, func(key string, val []byte) bool {
			t2.putInternal(key, val)
			return true
		},
	)
}

type iterator struct {
	*rbTree
	current *rbNode
	index   int
}

func (t *rbTree) Iter() *iterator {
	node := t.min(t.root)
	if node == t.NIL {
		return nil
	}
	it := &iterator{
		rbTree:  t,
		current: node,
		index:   int(t.size),
	}
	return it
}

func (it *iterator) First() *Entry {
	node := it.min(it.root)
	if node == it.NIL {
		return nil
	}
	if it.current != nil && it.current == node {
		return it.current.entry
	}
	it.current = node
	it.index = int(it.size)
	return it.current.entry
}

func (it *iterator) Last() *Entry {
	node := it.max(it.root)
	if node == it.NIL {
		return nil
	}
	if it.current != nil && it.current == node {
		return it.current.entry
	}
	it.current = node
	it.index = int(it.size)
	return it.current.entry
}

func (it *iterator) Next() *Entry {
	next := it.successor(it.current)
	if next == it.NIL {
		return nil
	}
	it.index--
	it.current = next
	return next.entry
}

func (it *iterator) Prev() *Entry {
	prev := it.predecessor(it.current)
	if prev == it.NIL {
		return nil
	}
	it.index--
	it.current = prev
	return prev.entry
}

func (it *iterator) HasMore() bool {
	return it.index > 1
}

type RangeFn func(key string, val []byte) bool

func (t *rbTree) Scan(iter RangeFn) {
	t.ascend(t.root, t.min(t.root).entry, iter)
}

func (t *rbTree) ScanBack(iter RangeFn) {
	t.descend(t.root, t.max(t.root).entry, iter)
}

func (t *rbTree) ScanRange(start, end *Entry, iter RangeFn) {
	t.ascendRange(t.root, start, end, iter)
}

func (t *rbTree) String() string {
	var sb strings.Builder
	t.ascend(
		t.root, t.min(t.root).entry, func(key string, val []byte) bool {
			entry := &Entry{key, val}
			sb.WriteString(entry.String())
			return true
		},
	)
	return sb.String()
}

func (t *rbTree) Close() {
	t.NIL = nil
	t.root = nil
	t.count = 0
	return
}

func (t *rbTree) Reset() {
	t.NIL = nil
	t.root = nil
	t.count = 0
	runtime.GC()
	n := &rbNode{
		left:   nil,
		right:  nil,
		parent: nil,
		color:  BLACK,
		entry:  empty,
	}
	t.NIL = n
	t.root = n
	t.count = 0
	t.size = 0
}

func (t *rbTree) insert(z *rbNode) (*rbNode, bool) {
	x := t.root
	y := t.NIL
	for x != t.NIL {
		y = x
		if compare(z.entry, x.entry) == -1 {
			x = (*rbNode)(x.getLeftPtr())
		} else if compare(x.entry, z.entry) == -1 {
			x = (*rbNode)(x.getRightPtr())
		} else {
			t.size -= int64(x.entry.Size())
			t.size += int64(z.entry.Size())
			// originally we were just returning x
			// without updating the Entry, but if we
			// want it to have similar behavior to
			// a hashmap then we need to update any
			// entries that already exist in the tree
			x.entry = z.entry
			return x, true // true means an existing
			// value was found and updated. It should
			// be noted that we don't need to re-balance
			// the tree because they keys are not changing
			// and the tree is balance is maintained by
			// the keys and not their values.
		}
	}
	z.assnParentPtr(unsafe.Pointer(y))

	if y == t.NIL {
		t.root = z
	} else if compare(z.entry, y.entry) == -1 {
		y.assnLeftPtr(unsafe.Pointer(z))
	} else {
		y.assnRightPtr(unsafe.Pointer(z))
	}
	t.count++
	t.size += int64(z.entry.Size())
	t.insertFixup(z)
	return z, false
}

func (t *rbTree) leftRotate(x *rbNode) {
	if x.right == unsafe.Pointer(t.NIL) {
		return
	}
	y := (*rbNode)(x.getRightPtr())
	x.assnRightPtr(y.getLeftPtr())
	if y.getLeftPtr() != unsafe.Pointer(t.NIL) {
		(*rbNode)(y.getLeftPtr()).assnParentPtr(unsafe.Pointer(x))
	}
	y.assnParentPtr(x.getParent())
	if x.getParent() == unsafe.Pointer(t.NIL) {
		t.root = y
	} else if x == x.getLeftUncle() {
		x.getParent().assnLeftPtr(y)
	} else {
		x.getParent().assnRightPtr(y)
	}
	y.assnLeftPtr(x)
	x.assnParentPtr(y)
}

func (t *rbTree) rightRotate(x *rbNode) {
	if x.getLeftPtr() == t.NIL {
		return
	}
	y := x.getLeftPtr()
	x.assnLeftPtr(y.getRightPtr())
	if y.getRightPtr() != t.NIL {
		y.getRightPtr().assnParentPtr(x)
	}
	y.assnParentPtr(x.getParent())

	if x.getParent() == t.NIL {
		t.root = y
	} else if x == x.getLeftUncle() {
		x.getParent().assnLeftPtr(y)
	} else {
		x.getParent().assnRightPtr(y)
	}
	y.assnRightPtr(x)
	x.assnParentPtr(y)
}

func (t *rbTree) insertFixup(z *rbNode) {
	for z.parent.color == RED {
		if z.getParent() == z.getGrandparent().getLeftPtr() {
			y := z.getGrandparent().getRightPtr()
			if y.color == RED {
				z.getParent().color = BLACK
				y.color = BLACK
				z.getGrandparent().color = RED
				z = z.getGrandparent()
			} else {
				if z == z.getRightUncle() {
					z = z.getParent()
					t.leftRotate(z)
				}
				z.getParent().color = BLACK
				z.getGrandparent().color = RED
				t.rightRotate(z.getGrandparent())
			}
		} else {
			y := z.getGrandparent().getLeftPtr()
			if y.color == RED {
				z.getParent().color = BLACK
				y.color = BLACK
				z.getGrandparent().color = RED
				z = z.getGrandparent()
			} else {
				if z == z.getLeftUncle() {
					z = z.getParent()
					t.rightRotate(z)
				}
				z.getParent().color = BLACK
				z.getGrandparent().color = RED
				t.leftRotate(z.getGrandparent())
			}
		}
	}
	t.root.color = BLACK
}

// trying out a slightly different search method
// that (hopefully) will not return nil values and
// instead will return approximate node matches
func (t *rbTree) searchApprox(x *rbNode) *rbNode {
	p := t.root
	for p != t.NIL {
		if compare(p.entry, x.entry) == -1 {
			if p.right == unsafe.Pointer(t.NIL) {
				break
			}
			p = p.right
		} else if compare(x.entry, p.entry) == -1 {
			if p.left == unsafe.Pointer(t.NIL) {
				break
			}
			p = p.left
		} else {
			break
		}
	}
	return p
}

func (t *rbTree) search(x *rbNode) *rbNode {
	p := t.root
	for p != t.NIL {
		if compare(p.entry, x.entry) == -1 {
			p.assnRightPtr(p.right)
		} else if compare(x.entry, p.entry) == -1 {
			p.assnLeftPtr(p.left)
		} else {
			break
		}
	}
	return p
}

// min traverses from root to left recursively until left is NIL
func (t *rbTree) min(x *rbNode) *rbNode {
	if x == t.NIL {
		return t.NIL
	}
	for x.left != t.NIL {
		x = x.left
	}
	return x
}

// max traverses from root to right recursively until right is NIL
func (t *rbTree) max(x *rbNode) *rbNode {
	if x == t.NIL {
		return t.NIL
	}
	for x.right != t.NIL {
		x = x.right
	}
	return x
}

func (t *rbTree) predecessor(x *rbNode) *rbNode {
	if x == t.NIL {
		return t.NIL
	}
	if x.left != t.NIL {
		return t.max(x.left)
	}
	y := x.parent
	for y != t.NIL && x == y.left {
		x = y
		y = y.parent
	}
	return y
}

func (t *rbTree) successor(x *rbNode) *rbNode {
	if x == t.NIL {
		return t.NIL
	}
	if x.right != t.NIL {
		return t.min(x.right)
	}
	y := x.parent
	for y != t.NIL && x == y.right {
		x = y
		y = y.parent
	}
	return y
}

func (t *rbTree) delete(key *rbNode) *rbNode {
	z := t.search(key)
	if z == t.NIL {
		return t.NIL
	}
	ret := &rbNode{t.NIL, t.NIL, t.NIL, z.color, z.entry}
	var y *rbNode
	var x *rbNode
	if z.left == t.NIL || z.right == t.NIL {
		y = z
	} else {
		y = t.successor(z)
	}
	if y.left != t.NIL {
		x = y.left
	} else {
		x = y.right
	}
	x.parent = y.parent

	if y.parent == t.NIL {
		t.root = x
	} else if y == y.parent.left {
		y.parent.left = x
	} else {
		y.parent.right = x
	}
	if y != z {
		z.entry = y.entry
	}
	if y.color == BLACK {
		t.deleteFixup(x)
	}
	t.size -= int64(ret.entry.Size())
	t.count--
	return ret
}

func (t *rbTree) deleteFixup(x *rbNode) {
	for x != t.root && x.color == BLACK {
		if x == x.parent.left {
			w := x.parent.right
			if w.color == RED {
				w.color = BLACK
				x.parent.color = RED
				t.leftRotate(x.parent)
				w = x.parent.right
			}
			if w.left.color == BLACK && w.right.color == BLACK {
				w.color = RED
				x = x.parent
			} else {
				if w.right.color == BLACK {
					w.left.color = BLACK
					w.color = RED
					t.rightRotate(w)
					w = x.parent.right
				}
				w.color = x.parent.color
				x.parent.color = BLACK
				w.right.color = BLACK
				t.leftRotate(x.parent)
				// this is to exit while loop
				x = t.root
			}
		} else {
			w := x.parent.left
			if w.color == RED {
				w.color = BLACK
				x.parent.color = RED
				t.rightRotate(x.parent)
				w = x.parent.left
			}
			if w.left.color == BLACK && w.right.color == BLACK {
				w.color = RED
				x = x.parent
			} else {
				if w.left.color == BLACK {
					w.right.color = BLACK
					w.color = RED
					t.leftRotate(w)
					w = x.parent.left
				}
				w.color = x.parent.color
				x.parent.color = BLACK
				w.left.color = BLACK
				t.rightRotate(x.parent)
				x = t.root
			}
		}
	}
	x.color = BLACK
}

func (t *rbTree) ascend(x *rbNode, entry *Entry, iter RangeFn) bool {
	if x == t.NIL {
		return true
	}
	if !(compare(x.entry, entry) == -1) {
		if !t.ascend(x.left, entry, iter) {
			return false
		}
		if !iter(x.entry.Key, x.entry.Val) {
			return false
		}
	}
	return t.ascend(x.right, entry, iter)
}

func (t *rbTree) descend(x *rbNode, pivot *Entry, iter RangeFn) bool {
	if x == t.NIL {
		return true
	}
	if !(compare(pivot, x.entry) == -1) {
		if !t.descend(x.right, pivot, iter) {
			return false
		}
		if !iter(x.entry.Key, x.entry.Val) {
			return false
		}
	}
	return t.descend(x.left, pivot, iter)
}

func (t *rbTree) ascendRange(x *rbNode, inf, sup *Entry, iter RangeFn) bool {
	if x == t.NIL {
		return true
	}
	if !(compare(x.entry, sup) == -1) {
		return t.ascendRange(x.left, inf, sup, iter)
	}
	if compare(x.entry, inf) == -1 {
		return t.ascendRange(x.right, inf, sup, iter)
	}
	if !t.ascendRange(x.left, inf, sup, iter) {
		return false
	}
	if !iter(x.entry.Key, x.entry.Val) {
		return false
	}
	return t.ascendRange(x.right, inf, sup, iter)
}
