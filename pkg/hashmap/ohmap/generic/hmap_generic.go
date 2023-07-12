package generic

import (
	"fmt"
	"unsafe"

	"github.com/cagnosolutions/go-data/pkg/hash/murmur3"
	"github.com/cagnosolutions/go-data/pkg/hashmap/ohmap"
)

const (
	dibBitSize        = 16 // 0xffff
	dibBitMask        = ^uint64(0) >> hashBitSize
	hashBitSize       = 64 - dibBitSize // 0xffffffffffff
	hashBitMask       = ^uint64(0) >> dibBitSize
	defaultLoadFactor = 0.85 // must be between 55% and 95%
	defaultCapacity   = 32
)

// entry is a key value pair that is found in each bucket
type entry[K comparable, V any] struct {
	key K
	val V
}

func newBucket[K comparable, V any](hash uint64, key K, value V) bucket[K, V] {
	return bucket[K, V]{
		bitfield: makeBitField(hash, 1),
		entry: entry[K, V]{
			key: key,
			val: value,
		},
	}
}

// bucket represents a single slot in the Map table
type bucket[K comparable, V any] struct {
	bitfield uint64 // bitfield storing dib in 16 bits, and the hash key in 48 bits
	entry[K, V]
}

func (b *bucket[K, V]) getDIB() int {
	return int(b.bitfield & dibBitMask)
	// return int(b.dib)
}

func (b *bucket[K, V]) getHash() uint64 {
	return b.bitfield >> dibBitSize
	// return b.hash
}

func (b *bucket[K, V]) setDIB(dib int) {
	b.bitfield = b.bitfield>>dibBitSize<<dibBitSize | uint64(dib)&dibBitMask
	// b.dib = uint16(dib)
}

func (b *bucket[K, V]) setHash(hash int) {
	b.bitfield = uint64(hash)<<dibBitSize | b.bitfield&dibBitMask
	// b.hash = uint64(hash)
}

// checkHashAndKey checks if this bucket matches the specified hk and key
func (b *bucket[K, V]) checkHashAndKey(hashkey uint64, key K) bool {
	// return uint64(b.getHash()) == hashkey && b.entry.key == key
	return b.getHash() == hashkey && b.key == key
}

func (b *bucket[K, V]) clear() {
	b.bitfield = 0
	b.entry = entry[K, V]{}
}

func (b *bucket[K, V]) String() string {
	return fmt.Sprintf(
		"{bitfield:%d, dib:%d, hash:%d, key:%v, val:%v}\n",
		b.bitfield, b.getDIB(), b.getHash(), b.key, b.val,
	)
}

func makeBitField(hash, dib uint64) uint64 {
	return hash<<dibBitSize | dib&dibBitMask
}

// Map represents a closed hashing hashtable implementation
type Map[K comparable, V any] struct {
	hash    hashFunc[K]
	mask    uint64
	expand  uint
	shrink  uint
	keys    uint
	cap     uint
	buckets []bucket[K, V]
}

// defaultHashFunc is the default hashFunc used. This is here mainly as
// a convenience for the sharded hashmap to utilize
func defaultHashFunc[K comparable](key K) uint64 {
	return murmur3.Sum64([]byte(stringOf[K](key)))
}

// hashFunc is a type definition for what a hash function should look like
type hashFunc[K comparable] func(key K) uint64

func stringOf[K comparable](k K) string {
	var r string
	switch ((any)(k)).(type) {
	case string:
		r = *(*string)(unsafe.Pointer(&k))
	default:
		r = *(*string)(
			unsafe.Pointer(
				&struct {
					data unsafe.Pointer
					size int
				}{
					data: unsafe.Pointer(&k),
					size: int(unsafe.Sizeof(k)),
				},
			),
		)

		// new way??
		// r = unsafe.String((*byte)(unsafe.Pointer(&k)), int(unsafe.Sizeof(k)))
	}
	return r
}

// NewHashMap returns a new Map instantiated with the specified cap or
// the DefaultMapSize, whichever is larger
func NewMap[K comparable, V any](cap uint) *Map[K, V] {
	return newHashMap[K, V](cap, defaultHashFunc[K])
}

// alignBucketCount aligns buckets to ensure all sizes are powers of two
func alignBucketCount(size uint) uint64 {
	count := uint(defaultCapacity)
	for count < size {
		count *= 2
	}
	return uint64(count)
}

// newHashMap is the internal variant of the previous function
// and is mainly used internally
func newHashMap[K comparable, V any](cap uint, hash hashFunc[K]) *Map[K, V] {
	bukCnt := alignBucketCount(cap)
	if hash == nil {
		hash = defaultHashFunc[K]
	}
	m := &Map[K, V]{
		hash:    hash,
		mask:    bukCnt - 1, // this minus one is extremely important for using a mask over modulo
		expand:  uint(float64(bukCnt) * defaultLoadFactor),
		shrink:  uint(float64(bukCnt) * (1 - defaultLoadFactor)),
		keys:    0,
		cap:     cap,
		buckets: make([]bucket[K, V], bukCnt),
	}
	return m
}

