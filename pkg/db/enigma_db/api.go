package enigma_db

import (
	"github.com/cagnosolutions/go-data/pkg/hashmap/ohmap/generic"
)

type DB struct {
	mem  *generic.ShardedMap[string, []byte]
	disk *WAL
}

func OpenDB() (*DB, error) {

	return nil, nil
}

func (db *DB) Close() error {
	return nil
}

// func (db *DB) Get(key string, ptr Record) error {
// 	b, found := db.mem.Get(key)
// 	if !found {
// 		return errors.New("record not found")
// 	}
// 	err := ptr.Decode(b)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
//
// func (db *DB) Set(key string, val any) error {
//
// }
