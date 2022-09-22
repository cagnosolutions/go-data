package disk

import (
	"encoding/binary"
)

const (
	order    = 255
	keylen   = 12
	pageSize = 4096
)

type tree struct {
	root *node
}

func NewNode() *node {
	n := &node{
		nkeys:  254,
		keys:   [254][12]byte{},
		ptrs:   [255]uint32{},
		parent: 99999,
		isLeaf: false,
	}
	c := [12]byte{'k', '-', 'd', 'e', 'a', 'd', 'b', 'e', 'e', 'f', '+', '+'}
	for i := range n.keys {
		n.keys[i] = c
		n.ptrs[i] = uint32(i)
	}
	return n
}

func Encode(b []byte, n *node) {
	encodeNode(b, n)
}

type node struct {
	nkeys  uint16 // max keys per node 65535
	keys   [order - 1][keylen]byte
	ptrs   [order]uint32
	parent uint32
	isLeaf bool
}

func boolToByte(ok bool) uint8 {
	if ok {
		return 1
	}
	return 0
}

// example:
// a full root node could have 255 keys and child pointers
// and each of those children could have 255 keys and child
// pointers, and each one of those could have 255 keys, and
// point to leaf nodes...
// ...which would be 255 * 255 * 255 = 16,581,375 records.

// byte 0 = isLeaf
// byte 1 = isRoot
// bytes 2-6 = parent_pointer (page)
// bytes 6-8 = num_keys
// bytes 9-12 = right child
// bytes 12-4092 = child pointer + key pairs
// bytes 4092-4094 = leftover space
//
// **each child pointer + key pair = 16 bytes
// which means we can fit 255 keys and pointers
// plus the rest of the node header information
// in a single 4096 byte page.

func encodeNode(b []byte, np *node) {
	// write node header
	var n int
	// is leaf?
	b[n] = boolToByte(np.isLeaf)
	n += 1
	// is root?
	b[n] = boolToByte(np.parent == 0)
	n += 1
	// parent pointer
	binary.LittleEndian.PutUint32(b[n:n+4], np.parent)
	n += 4
	// num keys
	binary.LittleEndian.PutUint16(b[n:n+2], np.nkeys)
	n += 2
	// right child
	binary.LittleEndian.PutUint32(b[n:n+4], np.ptrs[np.nkeys])
	n += 4
	// encode child pointers and keys...
	for i := range np.keys {
		// encode child pointer at i
		binary.LittleEndian.PutUint32(b[n:n+4], np.ptrs[i])
		n += 4
		// encode key at i
		copy(b[n:n+keylen], np.keys[i][:])
		n += keylen
	}
	// encode remaining "space"
	leftover := uint16(4<<10 - (n + 2))
	binary.LittleEndian.PutUint16(b[n:], leftover)
	n += 2
}

func decodeNode(b []byte, n *node) {

}