func (m *Map[K, V]) getHashKey(key K) uint64 {
	return m.hash(key) >> dibBitSize
}

// resize grows or shrinks the Map by the newSize provided. It makes a
// new map with the new cap, copies everything over, and then frees the old map
func (m *Map[K, V]) resize(newSize uint) {
	newHM := newHashMap[K, V](newSize, m.hash)
	var buk bucket[K, V]
	for i := 0; i < len(m.buckets); i++ {
		buk = m.buckets[i]
		if buk.getDIB() > 0 {
			newHM.insertInternal(buk.getHash(), buk.entry.key, buk.entry.val)
		}
	}
	tsize := m.cap
	*m = *newHM
	m.cap = tsize
}

// Get returns a value for a given key, or returns false if none could be found
// Get can be considered the exported version of the lookup call
func (m *Map[K, V]) Get(key K) (V, bool) {
	return m.lookup(0, key)
}

// lookup returns a value for a given key, or returns false if none could be found
func (m *Map[K, V]) lookup(hashkey uint64, key K) (value V, ok bool) {
	// check if map is empty
	if len(m.buckets) == 0 {
		// hopefully this should never really happen
		// do we really need to check this here?
		// *m = *newHashMap[K, V](defaultCapacity, m.hash)
		return value, false
	}
	if hashkey == 0 {
		// calculate the hk value
		hashkey = m.getHashKey(key)
	}
	// mask the hk to get the initial index
	i := hashkey & m.mask
	// search the position linearly
	for {
		// haven't located anything
		if m.buckets[i].getDIB() == 0 {
			return value, false
		}
		// check for matching hashes and keys
		if m.buckets[i].checkHashAndKey(hashkey, key) {
			return m.buckets[i].entry.val, true
		}
		// keep on probing
		i = (i + 1) & m.mask
	}
}

// helper
func (m *Map[K, V]) getBucket(index uint64) *bucket[K, V] {
	return &m.buckets[index]
}

// Set inserts a key value entry and returns the previous value or false
// Set can be considered the exported version of the insert call
func (m *Map[K, V]) Set(key K, value V) (V, bool) {
	return m.insert(0, key, value)
}

// insert inserts a key value entry and returns the previous value, or false
func (m *Map[K, V]) insert(hashkey uint64, key K, value V) (V, bool) {
	// check if map is empty
	if len(m.buckets) == 0 {
		// create a new map with default cap
		*m = *newHashMap[K, V](defaultCapacity, m.hash)
	}
	// check and see if we need to resize
	if m.keys >= m.expand {
		// if we do, then double the map cap
		m.resize(uint(len(m.buckets)) * 2)
	}
	if hashkey == 0 {
		// calculate the hk value
		hashkey = m.getHashKey(key)
	}
	// call the internal insert to insert the entry
	return m.insertInternal(hashkey, key, value)
}

// insertInternal inserts a key value entry and returns the previous value, or false
func (m *Map[K, V]) insertInternal(hashkey uint64, key K, value V) (prev V, updated bool) {
	// create a new entry to insert
	newb := bucket[K, V]{
		bitfield: makeBitField(hashkey, 1),
		entry: entry[K, V]{
			key: key,
			val: value,
		},
	}
	// mask the hash to get the initial index
	i := newb.getHash() & m.mask
	// search the position linearly
	for {
		// we found a spot, insert a new entry
		if m.buckets[i].getDIB() == 0 {
			m.buckets[i] = newb
			m.keys++
			// no previous value to return, as this is a new entry
			return prev, false
		}
		// found existing entry, check hashes and keys
		if m.buckets[i].checkHashAndKey(newb.getHash(), newb.entry.key) {
			// hashes and keys are a match--update entry and return previous values
			prev = m.buckets[i].entry.val
			m.buckets[i].val = newb.entry.val
			return prev, true
		}
		// we did not find an empty slot or an existing matching entry
		// so check this entries bf against our new entry's bf
		if m.buckets[i].getDIB() < newb.getDIB() {
			// current position's bf is less than our new entry's, swap
			newb, m.buckets[i] = m.buckets[i], newb
		}
		// keep on probing until we find what we're looking for.
		// increase our search index by one as well as our new
		// entry's bf, then continue with the linear probe.
		i = (i + 1) & m.mask
		newb.setDIB(newb.getDIB() + 1)
	}
}

