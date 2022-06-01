package pager

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

// clockReplacer represents a clock based replacement cache
type clockReplacer[K comparable, V any] struct {
	list *circularList[K, V]
	ptr  **node[K, V]
}

// newClockReplacer instantiates and returns a new clockReplacer
func newClockReplacer[K comparable, V any](size int) *clockReplacer[K, V] {
	list := newCircularList[K, V](size)
	return &clockReplacer[K, V]{
		list: list,
		ptr:  &list.head,
	}
}

// pin takes a frameID and pins a pageFrame indicating that it should not
// be victimized until it is unpinned
func (c *clockReplacer[K, V]) pin(k K) {
	n := c.list.find(k)
	if n == nil {
		return
	}
	if (*c.ptr) == n {
		c.ptr = &(*c.ptr).next
	}
	c.list.remove(k)
}

// unpin takes a frameID and unpins a pageFrame indicating that it is now
// available for victimization
func (c *clockReplacer[K, V]) unpin(k K, v V) {
	if !c.list.hasKey(k) {
		c.list.insert(k, v)
		if c.list.size == 1 {
			c.ptr = &c.list.head
		}
	}
}

// victim removes the victim pageFrame as defined by the replacement policy
// and returns the frameID of the victim
func (c *clockReplacer[K, V]) victim() (K, V) {
	if c.list.size == 0 {
		var zK K
		var zV V
		return zK, zV
	}
	var ck K
	var cv V
	cn := *c.ptr
	for {
		if !reflect.DeepEqual(cn.val, nil) {
			cn.val = *new(V)
			c.ptr = &cn.next
		} else {
			ck = cn.key
			cv = cn.val
			c.ptr = &cn.next
			c.list.remove(cn.key)
			return ck, cv
		}
	}
}

// size returns the size of the clockReplacer
func (c *clockReplacer[K, V]) size() int {
	return c.list.size
}

// ErrListIsFull errors reports when the circular list is at capacity
var ErrListIsFull = errors.New("list is full; circular list capacity met")

// node is a node in a circular list.
type node[K comparable, V any] struct {
	key        K
	val        V
	prev, next *node[K, V]
}

// String is the stringer method for a node
func (n *node[K, V]) String() string {
	return fmt.Sprintf("%v<-[%v]->%v", n.prev.key, n.key, n.next.key)
}

// circularList is a circular list implementation.
type circularList[K comparable, V any] struct {
	head, tail *node[K, V]
	size       int
	capacity   int
}

// newCircularList instantiates and returns a pointer to a new
// circular list instance with the capacity set using the provided
// max integer.
func newCircularList[K comparable, V any](max int) *circularList[K, V] {
	return &circularList[K, V]{
		head:     nil,
		tail:     nil,
		size:     0,
		capacity: max,
	}
}

// find takes a key and attempts to locate and return a node with
// the matching key. If said node cannot be found, find returns nil.
func (c *circularList[K, V]) find(k K) *node[K, V] {
	ptr := c.head
	for i := 0; i < c.size; i++ {
		if ptr.key == k {
			return ptr
		}
		ptr = ptr.next
	}
	return nil
}

// hasKey takes a key and returns a boolean indicating true if that
// key is found within the list, and false if it is not.
func (c *circularList[K, V]) hasKey(k K) bool {
	return c.find(k) != nil
}

// insert takes a key and value and inserts it into the list, unless
// the list is at its capacity.
func (c *circularList[K, V]) insert(k K, v V) error {
	// check capacity
	if c.size == c.capacity {
		return ErrListIsFull
	}
	// create new node to insert
	nn := &node[K, V]{
		key:  k,
		val:  v,
		prev: nil,
		next: nil,
	}
	// if the list is empty, insert at the head position and return
	if c.size == 0 {
		nn.next = nn
		nn.prev = nn
		c.head = nn
		c.tail = nn
		c.size++
		return nil
	}
	// if a node with a matching key is already in the list, simply
	// update the value and return
	if n := c.find(k); n != nil {
		n.val = v
		return nil
	}
	// in any other case, insert the new node. the new node becomes
	// the new head, pushing the current head next in line. link the
	// new nodes head's previous to the current tail.
	nn.next = c.head
	nn.prev = c.tail
	c.tail.next = nn
	if c.head == c.tail {
		c.head.next = nn
	}
	c.tail = nn
	c.head.prev = c.tail
	c.size++
	return nil
}

// remove takes a key and attempts to locate and remove the node with
// the matching key.
func (c *circularList[K, V]) remove(k K) {
	// attempt to locate the node
	n := c.find(k)
	if n == nil {
		return
	}
	// if the list contains only one node, free it up
	if c.size == 1 {
		c.head = nil
		c.tail = nil
		c.size--
		return
	}
	// if the node that was found was the head or tail, we
	// adjust the pointers
	if n == c.head {
		c.head = c.head.next
	}
	if n == c.tail {
		c.tail = c.tail.prev
	}
	n.next.prev = n.prev
	n.prev.next = n.next
	n = nil // problematic??
	c.size--
}

// isFull returns a boolean indicating true if the list is at capacity
// and false if there is still room.
func (c *circularList[K, V]) isFull() bool {
	return c.size == c.capacity
}

// scan is a simple closure based iterator
func (c *circularList[K, V]) scan(iter func(n *node[K, V]) bool) {
	ptr := c.head
	for i := 0; i < c.size; i++ {
		if !iter(ptr) {
			break
		}
		ptr = ptr.next
	}
}

// String is the circular list's stringer method.
func (c *circularList[K, V]) String() string {
	if c.size == 0 {
		return "nil"
	}
	var sb strings.Builder
	sb.Grow(c.size * int(unsafe.Sizeof(node[K, V]{})))
	ptr := c.head
	sb.WriteString(fmt.Sprintf("%v <- ", ptr.prev.key))
	for i := 0; i < c.size; i++ {
		if i == c.size-1 {
			sb.WriteString(fmt.Sprintf("%v", ptr.key))
		} else {
			sb.WriteString(fmt.Sprintf("%v, ", ptr.key))
		}
		ptr = ptr.next
	}
	sb.WriteString(fmt.Sprintf(" -> %v", ptr.key))
	return sb.String()
}
