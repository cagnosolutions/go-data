package dbms

import (
	"fmt"
	"math/big"
	"math/bits"
	"math/rand"
	"path/filepath"
	"testing"
	"time"
)

func TestBitsetIndex_Clear(t *testing.T) {
	bs := NewBitsetIndex()
	want := make([]uint64, bitsetSize)
	copy(want, bs[:])
	for i := 0; i < bs.Bits(); i += 8 {
		bs.SetBit(uint(i))
	}
	for j := range bs {
		if (*bs)[j] == want[j] {
			t.Error("bitset failed to populate")
		}
	}
	bs.Clear()
	for j := range bs {
		if (*bs)[j] != want[j] {
			t.Error("bitset failed to clear")
		}
	}
}

func TestBitsetIndex_ReadWrite(t *testing.T) {
	// make new bitset
	bs := NewBitsetIndex()
	for i := 0; i < bs.Bits(); i += 2 {
		// populate
		bs.SetBit(uint(i))
	}
	// make our thing to test against
	want := make([]uint64, bitsetSize)
	copy(want, bs[:])
	// check bitset population
	for j := range bs {
		if (*bs)[j] != want[j] {
			t.Error("bitset failed to populate")
		}
	}
	// write to current
	err := bs.WriteFile(filepath.Join(basePath, "dat-current.idx"))
	if err != nil {
		t.Error(err)
	}
	// clear bitset, and check clear
	bs.Clear()
	for j := range bs {
		if (*bs)[j] == want[j] {
			t.Error("bitset failed to clear")
		}
	}
	// read from current
	err = bs.ReadFile(filepath.Join(basePath, "dat-current.idx"))
	if err != nil {
		t.Error(err)
	}
	for j := range bs {
		if (*bs)[j] != want[j] {
			t.Error("bitset failed to read current back in correctly")
		}
	}
	fmt.Println(bs)
}

func TestBitsetIndex(t *testing.T) {
	bs := NewBitsetIndex()
	fmt.Println(bs)
	for i := uint(0); i < 32; i++ {
		bs.SetBit(i)
	}
	fmt.Println(bs)
	for i := uint(32); i < 64; i++ {
		bs.SetBit(i)
	}
	fmt.Println(bs)
	for i := uint(64); i < 92; i++ {
		bs.SetBit(i)
	}
	fmt.Println(bs)
	for i := uint(92); i < 128; i++ {
		bs.SetBit(i)
	}
	fmt.Println(bs)
	for i := uint(768); i < 1024; i++ {
		bs.SetBit(i)
	}
	fmt.Println(bs)
}

func BenchmarkBitsetIndex_GetFree0(b *testing.B) {
	bs := NewBitsetIndex()
	// for i := 0; i < 16; i++ {
	// 	(*bs)[i] = ^uint64(0)
	// }
	for i := uint(0); i < 1024; i++ {
		bs.SetBit(i)
	}
	bs.UnsetBit(1023)
	// fmt.Println(bs)
	for i := 0; i < b.N; i++ {
		var free int
		for j := 0; j < 1024; j++ {
			if !bs.HasBit(uint(j)) {
				free = j
				break
			}
		}
		if free != 1023 {
			b.Error("did not find the correct free bit")
		}
	}
}

func BenchmarkBitsetIndex_GetFree(b *testing.B) {
	bs := NewBitsetIndex()
	// for i := 0; i < 16; i++ {
	// 	(*bs)[i] = ^uint64(0)
	// }
	for i := uint(0); i < 1024; i++ {
		bs.SetBit(i)
	}
	bs.UnsetBit(1023)
	// fmt.Println(bs)
	for i := 0; i < b.N; i++ {
		free := bs.GetFree()
		if free != 1023 {
			b.Error("did not find the correct free bit")
		}
	}
}

