package ember

import (
	"encoding/binary"
	"fmt"
	mathbits "math/bits"
	"runtime"
	"sync"
)

type shard struct {
	mu sync.Mutex
	hm *hashmap // rhh
}

type shardedHashMap struct {
	mask   uint64
	hash   hashFunc
	shards []*shard
}

func newShardedHashMap(size uint, fn hashFunc) *shardedHashMap {
	shCount := alignShardCount(size)
	if fn == nil {
		fn = defaultHashFunc
	}
	shm := &shardedHashMap{
		mask:   shCount - 1,
		hash:   fn,
		shards: make([]*shard, shCount),
	}
	hmSize := initialMapShardSize(uint16(shCount))
	// log.Printf("new sharded hashmap with %d shards, each shard init with %d buckets\n", shCount, hmSize)
	for i := range shm.shards {
		shm.shards[i] = &shard{
			hm: newHashMap(hmSize, fn),
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

func (s *shardedHashMap) getShard(key string) (uint64, uint64) {
	// calculate the hashkey value
	hashkey := s.hash(key)
	// mask the hashkey to get the initial index
	i := hashkey & s.mask
	return i, hashkey
}

func (s *shardedHashMap) add(key string, val []byte) ([]byte, bool) {
	if _, ok := s.lookup(key); ok {
		return nil, false // returns false if it didnt add
	}
	return s.insert(key, val)
}

func (s *shardedHashMap) set(key string, val []byte) ([]byte, bool) {
	return s.insert(key, val)
}

func (s *shardedHashMap) setBit(key string, idx uint, bit uint) bool {
	if bit != 0 && bit != 1 {
		return false
	}
	buk, hashkey := s.getShard(key)
	s.shards[buk].mu.Lock()
	ret, _ := s.shards[buk].hm.lookup(hashkey, key)
	if bit == 1 {
		rawbytesSet(&ret, idx)
	}
	if bit == 0 {
		rawbytesSet(&ret, idx)
	}
	_, _ = s.shards[buk].hm.insert(hashkey, key, ret)
	s.shards[buk].mu.Unlock()
	return true
}

func (s *shardedHashMap) getBit(key string, idx uint) (uint, bool) {
	buk, hashkey := s.getShard(key)
	s.shards[buk].mu.Lock()
	ret, ok := s.shards[buk].hm.lookup(hashkey, key)
	if ret == nil || !ok || idx > uint(len(ret)*8) {
		return 0, false
	}
	s.shards[buk].mu.Unlock()
	bit := rawbytesGet(&ret, idx)
	return bit, bit != 0
}

func (s *shardedHashMap) setUint(key string, num uint64) (uint64, bool) {
	buk, hashkey := s.getShard(key)
	s.shards[buk].mu.Lock()
	val := make([]byte, 8)
	binary.LittleEndian.PutUint64(val, num)
	ret, ok := s.shards[buk].hm.insert(hashkey, key, val)
	if !ok {
		s.shards[buk].mu.Unlock()
		return 0, false
	}
	s.shards[buk].mu.Unlock()
	return binary.LittleEndian.Uint64(ret), true
}

func (s *shardedHashMap) getUint(key string) (uint64, bool) {
	buk, hashkey := s.getShard(key)
	s.shards[buk].mu.Lock()
	ret, ok := s.shards[buk].hm.lookup(hashkey, key)
	if !ok {
		s.shards[buk].mu.Unlock()
		return 0, false
	}
	s.shards[buk].mu.Unlock()
	return binary.LittleEndian.Uint64(ret), true
}

func (s *shardedHashMap) insert(key string, val []byte) ([]byte, bool) {
	buk, hashkey := s.getShard(key)
	s.shards[buk].mu.Lock()
	pv, ok := s.shards[buk].hm.insert(hashkey, key, val)
	s.shards[buk].mu.Unlock()
	return pv, ok
}

func (s *shardedHashMap) get(key string) ([]byte, bool) {
	return s.lookup(key)
}

func (s *shardedHashMap) lookup(key string) ([]byte, bool) {
	buk, hashkey := s.getShard(key)
	s.shards[buk].mu.Lock()
	pv, ok := s.shards[buk].hm.lookup(hashkey, key)
	s.shards[buk].mu.Unlock()
	return pv, ok
}

func (s *shardedHashMap) del(key string) ([]byte, bool) {
	return s.delete(key)
}

func (s *shardedHashMap) delete(key string) ([]byte, bool) {
	buk, hashkey := s.getShard(key)
	s.shards[buk].mu.Lock()
	pv, ok := s.shards[buk].hm.delete(hashkey, key)
	s.shards[buk].mu.Unlock()
	return pv, ok
}

func (s *shardedHashMap) Len() int {
	var length int
	for i := range s.shards {
		s.shards[i].mu.Lock()
		length += s.shards[i].hm.Len()
		s.shards[i].mu.Unlock()
	}
	return length
}

func (s *shardedHashMap) iter(it Iterator) {
	for i := range s.shards {
		s.shards[i].mu.Lock()
		s.shards[i].hm.Range(it)
		s.shards[i].mu.Unlock()
	}
}

func (s *shardedHashMap) stats() {
	for i := range s.shards {
		s.shards[i].mu.Lock()
		if pf := s.shards[i].hm.percentFull(); pf > 0 {
			fmt.Printf("shard %d, fill percent: %.4f\n", i, pf)
		}
		s.shards[i].mu.Unlock()
	}
}

func (s *shardedHashMap) close() {
	for i := range s.shards {
		s.shards[i].mu.Lock()
		destroyMap(s.shards[i].hm)
		s.shards[i].mu.Unlock()
	}
	s.shards = nil
	runtime.GC()
}
