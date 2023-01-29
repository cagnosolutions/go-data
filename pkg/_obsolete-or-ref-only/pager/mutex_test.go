package pager

import (
	"fmt"
	"hash"
	"hash/fnv"
	"strconv"
	"sync"
	"testing"
)

const (
	lockTypeNone   = 0
	lockTypeLower  = 1
	lockTypeHigher = 2
)

var add = func(sm *smap, mu *sync.RWMutex) {
	if mu != nil {
		mu.Lock()
		defer mu.Unlock()
	}
	for i := 0; i < 128; i++ {
		sm.set(strconv.Itoa(i), strconv.Itoa(i))
	}
}
var get = func(sm *smap, mu *sync.RWMutex) {
	if mu != nil {
		mu.RLock()
		defer mu.RUnlock()
	}
	for i := 0; i < 128; i++ {
		v := sm.get(strconv.Itoa(i))
		_ = v
	}
}
var del = func(sm *smap, mu *sync.RWMutex) {
	if mu != nil {
		mu.Lock()
		defer mu.Unlock()
	}
	for i := 0; i < 128; i++ {
		sm.del(strconv.Itoa(i))
	}
}

func Benchmark_SMap_Higher_Locks(b *testing.B) {
	sm := newSMap(16, lockTypeHigher)
	for i := 0; i < b.N; i++ {
		add(sm, nil)
		get(sm, nil)
		del(sm, nil)
	}
}

func Benchmark_SMap_Super_Higher_Locks(b *testing.B) {
	sm := newSMap(16, lockTypeNone)
	var mu sync.RWMutex
	for i := 0; i < b.N; i++ {
		add(sm, &mu)
		get(sm, &mu)
		del(sm, &mu)
	}
}

func Benchmark_SMap_Lower_Locks(b *testing.B) {
	sm := newSMap(16, lockTypeLower)
	for i := 0; i < b.N; i++ {
		add(sm, nil)
		get(sm, nil)
		del(sm, nil)
	}
}

func Test_SMap_Higher_Locks(t *testing.T) {
	sm := newSMap(16, lockTypeHigher)
	go add(sm, nil)
	go del(sm, nil)
	go get(sm, nil)
	fmt.Println(sm)
}

func Test_SMap_Lower_Locks(t *testing.T) {
	sm := newSMap(16, lockTypeLower)
	go add(sm, nil)
	go del(sm, nil)
	go get(sm, nil)
	fmt.Println(sm)
}

func Test_SMap_No_Locks(t *testing.T) {
	sm := newSMap(16, lockTypeNone)
	go add(sm, nil)
	go del(sm, nil)
	go get(sm, nil)
	fmt.Println(sm)
}

type shard struct {
	data map[string]string
}

func (s shard) set(k, v string, lock *sync.RWMutex) {
	if lock != nil {
		lock.Lock()
		defer lock.Unlock()
	}
	s.data[k] = v
}

func (s shard) get(k string, lock *sync.RWMutex) string {
	if lock != nil {
		lock.RLock()
		defer lock.RUnlock()
	}
	v, found := s.data[k]
	if !found {
		return ""
	}
	return v
}

func (s shard) del(k string, lock *sync.RWMutex) {
	if lock != nil {
		lock.Lock()
		defer lock.Unlock()
	}
	delete(s.data, k)
}

func (s shard) String() string {
	return fmt.Sprintf("%+v\n", s.data)
}

type smap struct {
	shards   []shard
	mask     uint32
	hash     hash.Hash32
	lock     sync.RWMutex
	lockType int
}

func (s *smap) String() string {
	if s.lockType > lockTypeNone {
		s.lock.RLock()
		defer s.lock.RUnlock()
	}
	var ss string
	for i, sh := range s.shards {
		ss += fmt.Sprintf("shard[%d]=%d entries\n", i, len(sh.data))
	}
	return ss
}

func newSMap(count int, lockType int) *smap {
	if 8 > count || count > 64 {
		count = 8
	}
	shards := make([]shard, count)
	for i := range shards {
		shards[i] = shard{
			data: make(map[string]string),
		}
	}
	return &smap{
		shards:   shards,
		mask:     uint32(count - 1),
		hash:     fnv.New32a(),
		lockType: lockType,
	}
}

func (s *smap) bucket(k string) int {
	_, err := s.hash.Write([]byte(k))
	if err != nil {
		panic(err)
	}
	buk := s.hash.Sum32() & s.mask
	// fmt.Printf("hash=%v, k=%q, buk=%d\n", s.hash.Sum32(), k, buk)
	s.hash.Reset()
	return int(buk)
}

func (s *smap) set(k, v string) {
	if s.lockType == lockTypeHigher {
		s.lock.Lock()
		defer s.lock.Unlock()
	}
	sh := s.shards[s.bucket(k)]
	if s.lockType == lockTypeLower {
		sh.set(k, v, &s.lock)
		return
	}
	sh.set(k, v, nil)
}

func (s *smap) get(k string) string {
	if s.lockType == lockTypeHigher {
		s.lock.RLock()
		defer s.lock.RUnlock()
	}
	sh := s.shards[s.bucket(k)]
	if s.lockType == lockTypeLower {
		return sh.get(k, &s.lock)
	}
	return sh.get(k, nil)
}

func (s *smap) del(k string) {
	if s.lockType == lockTypeHigher {
		s.lock.Lock()
		defer s.lock.Unlock()
	}
	sh := s.shards[s.bucket(k)]
	if s.lockType == lockTypeLower {
		sh.del(k, &s.lock)
		return
	}
	sh.del(k, nil)
}
