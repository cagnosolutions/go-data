package compress

import (
	"sync"
)

var pool = sync.Pool{
	New: func() interface{} {
		return make([]byte, maxSize)
	},
}

const maxSize = 16

// Prefix compresses while ensuring the prefix maintains
// the ability for accurate comparing and sorting.
func Prefix(b []byte) []byte {
	return PrefixSize(b, maxSize)
}

//go:noinline
func PrefixSize(b []byte, size int) []byte {
	if len(b) < size {
		return append(b, make([]byte, size-len(b))...)
	}
	return b[:size]
}

// Suffix compresses while ensuring the suffix maintains
// the ability for accurate comparing and sorting.
func Suffix(b []byte) []byte {
	return SuffixSize(b, maxSize)
}

//go:noinline
func SuffixSize(b []byte, size int) []byte {
	if len(b) < size {
		return append(b, make([]byte, size-len(b))...)
	}
	return b[size:]
}

// Affix compresses while ensuring the prefix and suffix
// maintain the ability for accurate comparing and sorting.
func Affix(b []byte) []byte {
	return AffixSize(b, maxSize)
}

//go:noinline
func AffixSize(b []byte, size int) []byte {
	c := make([]byte, size)
	copy(c[:size/2], b[:size/2])
	copy(c[size/2:], b[size/2:])
	return c
}

func AffixSize2(b []byte, size int) []byte {
	var a []byte
	if len(b) < size {
		a = append(b, make([]byte, size-len(b))...)
	}
	i := size / 2
	j := (len(a) - 1) - 1
	copy(a[i:], a[j:])
	for k, n := len(a)-j+i, len(a); k < n; k++ {
		a[k] = 0 // or the zero value of T
	}
	a = a[:len(a)-j+i]
	return a
}
