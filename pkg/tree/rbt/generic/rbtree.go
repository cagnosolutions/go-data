package generic

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"unsafe"
)

const (
	RED   = 0
	BLACK = 1
)

type Entry[K Ordered, V any] struct {
	Key K
	Val V
}

func compare[K Ordered](x, y K) int {
	if x > y {
		return +1
	}
	if x < y {
		return -1
	}
	return 0
}

type rbNode[K Ordered, V any] struct {
	left   *rbNode[K, V]
	right  *rbNode[K, V]
	parent *rbNode[K, V]
	color  uint
	entry  *Entry[K, V]
}

// RBTree is a struct representing a RBTree
type RBTree[K Ordered, V any] struct {
	lock  sync.RWMutex
	NIL   *rbNode[K, V]
	root  *rbNode[K, V]
	count int
	size  int64
	empty Entry[K, V]
	pool  *sync.Pool
}

func (t *RBTree[K, V]) initPool(key K, val V) {
	if t.pool == nil {
		t.pool = &sync.Pool{
			New: func() any {
				return &rbNode[K, V]{
					left:   t.NIL,
					right:  t.NIL,
					parent: t.NIL,
					color:  RED,
					entry: &Entry[K, V]{
						Key: key,
						Val: val,
					},
				}
			},
		}
	}
}

func (t *RBTree[K, V]) getSearchNode(key K) *rbNode[K, V] {
	n := t.pool.Get().(*rbNode[K, V])
	n.entry.Key = key
	return n
}

func (t *RBTree[K, V]) putSearchNode(n *rbNode[K, V]) {
	n.entry.Key = t.empty.Key
	n.entry.Val = t.empty.Val
	t.pool.Put(n)
}

// NewTree creates and returns a new RBTree
func NewTree[K Ordered, V any]() *RBTree[K, V] {
	empty := new(Entry[K, V])
	n := &rbNode[K, V]{
		left:   nil,
		right:  nil,
		parent: nil,
		color:  BLACK,
		entry:  empty,
	}
	t := &RBTree[K, V]{
		NIL:   n,
		root:  n,
		count: 0,
	}
	t.initPool(empty.Key, empty.Val)
	return t
}

func (t *RBTree[K, V]) GetClone() *RBTree[K, V] {
	t.Lock()
	defer t.Unlock()
	clone := NewTree[K, V]()
	t.cloneEntries(clone)
	return clone
}

func (t *RBTree[K, V]) Lock() {
	t.lock.Lock()
}

func (t *RBTree[K, V]) Unlock() {
	t.lock.Unlock()
}

func (t *RBTree[K, V]) RLock() {
	t.lock.RLock()
}

func (t *RBTree[K, V]) RUnlock() {
	t.lock.RUnlock()
}

// Has tests and returns a boolean value if the
// provided key exists in the tree
func (t *RBTree[K, V]) Has(key K) bool {
	_, ok := t.getInternal(key)
	return ok
}

// Add adds the provided key and value only if it does not
// already exist in the tree. It returns false if the key and
// value was not able to be added, and true if it was added
// successfully
func (t *RBTree[K, V]) Add(key K, val V) bool {
	_, ok := t.getInternal(key)
	if ok {
		// key already exists, so we are not adding
		return false
	}
	t.putInternal(key, val)
	return true
}

func (t *RBTree[K, V]) Put(key K, val V) (V, bool) {
	return t.putInternal(key, val)
}

func (t *RBTree[K, V]) putInternal(key K, val V) (V, bool) {
	if compare(key, t.empty.Key) == 0 {
		return val, false
	}
	// insert returns the node inserted
	// and if the node returned already
	// existed and/or was updated
	ret, ok := t.insert(
		&rbNode[K, V]{
			left:   t.NIL,
			right:  t.NIL,
			parent: t.NIL,
			color:  RED,
			entry: &Entry[K, V]{
				Key: key,
				Val: val,
			},
		},
	)
	return ret.entry.Val, ok
}

func (t *RBTree[K, V]) Get(key K) (V, bool) {
	return t.getInternal(key)
}

// GetNearMin performs an approximate search for the specified key
// and returns the closest key that is less than (the predecessor)
// to the searched key as well as a boolean reporting true if an
// exact match was found for the key, and false if it is unknown
// or and exact match was not found
func (t *RBTree[K, V]) GetNearMin(key K) (val V, exactMatch bool) {
	if compare(key, t.empty.Key) == 0 {
		return val, false
	}
	ret := t.searchApprox(t.getSearchNode(key))
	prev := t.predecessor(ret).entry.Val
	if compare(key, t.empty.Key) == 0 {
		prev, _ = t.Min()
	}
	return prev, compare(ret.entry.Key, key) == 0
}

