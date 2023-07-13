package generic

import (
	"hash"
	"unsafe"
)

type hdr struct {
	data unsafe.Pointer
	size uintptr
}

const (
	keyString = 1
	keyOther  = 2
)

type Hasher32[K comparable] struct {
	ktyp uint8
	ksiz uint16
	hash.Hash32
}

func NewHasher32[K comparable](h32 hash.Hash32) *Hasher32[K] {
	h := &Hasher32[K]{Hash32: h32}
	var k K
	switch ((any)(k)).(type) {
	case string:
		h.ktyp = keyString
		h.ksiz = 0
	default:
		h.ktyp = keyOther
		h.ksiz = uint16(unsafe.Sizeof(k))
	}
	return h
}

func (h *Hasher32[K]) Hash(k K) uint32 {
	if h.ktyp == keyString {
		return hash32(h, (unsafe.Pointer(&k)))
	}
	return hash32(h, (unsafe.Pointer(&hdr{unsafe.Pointer(&k), uintptr(h.ksiz)})))
}

func hash32(h hash.Hash32, p unsafe.Pointer) (sum uint32) {
	_, err := h.Write([]byte(*(*string)(p)))
	if err != nil {
		panic(err)
	}
	sum = h.Sum32()
	h.Reset()
	return sum
}

type Hasher64[K comparable] struct {
	ktyp uint8
	ksiz uint16
	hash.Hash64
}

func NewHasher64[K comparable](h64 hash.Hash64) *Hasher64[K] {
	h := &Hasher64[K]{Hash64: h64}
	var k K
	switch ((any)(k)).(type) {
	case string:
		h.ktyp = keyString
		h.ksiz = 0
	default:
		h.ktyp = keyOther
		h.ksiz = uint16(unsafe.Sizeof(k))
	}
	return h
}

func (h *Hasher64[K]) Hash(k K) uint64 {
	if h.ktyp == keyString {
		return hash64(h, (unsafe.Pointer(&k)))
	}
	return hash64(h, (unsafe.Pointer(&hdr{unsafe.Pointer(&k), uintptr(h.ksiz)})))
}

func hash64(h hash.Hash64, p unsafe.Pointer) (sum uint64) {
	_, err := h.Write([]byte(*(*string)(p)))
	if err != nil {
		panic(err)
	}
	sum = h.Sum64()
	h.Reset()
	return sum
}

func toString[K comparable](k K) string {
	switch (any(k)).(type) {
	case string:
		return *(*string)(unsafe.Pointer(&k))
	default:
		return *(*string)(unsafe.Pointer(&hdr{unsafe.Pointer(&k), unsafe.Sizeof(k)}))
	}
}

func toBytes[K comparable](k K) []byte {
	switch (any(k)).(type) {
	case string:
		return *(*[]byte)(unsafe.Pointer(&k))
	default:
		return *(*[]byte)(unsafe.Pointer(&hdr{unsafe.Pointer(&k), unsafe.Sizeof(k)}))
	}
}