func TestBitsetIndex_PageOffsetAfter(t *testing.T) {
	bs := NewBitsetIndex()
	for i := uint(0); i < 1024; i++ {
		bs.SetBit(i)
	}
	bs.UnsetBit(17)
	for i := uint(767); i < 900; i++ {
		if i%2 == 0 {
			bs.UnsetBit(i)
		}
	}
	bs.UnsetBit(900)
	bs.UnsetBit(905)
	bs.UnsetBit(908)
	after := 0
	n := bs.PageOffsetAfter(after)
	fmt.Printf("page offset after %d is %d\n", after, n)
	after = n

	n = bs.PageOffsetAfter(after)
	fmt.Printf("page offset after %d is %d\n", after, n)
	after = n

	n = bs.PageOffsetAfter(after)
	fmt.Printf("page offset after %d is %d\n", after, n)
	after = n

	n = bs.PageOffsetAfter(after)
	fmt.Printf("page offset after %d is %d\n", after, n)
	after = n

	n = bs.PageOffsetAfter(after)
	fmt.Printf("page offset after %d is %d\n", after, n)
	after = n

	n = bs.PageOffsetAfter(after)
	fmt.Printf("page offset after %d is %d\n", after, n)
	after = n

	n = bs.PageOffsetAfter(after)
	fmt.Printf("page offset after %d is %d\n", after, n)
	after = n

	n = bs.PageOffsetAfter(after)
	fmt.Printf("page offset after %d is %d\n", after, n)
	after = n

	n = bs.PageOffsetAfter(after)
	fmt.Printf("page offset after %d is %d\n", after, n)
	after = n

	n = bs.PageOffsetAfter(after)
	fmt.Printf("page offset after %d is %d\n", after, n)
	after = n

}

func BenchmarkBitsetIndex_GetBit(b *testing.B) {
	bs := NewBitsetIndex()
	bs.SetAll()
	b.ResetTimer()
	var bit uint64
	for i := 0; i < b.N; i++ {
		for j := uint(0); j < 1024; j++ {
			bit = bs.GetBit(j)
			fmt.Println(bit)
			if bit == 0 {
				b.Error("something happened")
			}
		}
	}
	_ = bit
}

func BenchmarkBitsetIndex_FindBit(b *testing.B) {
	bs := NewBitsetIndex()
	bs.SetAll()
	b.ResetTimer()
	var slot, index uint64
	for i := 0; i < b.N; i++ {
		for j := uint(0); j < 1024; j++ {
			bit := bs.FindBit(j)
			_ = bit
			// fmt.Printf("slot=%d, index=%d\n", slot, index)
		}
	}
	_ = slot
	_ = index
}

func BenchmarkBitsetIndex_Range(b *testing.B) {
	bs := NewBitsetIndex()
	bs.SetAll()
	var count int
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bs.Range(
			64, 512, func(index int, bit uint64) bool {
				if bit == 0 {
					count = 1
					return false
				}
				return true
			},
		)
		if count != 0 {
			b.Error("something went wrong")
		}
		count = 0
	}
	_ = count
}

func TestBitsetIndex_Loop(t *testing.T) {
	bs := NewBitsetIndex()
	bs.SetAll()
	fmt.Println(bs)
	bs.UnsetBit(65)
	bs.UnsetBit(129)
	bs.UnsetBit(257)
	bs.UnsetBit(385)
	bs.Range(
		64, 512, func(index int, bit uint64) bool {
			fmt.Printf(">>> bit[%d]=%d\n", index, bit)
			return true
		},
	)

	// for i := beg; i < end; i++ {
	// 	fmt.Printf("beg=%d, end=%d, find=%d, mask=%d, i=%d\n", beg, end, find, mask, i)
	// 	if bs[i] > 0 {
	// 		find |= mask
	// 	}
	// 	mask <<= 1
	// }
}

func TestBitsetIndex_Info(t *testing.T) {
	bs := NewBitsetIndex()
	for i := uint(0); i < 768; i++ {
		bs.SetBit(i)
	}
	bs.UnsetBit(17)
	bs.SetBit(900)
	// test out our info
	bi := bs.Info()
	fmt.Println(bi)
}

func BenchmarkBitsetIndex_Info(b *testing.B) {
	bs := NewBitsetIndex()
	for i := uint(0); i < 768; i++ {
		bs.SetBit(i)
	}
	bs.UnsetBit(17)
	bs.SetBit(900)
	var r any
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bi := bs.Info()
		r = bi
	}
	_ = r
}

var (
	testInt = genBigInt(0, 8<<10)
)

func TestTestInts(t *testing.T) {
	fmt.Printf("%#v\n", testInt)
	fmt.Printf("%#v\n", genBigInt(time.Now().UnixNano(), 8<<10))
}

func genBigInt(seed int64, maxlen int) *big.Int {
	src := rand.NewSource(seed)
	rng := rand.New(src)
	l := rng.Intn(maxlen)
	b := make([]byte, l+3)
	for i := 0; i < l; i += 4 {
		x := rng.Int63()
		b[i+0] = byte((x >> 0) & 0xff)
		b[i+1] = byte((x >> 8) & 0xff)
		b[i+2] = byte((x >> 16) & 0xff)
		b[i+3] = byte((x >> 24) & 0xff)
	}
	return big.NewInt(0).SetBytes(b[:l])
}

func BitCountBitsOnesCount64(n *big.Int) int {
	count := 0
	for _, v := range n.Bits() {
		count += bits.OnesCount64(uint64(v))
	}
	return count
}

