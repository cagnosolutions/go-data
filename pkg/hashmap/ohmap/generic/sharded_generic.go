package generic

import (
	"fmt"
	mathbits "math/bits"
	"strings"
	"sync"

	"github.com/cagnosolutions/go-data/pkg/hash/murmur3"
	"github.com/cagnosolutions/go-data/pkg/hashmap/ohmap"
)

type shard[K comparable, V any] struct {
	mu sync.Mutex
	hm *Map[K, V] // rhh
}

type ShardedMap[K comparable, V any] struct {
	mask   uint64
	hash   hashFunc[K]
	shards []*shard[K, V]
}

// NewShardedHashMap returns a new hashMap instantiated with the specified size or
// the defaultMapSize, whichever is larger
func NewShardedMap[K comparable, V any](size uint) *ShardedMap[K, V] {
	return newShardedMap[K, V](size, defaultHashFunc[K])
}

func newShardedMap[K comparable, V any](size uint, fn hashFunc[K]) *ShardedMap[K, V] {
	shCount := alignShardCount(size)
	if fn == nil {
		fn = defaultHashFunc[K]
	}
	shm := &ShardedMap[K, V]{
		mask:   shCount - 1,
		hash:   fn,
		shards: make([]*shard[K, V], shCount),
	}
	hmSize := initialMapShardSize(uint16(shCount))
	// log.Printf("new sharded hashmap with %d shards, each shard init with %d buckets\n", shCount, hmSize)
	for i := range shm.shards {
		shm.shards[i] = &shard[K, V]{
			hm: newHashMap[K, V](hmSize, fn),
		}
	}
	return shm
}

func alignShardCount(size uint) uint64 {
	count := uint(16)
	for count < size {
		count *= 2
	}
	return uint64(count)
}

func initialMapShardSize(x uint16) uint {
	return uint(mathbits.Reverse16(x)) / 2
}

func (s *ShardedMap[K, V]) getHash(key K) (uint64, uint64) {
	// calculate the hashkey value
	hashkey := s.hash(key)
	// mask the hashkey to get the initial index
	i := hashkey & s.mask
	return i, hashkey
}

func hashStr(k string) uint64 {
	return murmur3.Sum64([]byte(k))
}

func (s *ShardedMap[K, V]) getShard(key K) (buk uint64, hashkey uint64) {
	skey := stringOf[K](key)
	// check for compound key first
	if n := strings.IndexByte(skey, ':'); n != -1 {
		// get the initial hash key
		hashkey = hashStr(skey[:n])
		// mask the hashk ey to get the initial index
		buk = hashkey & s.mask
		return buk, hashkey
	}
	// otherwise just perform normal operation and grab the hash key
	hashkey = s.hash(key)
	// mask the hashkey to get the initial index
	buk = hashkey & s.mask
	return buk, hashkey
}

func (s *ShardedMap[K, V]) Get(key K) (value V, found bool) {
	// first, we need to compute the shard
	buk, hashkey := s.getShard(key)
	// now, we lock our shard
	s.shards[buk].mu.Lock()
	// call our internal hashmap method
	value, found = s.shards[buk].hm.lookup(hashkey, key)
	// don't forget to unlock!
	s.shards[buk].mu.Unlock()
	// return what we have found
	return value, found
}

func (s *ShardedMap[K, V]) GetCollection(prefix K) (values []V, count int) {
	// first, we need to compute the shard
	buk, _ := s.getShard(prefix)
	// now, we lock our shard
	s.shards[buk].mu.Lock()
	// call our internal hashmap method to get all keys
	// matching supplied prefix
	values = s.shards[buk].hm.Filter(
		func(key K) bool {
			ok := strings.Contains(stringOf[K](key), stringOf[K](prefix))
			if ok {
				count++
			}
			return ok
		},
	)
	// don't forget to unlock!
	s.shards[buk].mu.Unlock()
	// return what we have found
	return values, count
}

func (s *ShardedMap[K, V]) Add(key K, val V) (value V, updated bool) {
	// first, we need to compute the shard
	buk, hashkey := s.getShard(key)
	// now, we lock our shard
	s.shards[buk].mu.Lock()
	// call our internal hashmap method to see if there is
	// already an existing value
	if _, exists := s.shards[buk].hm.lookup(hashkey, key); exists {
		return value, updated
	}
	// otherwise, we call our internal hashmap method to add
	// the new key and value
	value, updated = s.shards[buk].hm.insert(hashkey, key, val)
	// don't forget to unlock!
	s.shards[buk].mu.Unlock()
	// return what we have found
	return value, !updated
}

func (s *ShardedMap[K, V]) Set(key K, val V) (prev V, updated bool) {
	// first, we need to compute the shard
	buk, hashkey := s.getShard(key)
	// now, we lock our shard
	s.shards[buk].mu.Lock()
	// call our internal hashmap method
	prev, updated = s.shards[buk].hm.insert(hashkey, key, val)
	// don't forget to unlock!
	s.shards[buk].mu.Unlock()
	// return what we have found
	return prev, updated
}

func (s *ShardedMap[K, V]) Del(key K, val V) (prev V, removed bool) {
	// first, we need to compute the shard
	buk, hashkey := s.getShard(key)
	// now, we lock our shard
	s.shards[buk].mu.Lock()
	// call our internal hashmap method
	prev, removed = s.shards[buk].hm.delete(hashkey, key)
	// don't forget to unlock!
	s.shards[buk].mu.Unlock()
	// return what we have found
	return prev, removed
}

// PercentFull returns the current load factor of the HashMap
func (s *ShardedMap[K, V]) NumKeys() int {
	var keys int
	for _, sh := range s.shards {
		keys += sh.hm.Len()
	}
	return keys
}

func (s *ShardedMap[K, V]) NumBuckets() int {
	var buks int
	for _, sh := range s.shards {
		buks += len(sh.hm.buckets)
	}
	return buks
}

// PercentFull returns the current load factor of the HashMap
func (s *ShardedMap[K, V]) PercentFull() float64 {
	return float64(s.NumKeys()) / float64(s.NumBuckets())
}

func (s *ShardedMap[K, V]) Size() string {
	format := "sharded map containing %d entries (%d shards) is using %d bytes (%.2f kb, %.2f mb) of ram\n"
	sz := ohmap.Sizeof(s)
	kb := float64(sz / 1024)
	mb := float64(sz / 1024 / 1024)
	var keys int
	for _, sh := range s.shards {
		keys += sh.hm.Len()
	}
	return fmt.Sprintf(format, keys, len(s.shards), sz, kb, mb)
}

func (s *ShardedMap[K, V]) Details() string {
	ss := s.Size()
	ss += fmt.Sprintf("it is currently %.2f percent full\n", s.PercentFull())
	for i, sh := range s.shards {
		ss += fmt.Sprintf("shard[%d]\n", i)
		ss += "\tdetails:\n"
		for i := 0; i < len(sh.hm.buckets); i++ {
			if sh.hm.buckets[i].getDIB() > 0 {
				ss += fmt.Sprintf("\t\tbucket[%d]=%s", i, sh.hm.buckets[i].String())
			}
		}
	}
	return ss
}
