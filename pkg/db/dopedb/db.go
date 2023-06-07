package dopedb

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"
)

const (
	defaultBasePathDB     = "db"
	defaultMaxMemory      = 0
	defaultSyncOnInterval = -1 // 5 * time.Minute
	defaultShardCount     = 16
)

var DefaultDBConfig = defaultDBConfig

var defaultDBConfig = &DBConfig{
	BasePath:       defaultBasePathDB,
	MaxMemory:      defaultMaxMemory,
	SyncOnInterval: defaultSyncOnInterval,
	Shards:         defaultShardCount,
}

type DBConfig struct {

	// BasePath is the root folder the database will store data and
	// backup files in.
	BasePath string

	// MaxMemory is a soft limit on the amount of memory the database
	// will use (in bytes)
	MaxMemory uint64

	// SyncOnInterval instructs the database to sync the actions
	// performed at the provided interval. If set to -1, the AOF
	// backup will not be used at all. If it is set to 0, it
	// will sync on every operation. If set to 300 for example,
	// it will perform a sync every 300 seconds. Note that it will
	// not perform a sync if no requests are being made no matter
	// what the sync policy is.
	SyncOnInterval time.Duration

	Shards uint
}

type DB struct {
	conf *DBConfig
	data *ShardedHashMap
	wal  *WAL
}

func NewDB(conf *DBConfig) (*DB, error) {
	if conf == nil {
		conf = defaultDBConfig
	}
	var wal *WAL = nil
	var err error
	if conf.SyncOnInterval > -1 {
		walConf := &WALConfig{
			BasePath:    filepath.ToSlash(filepath.Join(conf.BasePath, "wal/data")),
			MaxFileSize: 16 << 10,
			SyncOnWrite: false,
		}
		if conf.SyncOnInterval == 0 {
			walConf.SyncOnWrite = true
		}
		wal, err = OpenWAL(walConf)
		if err != nil {
			return nil, err
		}
	}
	return &DB{
		conf: conf,
		data: NewShardedHashMap(conf.Shards),
		wal:  wal,
	}, nil
}

func (db *DB) checkAOFWrite() {
	if db.wal == nil {
		return
	}
}

type Record interface {
	Decode(b []byte) error
	Encode() ([]byte, error)
}

func GetAs[T Record](db *DB, k string, v T) error {
	b, found := db.data.Get(k)
	if !found {
		return fmt.Errorf("error: key=%q could not be found", k)
	}
	err := v.Decode(b)
	if err != nil {
		return err
	}
	return nil
}

func SetAs[T Record](db *DB, k string, v T) error {
	b, err := v.Encode()
	if err != nil {
		return err
	}
	_, updated := db.data.Set(k, b)
	_ = updated
	return nil
}

func numOp(v string, op int) string {
	n, err := strconv.Atoi(v)
	if err != nil {
		return ""
	}
	n += op
	return strconv.Itoa(n)
}

// func (db *DB) SetStr(k, v string) error {
// 	return SetAs(db, k, v)
// }
//
// func (db *DB) GetStr(k string) (string, error) {
// 	return GetAs()
// }

// func (db *DB) get(k string) string {
// 	v, ok := getAs[string](db, k)
// 	if !ok {
// 		return ""
// 	}
// 	return v
// }
//
// func (db *DB) incr(k string) string {
// 	v, ok := getAs[string](db, k)
// 	if !ok {
// 		return ""
// 	}
// 	v = numOp(v, +1)
// 	shard := db.hasher(k)
// 	db.st[shard].Store(k, v)
// 	return v
// }
//
// func (db *DB) decr(k string) string {
// 	v, ok := getAs[string](db, k)
// 	if !ok {
// 		return ""
// 	}
// 	v = numOp(v, -1)
// 	shard := db.hasher(k)
// 	db.st[shard].Store(k, v)
// 	return v
// }
//
// func (db *DB) zset(k string, v ...string) {
// 	shard := db.hasher(k)
// 	db.st[shard].Store(k, v)
// }
//
// func (db *DB) zget(k string) []string {
// 	v, ok := getAs[[]string](db, k)
// 	if !ok {
// 		return nil
// 	}
// 	return v
// }
// func (db *DB) hset(k string, v ...string) {
// 	if len(v)%2 != 0 {
// 		v = append(v, "")
// 	}
// 	m := make(map[string]string, len(v))
// 	for i := 0; i < len(v); i += 2 {
// 		m[v[i]] = v[i+1]
// 	}
// 	shard := db.hasher(k)
// 	db.st[shard].Store(k, m)
// }
//
// func (db *DB) hget(k string) map[string]string {
// 	v, ok := getAs[map[string]string](db, k)
// 	if !ok {
// 		return nil
// 	}
// 	return v
// }
