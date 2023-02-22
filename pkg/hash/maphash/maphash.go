package maphash

import (
	"hash"
	"hash/maphash"
)

// Make sure interfaces are correctly implemented.
var (
	_ hash.Hash   = new(maphash.Hash)
	_ hash.Hash64 = new(maphash.Hash)
)

var digest64 maphash.Hash

func New64() hash.Hash64 {
	var h maphash.Hash
	return &h
}

func Sum64(data []byte) uint64 {
	d := &digest64
	d.Write(data) // Write never fails, n and err are simply for io.Writer compatibility
	return d.Sum64()
}
