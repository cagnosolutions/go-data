package buffer

import (
	"fmt"
	"strings"
	"unsafe"
)

// node is a node in a circular list.
type dllNode[K comparable, V any] struct {
	key        K
	val        V
	prev, next *dllNode[K, V]
}

// String is the stringer method for a node
func (n *dllNode[K, V]) String() string {
	return fmt.Sprintf("%v<-[%v]->%v", n.prev.key, n.key, n.next.key)
}

// circularList is a circular list implementation.
type circularList[K comparable, V any] struct {
	head, tail *dllNode[K, V]
	size       uint16
	capacity   uint16
}

// newCircularList instantiates and returns a pointer to a new
// circular list instance with the capacity set using the provided
// max integer.
func newCircularList[K comparable, V any](max uint16) *circularList[K, V] {
	return &circularList[K, V]{
		head:     nil,
		tail:     nil,
		size:     0,
		capacity: max,
	}
}

// find takes a key and attempts to locate and return a node with
// the matching key. If said node cannot be found, find returns nil.
func (c *circularList[K, V]) find(k K) *dllNode[K, V] {
	ptr := c.head
	for i := uint16(0); i < c.size; i++ {
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
	nn := &dllNode[K, V]{
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
func (c *circularList[K, V]) scan(iter func(n *dllNode[K, V]) bool) {
	ptr := c.head
	for i := uint16(0); i < c.size; i++ {
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
	sb.Grow(int(c.size * uint16(unsafe.Sizeof(dllNode[K, V]{}))))
	ptr := c.head
	sb.WriteString(fmt.Sprintf("%v <- ", ptr.prev.key))
	for i := uint16(0); i < c.size; i++ {
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
