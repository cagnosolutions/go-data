package bpt

import (
	"hash/fnv"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

// hashSuffix generates a short hash-based suffix to improve uniqueness
func hashSuffix(s string, length int) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	hashValue := h.Sum32()
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = alphabet[(hashValue>>uint(i*5))%uint32(len(alphabet))]
	}
	return string(result)
}

// compress truncates or modifies the input byte slice `key` to a fixed length `n` while
// preserving uniqueness. If `key` is shorter than `n`, it is padded with zeros. Otherwise,
// a prefix, sampled characters, and a hash are used. The prefix is the first `prefixLen`
// bytes of `key`, while the suffix includes sampled characters and a hash value.
func compress(key []byte, n int) []byte {
	// handle early error
	if key == nil || len(key) == 0 {
		return nil
	}

	// precompute fixed size result
	if len(key) <= n {
		res := make([]byte, n)
		copy(res, key)
		return res // return early
	}

	prefixLen := n / 2         // Keep the first half unchanged
	suffixLen := n - prefixLen // Space left for sampled chars + hash

	res := make([]byte, n)

	copy(res, key[:prefixLen]) // Copy prefix directly

	// Sample key points from the remaining characters
	step := float64(len(key)-prefixLen) / float64(suffixLen)
	for i := 0; i < suffixLen-2; i++ {
		res[prefixLen+i] = key[prefixLen+int(step*float64(i))]
	}

	// Append 2-character hash suffix for more even key distribution
	h := fnv.New32a()
	h.Write(key)
	hash := h.Sum32()
	alphalen := len(alphabet)
	res[prefixLen+suffixLen-2] = alphabet[(hash>>uint(0*5))%uint32(alphalen)]
	res[prefixLen+suffixLen-1] = alphabet[(hash>>uint(1*5))%uint32(alphalen)]

	// return result
	return res
}

func compressOpt2(key []byte, res *[8]byte) {
	// handle early errors
	if key == nil || len(key) == 0 {
		return
	}

	const sz = len(res)

	// precompute fixed size result
	if len(key) <= sz {
		copy((*res)[:], key)
		return // return early
	}

	prefixLen := sz / 2         // Keep the first half unchanged
	suffixLen := sz - prefixLen // Space left for sampled chars + hash

	copy((*res)[:], key[:prefixLen]) // Copy prefix directly

	// Sample key points from the remaining characters
	step := float64(len(key)-prefixLen) / float64(suffixLen)
	for i := 0; i < suffixLen-2; i++ {
		(*res)[prefixLen+i] = key[prefixLen+int(step*float64(i))]
	}

	// Append 2-character hash suffix for more even key distribution
	h := fnv.New32a()
	h.Write(key)
	hash := h.Sum32()
	alphalen := len(alphabet)
	(*res)[prefixLen+suffixLen-2] = alphabet[(hash>>uint(0*5))%uint32(alphalen)]
	(*res)[prefixLen+suffixLen-1] = alphabet[(hash>>uint(1*5))%uint32(alphalen)]

	// return result
	return
}

func compressOpt(key []byte, res *[]byte, n int) {
	// handle early errors
	if len(*res) < n {
		return
	}

	if key == nil || len(key) == 0 {
		return
	}

	// precompute fixed size result
	if len(key) <= n {
		copy((*res)[:], key)
		return // return early
	}

	prefixLen := n / 2         // Keep the first half unchanged
	suffixLen := n - prefixLen // Space left for sampled chars + hash

	copy((*res)[:], key[:prefixLen]) // Copy prefix directly

	// Sample key points from the remaining characters
	step := float64(len(key)-prefixLen) / float64(suffixLen)
	for i := 0; i < suffixLen-2; i++ {
		(*res)[prefixLen+i] = key[prefixLen+int(step*float64(i))]
	}

	// Append 2-character hash suffix for more even key distribution
	h := fnv.New32a()
	h.Write(key)
	hash := h.Sum32()
	alphalen := len(alphabet)
	(*res)[prefixLen+suffixLen-2] = alphabet[(hash>>uint(0*5))%uint32(alphalen)]
	(*res)[prefixLen+suffixLen-1] = alphabet[(hash>>uint(1*5))%uint32(alphalen)]

	// return result
	return
}

// compressString improves uniqueness while maintaining lexicographic order
func compressString(s string, n int) string {
	if len(s) <= n {
		return s
	}

	prefixLength := n / 2            // Keep the first half unchanged
	suffixLength := n - prefixLength // Space left for sampled chars + hash

	result := make([]byte, n)
	copy(result, s[:prefixLength]) // Copy prefix directly

	// Sample key points from the remaining characters
	step := float64(len(s)-prefixLength) / float64(suffixLength)
	for i := 0; i < suffixLength-2; i++ {
		index := prefixLength + int(step*float64(i))
		result[prefixLength+i] = s[index]
	}

	// Append 2-character hash suffix
	hashPart := hashSuffix(s, 2)
	copy(result[prefixLength+suffixLength-2:], hashPart)

	return string(result)
}