// GetNearMax performs an approximate search for the specified key
// and returns the closest key that is greater than (the successor)
// to the searched key as well as a boolean reporting true if an
// exact match was found for the key, and false if it is unknown or
// and exact match was not found
func (t *RBTree[K, V]) GetNearMax(key K) (val V, exactMatch bool) {
	if compare(key, t.empty.Key) == 0 {
		return val, false
	}
	ret := t.searchApprox(t.getSearchNode(key))
	return t.successor(ret).entry.Val, compare(ret.entry.Key, key) == 0
}

// GetApproxPrevNext performs an approximate search for the specified key
// and returns the searched key, the predecessor, and the successor and a
// boolean reporting true if an exact match was found for the key, and false
// if it is unknown or and exact match was not found
func (t *RBTree[K, V]) GetApproxPrevNext(key K) (matched V, prev V, next V, exactMatch bool) {
	if compare(key, t.empty.Key) == 0 {
		return matched, prev, next, false
	}
	ret := t.searchApprox(t.getSearchNode(key))
	return ret.entry.Val,
		t.predecessor(ret).entry.Val,
		t.successor(ret).entry.Val,
		compare(ret.entry.Key, key) == 0
}

func (t *RBTree[K, V]) getInternal(key K) (val V, found bool) {
	if compare(key, t.empty.Key) == 0 {
		return val, false
	}
	sn := t.getSearchNode(key)
	ret := t.search(sn)
	t.putSearchNode(sn)
	return ret.entry.Val, compare(key, t.empty.Key) != 0 && key == ret.entry.Key
}

func (t *RBTree[K, V]) Del(key K) (V, bool) {
	return t.delInternal(key)
}

func (t *RBTree[K, V]) delInternal(key K) (val V, found bool) {
	if compare(key, t.empty.Key) == 0 {
		return val, false
	}
	cnt := t.count
	sn := t.getSearchNode(key)
	ret := t.delete(t.getSearchNode(key))
	t.putSearchNode(sn)
	return ret.entry.Val, cnt == t.count+1
}

func (t *RBTree[K, V]) Len() int {
	return t.count
}

// Size returns the size in bytes
func (t *RBTree[K, V]) Size() int64 {
	return t.size
}

func (t *RBTree[K, V]) Min() (val V, found bool) {
	x := t.min(t.root)
	if x == t.NIL {
		return val, false
	}
	return x.entry.Val, true
}

func (t *RBTree[K, V]) Max() (val V, found bool) {
	x := t.max(t.root)
	if x == t.NIL {
		return val, false
	}
	return x.entry.Val, true
}

// helper function for clone
func (t *RBTree[K, V]) cloneEntries(t2 *RBTree[K, V]) {
	t.ascend(
		t.root, t.min(t.root).entry.Key, func(key K, val V) bool {
			t2.putInternal(key, val)
			return true
		},
	)
}

type iterator[K Ordered, V any] struct {
	*RBTree[K, V]
	current *rbNode[K, V]
	index   int
}

func (t *RBTree[K, V]) Iter() *iterator[K, V] {
	node := t.min(t.root)
	if node == t.NIL {
		return nil
	}
	it := &iterator[K, V]{
		RBTree:  t,
		current: node,
		index:   int(t.size),
	}
	return it
}

func (it *iterator[K, V]) First() (val V) {
	node := it.min(it.root)
	if node == it.NIL {
		return val
	}
	if it.current != nil && it.current == node {
		return it.current.entry.Val
	}
	it.current = node
	it.index = int(it.size)
	return it.current.entry.Val
}

func (it *iterator[K, V]) Last() (val V) {
	node := it.max(it.root)
	if node == it.NIL {
		return val
	}
	if it.current != nil && it.current == node {
		return it.current.entry.Val
	}
	it.current = node
	it.index = int(it.size)
	return it.current.entry.Val
}

func (it *iterator[K, V]) Next() (val V) {
	next := it.successor(it.current)
	if next == it.NIL {
		return val
	}
	it.index--
	it.current = next
	val = next.entry.Val
	return val
}

func (it *iterator[K, V]) Prev() (val V) {
	prev := it.predecessor(it.current)
	if prev == it.NIL {
		return val
	}
	it.index--
	it.current = prev
	val = prev.entry.Val
	return val
}

func (it *iterator[K, V]) HasMore() bool {
	return it.index > 1
}

type RangeFn[K Ordered, V any] func(key K, val V) bool

func (t *RBTree[K, V]) Scan(iter RangeFn[K, V]) {
	t.ascend(t.root, t.min(t.root).entry.Key, iter)
}

func (t *RBTree[K, V]) ScanBack(iter RangeFn[K, V]) {
	t.descend(t.root, t.max(t.root).entry.Key, iter)
}

func (t *RBTree[K, V]) ScanRange(start, end K, iter RangeFn[K, V]) {
	t.ascendRange(t.root, start, end, iter)
}

