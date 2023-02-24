package ember

import (
	"errors"
	"fmt"
	"log"
	"time"
)

type EmberConfig struct {

	// DataDir is the base directory where any
	// ember data will be stored.
	DataDir string

	// ShardCount is the number of shards to
	// use in the hashmap. Setting this too
	// high will cause extra overhead, obviously.
	// Setting it too low may have a negative
	// impact in a high load situation.
	ShardCount int

	// MaxSegments is the maximum number of wal
	// segments to keep for restoring before
	// older ones are truncated.
	MaxSegments int

	// MaxSegmentSize is the max size a wal segment
	// can be before it rotates into a new segment
	// file.
	MaxSegmentSize int64

	// SyncOnWrite enables you to force sync on write
	// which is slower, but far more durable. Or if it
	// is set to false, writes will be faster but possibly
	// a bit more risky. If SyncOnWrite is true, SyncInterval
	// is ignored.
	SyncOnWrite bool

	// SyncInterval is the longest time interval that
	// is allowed before a forceful sync is called.
	SyncInterval time.Duration
}

var DefaultEmberConfig = &EmberConfig{
	DataDir:        "ember_db",
	ShardCount:     128,
	MaxSegments:    16,
	MaxSegmentSize: 1 << 20,
	SyncOnWrite:    false,
	SyncInterval:   10 * time.Second,
}

type EmberDB struct {
	conf *EmberConfig
	db   *shardedHashMap
	wal  *WAL
}

func Open(conf *EmberConfig) (*EmberDB, error) {
	if conf == nil {
		conf = DefaultEmberConfig
	}
	f, err := OpenWAL(
		&WALConfig{
			BasePath:    conf.DataDir,
			MaxFileSize: conf.MaxSegmentSize,
			SyncOnWrite: conf.SyncOnWrite,
		},
	)
	if err != nil {
		return nil, err
	}
	db := &EmberDB{
		conf: conf,
		db:   newShardedHashMap(uint(conf.ShardCount), nil),
		wal:  f,
	}
	err = db.load()
	if err != nil {
		return nil, err
	}
	done := make(chan bool)
	background(
		done, func() {
			if err := db.wal.Sync(); err != nil {
				panic(err)
			}
		},
	)
	return db, nil
}

func (e *EmberDB) load() error {
	fn := func(b []byte) bool {
		// decode entry from wal
		it, err := decode(b)
		if err != nil {
			log.Printf("error decoding entry: %q\n", err)
		}
		// put the item back into the map
		e.db.set(it.k, it.v)
		return true
	}
	err := e.wal.Scan(fn)
	if err != nil {
		return err
	}
	return nil
}

func background(done <-chan bool, f func()) {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-done:
				log.Print("Stopping ticker")
				ticker.Stop()
				return
			case <-ticker.C:
				f()
				fmt.Println("TICK!")
			}
		}
	}()
}

const (
	m1    = 0xDEADBEEF
	m2    = 0xBEEFCAFE
	m3    = 0xDEADFACE
	m4    = 0xDEFEC8ED
	magic = m1 | m2 | m3 | m4
)

type item struct {
	k string
	v []byte
}

func (it *item) String() string {
	return fmt.Sprintf("k=%q, v=%q", it.k, it.v)
}

func encode(i *item) []byte {
	// get key length and value length
	klen, vlen := len(i.k), len(i.v)
	// create a buffer to encode into
	b := make([]byte, 8+klen+vlen)
	// add our magic header, then the key and value length
	bin.PutUint32(b[0:4], uint32(magic))
	bin.PutUint16(b[4:6], uint16(klen))
	bin.PutUint16(b[6:8], uint16(vlen))
	// copy our key and value
	copy(b[8:8+klen], i.k)
	copy(b[8+klen:8+klen+vlen], i.v)
	// return encoded item
	return b
}

func decode(b []byte) (*item, error) {
	// make sure this is a "proper" entry
	if len(b) < 8 {
		return nil, errors.New("decode: not enough to decode")
	}
	// check the magic bytes
	if bin.Uint32(b[0:4]) != magic {
		return nil, errors.New("decode: bad magic header")
	}
	// decode key length, and value length
	klen := bin.Uint16(b[4:6])
	vlen := bin.Uint16(b[6:8])
	// check key length + value length to ensure
	// that we can successfully decode it
	if int(8+klen+vlen) > len(b) {
		return nil, errors.New("decode: not enough to decode")
	}
	// create a buffer to copy into
	buf := make([]byte, klen+vlen)
	// do our copy
	copy(buf, b[8:8+klen])
	copy(buf[klen:], b[8+klen:8+klen+vlen])
	// create a new item, fill it out and return
	return &item{
		k: string(buf[:klen]),
		v: buf[klen:],
	}, nil
}

const (
	opSet = 0x04
	opDel = 0x08
)

func (e *EmberDB) Set(k string, v []byte) error {
	// first write to the wal
	_, err := e.wal.Write(encode(&item{k, v}))
	if err != nil {
		return err
	}
	// then to the map
	e.db.set(k, v)
	return nil
}

func (e *EmberDB) Get(k string) ([]byte, error) {
	// we don't need to write to the log, just get from the map
	v, found := e.db.get(k)
	if !found {
		return nil, errors.New("get: not found")
	}
	return v, nil
}

func (e *EmberDB) Del(k string) error {
	// first delete from the map
	old, found := e.db.del(k)
	if !found {
		return nil
	}
	// since it was successful, we can write delete op to the log
	_, err := e.wal.Write(encode(&item{k, old}))
	if err != nil {
		return err
	}
	return nil
}

func (e *EmberDB) Close() error {
	err := e.wal.Close()
	if err != nil {
		return err
	}
	e.db.close()
	return nil
}