// Del removes a value for a given key and returns the deleted value, or false
// Del can be considered the exported version of the delete call
func (m *Map[K, V]) Del(key K) (V, bool) {
	return m.delete(0, key)
}

// delete removes a value for a given key and returns the deleted value, or false
func (m *Map[K, V]) delete(hashkey uint64, key K) (prev V, removed bool) {
	// check if map is empty
	if len(m.buckets) == 0 {
		// nothing to see here folks
		return prev, false
	}
	if hashkey == 0 {
		// calculate the hk value
		hashkey = m.getHashKey(key)
	}
	// mask the hk to get the initial index
	i := hashkey & m.mask
	// search the position linearly
	for {
		// haven't located anything
		if m.buckets[i].getDIB() == 0 {
			return prev, false
		}
		// found existing entry, check hashes and keys
		if m.buckets[i].checkHashAndKey(hashkey, key) {
			// hashes and keys are a match--delete entry and return previous values
			prev = m.buckets[i].entry.val
			m.deleteInternal(i)
			return prev, true
		}
		// keep on probing until we find what we're looking for.
		// increase our search index by one as well as our new
		// entry's bf, then continue with the linear probe.
		i = (i + 1) & m.mask
	}
}

// delete removes a value for a given key and returns the deleted value, or false
func (m *Map[K, V]) deleteInternal(i uint64) {
	// set bf at bucket i
	m.buckets[i].setDIB(0)
	// tombstone index and shift
	for {
		pi := i
		i = (i + 1) & m.mask
		if m.buckets[i].getDIB() <= 1 {
			// im as free as a bird now!
			m.buckets[pi].clear()
			break
		}
		// shift
		m.buckets[pi] = m.buckets[i]
		m.buckets[pi].setDIB(m.buckets[pi].getDIB() - 1)
	}
	// decrement entry count
	m.keys--
	// check and see if we need to resize
	if m.keys <= m.shrink && uint(len(m.buckets)) > m.cap {
		// if it checks out, then resize down by 25%-ish
		m.resize(m.keys)
	}
}

// Range takes an iterator function and ranges the Map as long
// as the iterator function continues to be true. Range is not
// safe to perform an insert or remove operation while ranging!
func (m *Map[K, V]) Range(f func(key K, val V) bool) {
	for i := 0; i < len(m.buckets); i++ {
		if m.buckets[i].getDIB() < 1 {
			continue
		}
		if !f(m.buckets[i].key, m.buckets[i].val) {
			return
		}
	}
}

func (m *Map[K, V]) Filter(f func(key K) bool) (values []V) {
	for i := 0; i < len(m.buckets); i++ {
		if m.buckets[i].getDIB() < 1 {
			continue
		}
		if f(m.buckets[i].key) {
			values = append(values, m.buckets[i].val)
		}
	}
	return values
}

func (m *Map[K, V]) Keys() []K {
	keys := make([]K, 0, m.keys)
	for i := 0; i < len(m.buckets); i++ {
		if m.buckets[i].getDIB() > 0 {
			keys = append(keys, m.buckets[i].entry.key)
		}
	}
	return keys
}

func (m *Map[K, V]) Vals() []V {
	vals := make([]V, 0, m.keys)
	for i := 0; i < len(m.buckets); i++ {
		if m.buckets[i].getDIB() > 0 {
			vals = append(vals, m.buckets[i].entry.val)
		}
	}
	return vals
}

// PercentFull returns the current load factor of the HashMap
func (m *Map[K, V]) PercentFull() float64 {
	return float64(m.keys) / float64(len(m.buckets))
}

// Len returns the number of entries currently in the HashMap
func (m *Map[K, V]) Len() int {
	return int(m.keys)
}

// Close closes and frees the current hashmap. Calling any method
// on the HashMap after this will most likely result in a panic
func (m *Map[K, V]) Close() {
	destroyMap(m)
}

// destroy does exactly what is sounds like it does
func destroyMap[K comparable, V any](m *Map[K, V]) {
	m = nil
}

func (m *Map[K, V]) Size() string {
	format := "map containing %d entries is using %d bytes (%.2f kb, %.2f mb) of ram\n"
	sz := ohmap.Sizeof(m)
	kb := float64(sz / 1024)
	mb := float64(sz / 1024 / 1024)
	return fmt.Sprintf(format, int(m.keys), sz, kb, mb)
}

func (m *Map[K, V]) Details() string {
	ss := m.Size()
	ss += fmt.Sprintf("it is currently %.2f percent full\n", m.PercentFull())
	ss += "Bucket details...\n"
	for i := 0; i < len(m.buckets); i++ {
		if m.buckets[i].getDIB() > 0 {
			ss += fmt.Sprintf("\tbucket[%d]=%s", i, m.buckets[i].String())
		}
	}
	return ss
}