func (t *RBTree[K, V]) String() string {
	var sb strings.Builder
	t.ascend(
		t.root, t.min(t.root).entry.Key, func(key K, val V) bool {
			fmt.Fprintf(&sb, "key=%v, val=%v", key, val)
			return true
		},
	)
	return sb.String()
}

func (t *RBTree[K, V]) Close() {
	t.NIL = nil
	t.root = nil
	t.count = 0
	return
}

func (t *RBTree[K, V]) Reset() {
	t.NIL = nil
	t.root = nil
	t.count = 0
	runtime.GC()
	empty := new(Entry[K, V])
	n := &rbNode[K, V]{
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

func (t *RBTree[K, V]) insert(z *rbNode[K, V]) (*rbNode[K, V], bool) {
	x := t.root
	y := t.NIL
	for x != t.NIL {
		y = x
		if compare(z.entry.Key, x.entry.Key) == -1 {
			x = x.left
		} else if compare(x.entry.Key, z.entry.Key) == -1 {
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
	} else if compare(z.entry.Key, y.entry.Key) == -1 {
		y.left = z
	} else {
		y.right = z
	}
	t.count++
	t.size += int64(unsafe.Sizeof(z.entry))
	t.insertFixup(z)
	return z, false
}

func (t *RBTree[K, V]) leftRotate(x *rbNode[K, V]) {
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

func (t *RBTree[K, V]) rightRotate(x *rbNode[K, V]) {
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

func (t *RBTree[K, V]) insertFixup(z *rbNode[K, V]) {
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
func (t *RBTree[K, V]) searchApprox(x *rbNode[K, V]) *rbNode[K, V] {
	p := t.root
	for p != t.NIL {
		if compare(p.entry.Key, x.entry.Key) == -1 {
			if p.right == t.NIL {
				break
			}
			p = p.right
		} else if compare(x.entry.Key, p.entry.Key) == -1 {
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

func (t *RBTree[K, V]) search(x *rbNode[K, V]) *rbNode[K, V] {
	p := t.root
	for p != t.NIL {
		if compare(p.entry.Key, x.entry.Key) == -1 {
			p = p.right
		} else if compare(x.entry.Key, p.entry.Key) == -1 {
			p = p.left
		} else {
			break
		}
	}
	return p
}

// min traverses from root to left recursively until left is NIL
func (t *RBTree[K, V]) min(x *rbNode[K, V]) *rbNode[K, V] {
	if x == t.NIL {
		return t.NIL
	}
	for x.left != t.NIL {
		x = x.left
	}
	return x
}

// max traverses from root to right recursively until right is NIL
func (t *RBTree[K, V]) max(x *rbNode[K, V]) *rbNode[K, V] {
	if x == t.NIL {
		return t.NIL
	}
	for x.right != t.NIL {
		x = x.right
	}
	return x
}

func (t *RBTree[K, V]) predecessor(x *rbNode[K, V]) *rbNode[K, V] {
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

func (t *RBTree[K, V]) successor(x *rbNode[K, V]) *rbNode[K, V] {
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

func (t *RBTree[K, V]) delete(key *rbNode[K, V]) *rbNode[K, V] {
	z := t.search(key)
	if z == t.NIL {
		return t.NIL
	}
	ret := &rbNode[K, V]{
		t.NIL,
		t.NIL,
		t.NIL,
		z.color,
		z.entry,
	}
	var y *rbNode[K, V]
	var x *rbNode[K, V]
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

func (t *RBTree[K, V]) deleteFixup(x *rbNode[K, V]) {
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

func (t *RBTree[K, V]) ascend(x *rbNode[K, V], key K, iter RangeFn[K, V]) bool {
	if x == t.NIL {
		return true
	}
	if !(compare(x.entry.Key, key) == -1) {
		if !t.ascend(x.left, key, iter) {
			return false
		}
		if !iter(x.entry.Key, x.entry.Val) {
			return false
		}
	}
	return t.ascend(x.right, key, iter)
}

func (t *RBTree[K, V]) __Descend(pivot K, iter RangeFn[K, V]) {
	t.descend(t.root, pivot, iter)
}

func (t *RBTree[K, V]) descend(x *rbNode[K, V], pivot K, iter RangeFn[K, V]) bool {
	if x == t.NIL {
		return true
	}
	if !(compare(pivot, x.entry.Key) == -1) {
		if !t.descend(x.right, pivot, iter) {
			return false
		}
		if !iter(x.entry.Key, x.entry.Val) {
			return false
		}
	}
	return t.descend(x.left, pivot, iter)
}

func (t *RBTree[K, V]) ascendRange(x *rbNode[K, V], inf, sup K, iter RangeFn[K, V]) bool {
	if x == t.NIL {
		return true
	}
	if !(compare(x.entry.Key, sup) == -1) {
		return t.ascendRange(x.left, inf, sup, iter)
	}
	if compare(x.entry.Key, inf) == -1 {
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