// types and constants used in the functions below
// uint64_t is an unsigned 64-bit integer variable type (defined in C99 version of C language)
const (
	m1  uint64 = 0x5555555555555555 // binary: 0101...
	m2  uint64 = 0x3333333333333333 // binary: 00110011..
	m4  uint64 = 0x0f0f0f0f0f0f0f0f // binary:  4 zeros,  4 ones ...
	m8  uint64 = 0x00ff00ff00ff00ff // binary:  8 zeros,  8 ones ...
	m16 uint64 = 0x0000ffff0000ffff // binary: 16 zeros, 16 ones ...
	m32 uint64 = 0x00000000ffffffff // binary: 32 zeros, 32 ones
	h01 uint64 = 0x0101010101010101 // the sum of 256 to the power of 0,1,2,3...
)

// This is a naive implementation, shown for comparison,
// and to help in understanding the better functions.
// This algorithm uses 24 arithmetic operations (shift, add, and).
func popcount64a(x uint64) int {
	x = (x & m1) + ((x >> 1) & m1)    // put count of each  2 bits into those  2 bits
	x = (x & m2) + ((x >> 2) & m2)    // put count of each  4 bits into those  4 bits
	x = (x & m4) + ((x >> 4) & m4)    // put count of each  8 bits into those  8 bits
	x = (x & m8) + ((x >> 8) & m8)    // put count of each 16 bits into those 16 bits
	x = (x & m16) + ((x >> 16) & m16) // put count of each 32 bits into those 32 bits
	x = (x & m32) + ((x >> 32) & m32) // put count of each 64 bits into those 64 bits
	return int(x)
}

func BitCountPop64a(n *big.Int) int {
	count := 0
	for _, v := range n.Bits() {
		count += popcount64a(uint64(v))
	}
	return count
}

// This uses fewer arithmetic operations than any other known
// implementation on machines with slow multiplication.
// This algorithm uses 17 arithmetic operations.
func popcount64b(x uint64) int {
	x -= (x >> 1) & m1             // put count of each 2 bits into those 2 bits
	x = (x & m2) + ((x >> 2) & m2) // put count of each 4 bits into those 4 bits
	x = (x + (x >> 4)) & m4        // put count of each 8 bits into those 8 bits
	x += x >> 8                    // put count of each 16 bits into their lowest 8 bits
	x += x >> 16                   // put count of each 32 bits into their lowest 8 bits
	x += x >> 32                   // put count of each 64 bits into their lowest 8 bits
	return int(x & 0x7f)
}

func BitCountPop64b(n *big.Int) int {
	count := 0
	for _, v := range n.Bits() {
		count += popcount64b(uint64(v))
	}
	return count
}

// This uses fewer arithmetic operations than any other known
// implementation on machines with fast multiplication.
// This algorithm uses 12 arithmetic operations, one of which is a multiply operation.
func popcount64c(x uint64) int {
	x -= (x >> 1) & m1             // put count of each 2 bits into those 2 bits
	x = (x & m2) + ((x >> 2) & m2) // put count of each 4 bits into those 4 bits
	x = (x + (x >> 4)) & m4        // put count of each 8 bits into those 8 bits
	return int((x * h01) >> 56)    // returns left 8 bits of x + (x<<8) + (x<<16) + (x<<24) + ...
}

func BitCountFast(n *big.Int) int {
	var count int
	for _, x := range n.Bits() {
		for x != 0 {
			x &= x - 1
			count++
		}
	}
	return count
}

func BitCountPop64c(n *big.Int) int {
	count := 0
	for _, v := range n.Bits() {
		count += popcount64c(uint64(v))
	}
	return count
}

func benchBitCounter(b *testing.B, fn func(*big.Int) int) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = fn(testInt)
	}
}

func BenchmarkBitCounterOnesCount(b *testing.B) {
	b.ReportAllocs()
	benchBitCounter(b, BitCountBitsOnesCount64)
}

func BenchmarkBitCounterCountFast(b *testing.B) {
	b.ReportAllocs()
	benchBitCounter(b, BitCountFast)
}

func BenchmarkBitCounterPopcount64A(b *testing.B) {
	b.ReportAllocs()
	benchBitCounter(b, BitCountPop64a)
}

func BenchmarkBitCounterPopcount64B(b *testing.B) {
	b.ReportAllocs()
	benchBitCounter(b, BitCountPop64b)
}

func BenchmarkBitCounterPopcount64C(b *testing.B) {
	b.ReportAllocs()
	benchBitCounter(b, BitCountPop64c)
}
