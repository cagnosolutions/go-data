package generic

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"unsafe"
)

type Entry interface {
	// this < that  : -1
	// this == that :  0
	// this > that  : +1
	Compare(that Entry) int
	String() string
}

const (
	RED   = 0
	BLACK = 1
)

func compare(x, y Entry) int {
	return x.Compare(y)
}

type rbNode[T Entry] struct {
	left   *rbNode[T]
	right  *rbNode[T]
	parent *rbNode[T]
	color  uint
	entry  T
}

// RBTree is a struct representing a RBTree
type RBTree[T Entry] struct {
	lock  sync.RWMutex
	NIL   *rbNode[T]
	root  *rbNode[T]
	count int
	size  int64
	empty T
}

// NewTree creates and returns a new RBTree
func NewTree[T Entry]() *RBTree[T] {
	var empty T
	n := &rbNode[T]{
		left:   nil,
		right:  nil,
		parent: nil,
		color:  BLACK,
		entry:  empty,
	}
	return &RBTree[T]{
		NIL:   n,
		root:  n,
		count: 0,
	}
}

func (t *RBTree[T]) GetClone() *RBTree[T] {
	t.Lock()
	defer t.Unlock()
	clone := NewTree[T]()
	t.cloneEntries(clone)
	return clone
}

func (t *RBTree[T]) Lock() {
	t.lock.Lock()
}

func (t *RBTree[T]) Unlock() {
	t.lock.Unlock()
}

func (t *RBTree[T]) RLock() {
	t.lock.RLock()
}

func (t *RBTree[T]) RUnlock() {
	t.lock.RUnlock()
}

// Has tests and returns a boolean value if the
// provided key exists in the tree
func (t *RBTree[T]) Has(entry T) bool {
	_, ok := t.getInternal(entry)
	return ok
}

// Add adds the provided key and value only if it does not
// already exist in the tree. It returns false if the key and
// value was not able to be added, and true if it was added
// successfully
func (t *RBTree[T]) Add(entry T) bool {
	_, ok := t.getInternal(entry)
	if ok {
		// key already exists, so we are not adding
		return false
	}
	t.putInternal(entry)
	return true
}

func (t *RBTree[T]) Put(entry T) (T, bool) {
	return t.putInternal(entry)
}

func (t *RBTree[T]) putInternal(entry T) (val T, ok bool) {
	if entry.Compare(t.empty) == 0 {
		return val, false
	}
	// insert returns the node inserted
	// and if the node returned already
	// existed and/or was updated
	ret, ok := t.insert(
		&rbNode[T]{
			left:   t.NIL,
			right:  t.NIL,
			parent: t.NIL,
			color:  RED,
			entry:  entry,
		},
	)
	val = ret.entry
	return val, ok
}

func (t *RBTree[T]) Get(entry T) (T, bool) {
	return t.getInternal(entry)
}

// GetNearMin performs an approximate search for the specified key
// and returns the closest key that is less than (the predecessor)
// to the searched key as well as a boolean reporting true if an
// exact match was found for the key, and false if it is unknown
// or and exact match was not found
func (t *RBTree[T]) GetNearMin(entry T) (val T, exactMatch bool) {
	if entry.Compare(t.empty) == 0 {
		return val, false
	}
	ret := t.searchApprox(
		&rbNode[T]{
			left:   t.NIL,
			right:  t.NIL,
			parent: t.NIL,
			color:  RED,
			entry:  entry,
		},
	)
	prev := t.predecessor(ret).entry
	if entry.Compare(t.empty) == 0 {
		prev, _ = t.Min()
	}
	return prev, compare(ret.entry, entry) == 0
}

