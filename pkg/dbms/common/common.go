package common

import (
	"encoding/binary"
	"hash/crc32"
	"log"
	"os"
)

var Binary = binary.LittleEndian

var (
	debug = log.New(os.Stdout, "::[DEBUG] >> ", log.Lshortfile|log.Lmsgprefix)
)

/*
type hashSet[T comparable] map[T]struct{}

func makeMapSet[T comparable](size int) hashSet[T] {
	return make(hashSet[T], size)
}

func (hs hashSet[T]) add(data T) {
	hs[data] = struct{}{}
}

func (hs hashSet[T]) del(data T) {
	delete(hs, data)
}

func (hs hashSet[T]) get() (T, bool) {
	for d := range hs {
		return d, true
	}
	var zero T
	return zero, false
}
*/

// checksum is a checksum calculator
func checksum(p []byte) uint32 {
	return crc32.Checksum(p, crc32.MakeTable(crc32.Koopman))
}

func AlignSize(size, count uint) uint {
	return alignSize(size, count)
}

func alignSize(size, count uint) uint {
	for count < size {
		count *= 2
	}
	return count
}
