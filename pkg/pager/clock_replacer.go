package pager

import (
	"errors"
	"fmt"
	"strings"
	"unsafe"
)

var ErrNoVictimFound = errors.New("replacer may be empty; victim could not be found")

// clockReplacer represents a clock based replacement cache
type clockReplacer struct {
	list *circularList
	ptr  **node
}

// newClockReplacer instantiates and returns a new clockReplacer
func newClockReplacer(size int) *clockReplacer {
	list := newCircularList(size)
	return &clockReplacer{
		list: list,
		ptr:  &list.head,
	}
}

// pin takes a frameID and pins a pageFrame indicating that it should not
// be victimized until it is unpinned
func (c *clockReplacer) pin(pid frameID) {
	n := c.list.find(pid)
	if n == nil {
		return
	}
	if (*c.ptr) == n {
		c.ptr = &(*c.ptr).next
	}
	c.list.remove(pid)
}

// unpin takes a frameID and unpins a pageFrame indicating that it is now
// available for victimization
func (c *clockReplacer) unpin(k frameID) {
	if !c.list.hasKey(k) {
		c.list.insert(k, true)
		if c.list.size == 1 {
			c.ptr = &c.list.head
		}
	}
}

// victim removes the victim pageFrame as defined by the replacement policy
// and returns the frameID of the victim
func (c *clockReplacer) victim() *frameID {
	if c.list.size == 0 {
		return nil
	}
	var victim *frameID
	cn := *c.ptr
	for {
		if cn.val {
			cn.val = false
			c.ptr = &cn.next
		} else {
			fid := cn.key
			victim = &fid
			c.ptr = &cn.next
			c.list.remove(cn.key)
			return victim
		}
	}
}

// size returns the size of the clockReplacer
func (c *clockReplacer) size() int {
	return c.list.size
}

// ErrListIsFull errors reports when the circular list is at capacity
var ErrListIsFull = errors.New("list is full; circular list capacity met")

// node is a node in a circular list.
type node struct {
	key        frameID
	val        bool
	prev, next *node
}

// String is the stringer method for a node
func (n *node) String() string {
	return fmt.Sprintf("%v<-[%v]->%v", n.prev.key, n.key, n.next.key)
}

// circularList is a circular list implementation.
type circularList struct {
	head, tail *node
	size       int
	capacity   int
}

// newCircularList instantiates and returns a pointer to a new
// circular list instance with the capacity set using the provided
// max integer.
func newCircularList(max int) *circularList {
	return &circularList{
		head:     nil,
		tail:     nil,
		size:     0,
		capacity: max,
	}
}

// find takes a key and attempts to locate and return a node with
// the matching key. If said node cannot be found, find returns nil.
func (c *circularList) find(k frameID) *node {
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
func (c *circularList) hasKey(k frameID) bool {
	return c.find(k) != nil
}

// insert takes a key and value and inserts it into the list, unless
// the list is at its capacity.
func (c *circularList) insert(k frameID, v bool) error {
	// check capacity
	if c.size == c.capacity {
		return ErrListIsFull
	}
	// create new node to insert
	nn := &node{
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
func (c *circularList) remove(k frameID) {
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
func (c *circularList) isFull() bool {
	return c.size == c.capacity
}

// scan is a simple closure based iterator
func (c *circularList) scan(iter func(n *node) bool) {
	ptr := c.head
	for i := 0; i < c.size; i++ {
		if !iter(ptr) {
			break
		}
		ptr = ptr.next
	}
}

// String is the circular list's stringer method.
func (c *circularList) String() string {
	if c.size == 0 {
		return "nil"
	}
	var sb strings.Builder
	sb.Grow(c.size * int(unsafe.Sizeof(node{})))
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