// GetNearMax performs an approximate search for the specified key
// and returns the closest key that is greater than (the successor)
// to the searched key as well as a boolean reporting true if an
// exact match was found for the key, and false if it is unknown or
// and exact match was not found
func (t *RBTree[T]) GetNearMax(entry T) (val T, exactMatch bool) {
	if entry.Compare(t.empty) == 0 {
		return val, false
	}
	ret := t.searchApprox(
		&rbNode[T]{
			left:   t.NIL,
			right:  t.NIL,
			parent: t.NIL,
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
func (t *RBTree[T]) GetApproxPrevNext(entry T) (key T, prev T, next T, exactMatch bool) {
	if entry.Compare(t.empty) == 0 {
		return key, prev, next, false
	}
	ret := t.searchApprox(
		&rbNode[T]{
			left:   t.NIL,
			right:  t.NIL,
			parent: t.NIL,
			color:  RED,
			entry:  entry,
		},
	)
	return ret.entry, t.predecessor(ret).entry, t.successor(ret).entry, compare(ret.entry, entry) == 0
}

func (t *RBTree[T]) getInternal(entry T) (val T, found bool) {
	if entry.Compare(t.empty) == 0 {
		return val, false
	}
	ret := t.search(
		&rbNode[T]{
			left:   t.NIL,
			right:  t.NIL,
			parent: t.NIL,
			color:  RED,
			entry:  entry,
		},
	)
	return ret.entry, ret.entry.Compare(t.empty) != 0
}

func (t *RBTree[T]) Del(entry T) (T, bool) {
	return t.delInternal(entry)
}

func (t *RBTree[T]) delInternal(entry T) (val T, found bool) {
	if entry.Compare(t.empty) == 0 {
		return val, false
	}
	cnt := t.count
	ret := t.delete(
		&rbNode[T]{
			left:   t.NIL,
			right:  t.NIL,
			parent: t.NIL,
			color:  RED,
			entry:  entry,
		},
	)
	return ret.entry, cnt == t.count+1
}

func (t *RBTree[T]) Len() int {
	return t.count
}

// Size returns the size in bytes
func (t *RBTree[T]) Size() int64 {
	return t.size
}

func (t *RBTree[T]) Min() (entry T, found bool) {
	x := t.min(t.root)
	if x == t.NIL {
		return entry, false
	}
	return x.entry, true
}

func (t *RBTree[T]) Max() (entry T, found bool) {
	x := t.max(t.root)
	if x == t.NIL {
		return entry, false
	}
	return x.entry, true
}

// helper function for clone
func (t *RBTree[T]) cloneEntries(t2 *RBTree[T]) {
	t.ascend(
		t.root, t.min(t.root).entry, func(e T) bool {
			t2.putInternal(e)
			return true
		},
	)
}

type iterator[T Entry] struct {
	*RBTree[T]
	current *rbNode[T]
	index   int
}

func (t *RBTree[T]) Iter() *iterator[T] {
	node := t.min(t.root)
	if node == t.NIL {
		return nil
	}
	it := &iterator[T]{
		RBTree:  t,
		current: node,
		index:   int(t.size),
	}
	return it
}

func (it *iterator[T]) First() (entry T) {
	node := it.min(it.root)
	if node == it.NIL {
		return entry
	}
	if it.current != nil && it.current == node {
		return it.current.entry
	}
	it.current = node
	it.index = int(it.size)
	return it.current.entry
}

func (it *iterator[T]) Last() (entry T) {
	node := it.max(it.root)
	if node == it.NIL {
		return entry
	}
	if it.current != nil && it.current == node {
		return it.current.entry
	}
	it.current = node
	it.index = int(it.size)
	return it.current.entry
}

func (it *iterator[T]) Next() (entry T) {
	next := it.successor(it.current)
	if next == it.NIL {
		return entry
	}
	it.index--
	it.current = next
	entry = next.entry
	return entry
}

func (it *iterator[T]) Prev() (entry T) {
	prev := it.predecessor(it.current)
	if prev == it.NIL {
		return entry
	}
	it.index--
	it.current = prev
	entry = prev.entry
	return entry
}

func (it *iterator[T]) HasMore() bool {
	return it.index > 1
}

type RangeFn[T Entry] func(entry T) bool

func (t *RBTree[T]) Scan(iter RangeFn[T]) {
	t.ascend(t.root, t.min(t.root).entry, iter)
}

func (t *RBTree[T]) ScanBack(iter RangeFn[T]) {
	t.descend(t.root, t.max(t.root).entry, iter)
}

func (t *RBTree[T]) ScanRange(start, end T, iter RangeFn[T]) {
	t.ascendRange(t.root, start, end, iter)
}

func (t *RBTree[T]) String() string {
	var sb strings.Builder
	t.ascend(
		t.root, t.min(t.root).entry, func(entry T) bool {
			fmt.Fprintf(&sb, "%v", entry)
			return true
		},
	)
	return sb.String()
}

func (t *RBTree[T]) Close() {
	t.NIL = nil
	t.root = nil
	t.count = 0
	return
}

func (t *RBTree[T]) Reset() {
	t.NIL = nil
	t.root = nil
	t.count = 0
	runtime.GC()
	var empty T
	n := &rbNode[T]{
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

func (t *RBTree[T]) insert(z *rbNode[T]) (*rbNode[T], bool) {
	x := t.root
	y := t.NIL
	for x != t.NIL {
		y = x
		if compare(z.entry, x.entry) == -1 {
			x = x.left
		} else if compare(x.entry, z.entry) == -1 {
			x = x.right
		} else {
			t.size -= int64(unsafe.Sizeof(x.entry))
			t.size += int64(unsafe.Sizeof(z.entry))
			// originally we were just returning x
			// without updating the RBEntry, but if we
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
	z.parent = y
	if y == t.NIL {
		t.root = z
	} else if compare(z.entry, y.entry) == -1 {
		y.left = z
	} else {
		y.right = z
	}
	t.count++
	t.size += int64(unsafe.Sizeof(z.entry))
	t.insertFixup(z)
	return z, false
}

func (t *RBTree[T]) leftRotate(x *rbNode[T]) {
	if x.right == t.NIL {
		return
	}
	y := x.right
	x.right = y.left
	if y.left != t.NIL {
		y.left.parent = x
	}
	y.parent = x.parent
	if x.parent == t.NIL {
		t.root = y
	} else if x == x.parent.left {
		x.parent.left = y
	} else {
		x.parent.right = y
	}
	y.left = x
	x.parent = y
}

func (t *RBTree[T]) rightRotate(x *rbNode[T]) {
	if x.left == t.NIL {
		return
	}
	y := x.left
	x.left = y.right
	if y.right != t.NIL {
		y.right.parent = x
	}
	y.parent = x.parent

	if x.parent == t.NIL {
		t.root = y
	} else if x == x.parent.left {
		x.parent.left = y
	} else {
		x.parent.right = y
	}

	y.right = x
	x.parent = y
}

func (t *RBTree[T]) insertFixup(z *rbNode[T]) {
	for z.parent.color == RED {
		if z.parent == z.parent.parent.left {
			y := z.parent.parent.right
			if y.color == RED {
				z.parent.color = BLACK
				y.color = BLACK
				z.parent.parent.color = RED
				z = z.parent.parent
			} else {
				if z == z.parent.right {
					z = z.parent
					t.leftRotate(z)
				}
				z.parent.color = BLACK
				z.parent.parent.color = RED
				t.rightRotate(z.parent.parent)
			}
		} else {
			y := z.parent.parent.left
			if y.color == RED {
				z.parent.color = BLACK
				y.color = BLACK
				z.parent.parent.color = RED
				z = z.parent.parent
			} else {
				if z == z.parent.left {
					z = z.parent
					t.rightRotate(z)
				}
				z.parent.color = BLACK
				z.parent.parent.color = RED
				t.leftRotate(z.parent.parent)
			}
		}
	}
	t.root.color = BLACK
}

// trying out a slightly different search method
// that (hopefully) will not return nil values and
// instead will return approximate node matches
func (t *RBTree[T]) searchApprox(x *rbNode[T]) *rbNode[T] {
	p := t.root
	for p != t.NIL {
		if compare(p.entry, x.entry) == -1 {
			if p.right == t.NIL {
				break
			}
			p = p.right
		} else if compare(x.entry, p.entry) == -1 {
			if p.left == t.NIL {
				break
			}
			p = p.left
		} else {
			break
		}
	}
	return p
}

func (t *RBTree[T]) search(x *rbNode[T]) *rbNode[T] {
	p := t.root
	for p != t.NIL {
		if compare(p.entry, x.entry) == -1 {
			p = p.right
		} else if compare(x.entry, p.entry) == -1 {
			p = p.left
		} else {
			break
		}
	}
	return p
}

// min traverses from root to left recursively until left is NIL
func (t *RBTree[T]) min(x *rbNode[T]) *rbNode[T] {
	if x == t.NIL {
		return t.NIL
	}
	for x.left != t.NIL {
		x = x.left
	}
	return x
}

// max traverses from root to right recursively until right is NIL
func (t *RBTree[T]) max(x *rbNode[T]) *rbNode[T] {
	if x == t.NIL {
		return t.NIL
	}
	for x.right != t.NIL {
		x = x.right
	}
	return x
}

func (t *RBTree[T]) predecessor(x *rbNode[T]) *rbNode[T] {
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

func (t *RBTree[T]) successor(x *rbNode[T]) *rbNode[T] {
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

func (t *RBTree[T]) delete(key *rbNode[T]) *rbNode[T] {
	z := t.search(key)
	if z == t.NIL {
		return t.NIL
	}
	ret := &rbNode[T]{t.NIL, t.NIL, t.NIL, z.color, z.entry}
	var y *rbNode[T]
	var x *rbNode[T]
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
	t.size -= int64(unsafe.Sizeof(ret.entry))
	t.count--
	return ret
}

func (t *RBTree[T]) deleteFixup(x *rbNode[T]) {
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

func (t *RBTree[T]) ascend(x *rbNode[T], entry T, iter RangeFn[T]) bool {
	if x == t.NIL {
		return true
	}
	if !(compare(x.entry, entry) == -1) {
		if !t.ascend(x.left, entry, iter) {
			return false
		}
		if !iter(x.entry) {
			return false
		}
	}
	return t.ascend(x.right, entry, iter)
}

func (t *RBTree[T]) __Descend(pivot T, iter RangeFn[T]) {
	t.descend(t.root, pivot, iter)
}

func (t *RBTree[T]) descend(x *rbNode[T], pivot T, iter RangeFn[T]) bool {
	if x == t.NIL {
		return true
	}
	if !(compare(pivot, x.entry) == -1) {
		if !t.descend(x.right, pivot, iter) {
			return false
		}
		if !iter(x.entry) {
			return false
		}
	}
	return t.descend(x.left, pivot, iter)
}

func (t *RBTree[T]) ascendRange(x *rbNode[T], inf, sup T, iter RangeFn[T]) bool {
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
	if !iter(x.entry) {
		return false
	}
	return t.ascendRange(x.right, inf, sup, iter)
}
