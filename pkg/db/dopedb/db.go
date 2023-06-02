package dopedb

import (
	"hash"
	"hash/fnv"
	"sync"
)

const shards = 16

type DB struct {
	st [shards]sync.Map
	h  hash.Hash32
}

func NewDB() *DB {
	return &DB{
		st: [16]sync.Map{},
		h:  fnv.New32a(),
	}
}

func (db *DB) hasher(k string) uint32 {
	defer db.h.Reset()
	db.h.Write([]byte(k))
	return db.h.Sum32() & (shards - 1)
}
