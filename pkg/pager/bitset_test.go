package pager

import (
	"fmt"
	"testing"
)

const size = 64

func TestBitset_Size(t *testing.T) {
	// create new bitset
	bm := newBitset(size)
	// check the bitset fsize
	if bm.length != size {
		t.Error("bitset is the wrong fsize")
	}

	bm2 := newBitset(32768)
	if bm2.length != 32768 {
		t.Error("bitset is the wrong fsize")
	}
	if len(bm2.bits) != 512 {
		t.Error("bitset is the wrong fsize")
	}
	fmt.Println("bm2 size:", bm2.sizeof())
}

func TestBitset_Has(t *testing.T) {
	// create new bitset
	bm := newBitset(size)
	// set all the odds
	for i := 0; i < size; i++ {
		if i%2 == 0 {
			continue
		}
		bm.set(uint(i))
	}
	// check to see if all the odds are set
	for i := 0; i < size; i++ {
		if i%2 == 0 {
			if bm.has(uint(i)) {
				t.Error("even was set")
			}
			continue
		}
		if !bm.has(uint(i)) {
			t.Error("odd was not set")
		}
		bm.set(uint(i))
	}
}

var buf []byte

func TestBitset_WriteAndRead(t *testing.T) {
	// create new bitset
	bm := newBitset(size)
	// set all the odds
	for i := 0; i < size; i++ {
		if i%2 == 0 {
			continue
		}
		bm.set(uint(i))
	}
	// print out
	fmt.Println(bm)
	// dump to bytes
	b := make([]byte, size/8+1)
	n := bm.write(b)
	fmt.Printf("wrote %d bytes\n", n)
	// buf = bm.toBytes()
	// hex.Dump(buf)
	fmt.Printf("bytes:\n%b\n", b)

	// make a new one from the bytes
	// bm2 := newBitmapFromBytes(b)
	bm2 := newBitset(size * 3)
	fmt.Println("before:", bm2)
	// print out
	n = bm2.read(b)
	fmt.Printf("read %d bytes\n", n)
	fmt.Println("after:", bm2)
}

func TestBitset_FindFirst(t *testing.T) {
	// create new bitset
	bm := newBitset(size)
	// set the first 72 bits
	for i := 0; i < size; i++ {
		// except for bit 32 and 56
		if i < 72 && i != 32 && i != 56 {
			bm.set(uint(i))
		}
	}
	fmt.Println(bm)
	// find first
	at := bm.free()
	fmt.Printf("free space found at: %d\n", at)
	// set the found one
	bm.set(uint(at))
	fmt.Println(bm)
	// find the next free one
	at = bm.free()
	fmt.Printf("free space found at: %d\n", at)
	// set the found one
	bm.set(uint(at))
	fmt.Println(bm)
	// find the next free one
	at = bm.free()
	fmt.Printf("free space found at: %d\n", at)
	fmt.Println(bm)
}

func TestBitset_Resize(t *testing.T) {
	bm := newBitset(64)
	for i := 0; i < 128; i++ {
		bm.set(uint(i))
		if i%8 == 0 {
			fmt.Printf("setting %.3d\t%s\n", i, bm)
		}
	}
	bm.set(254)
	fmt.Println(bm)
}

func TestBitset_Aligns(t *testing.T) {
	for i := 0; i < 255; i += 16 {
		b := roundTo(i, 2)
		fmt.Printf("roundTo(%d, %d) produced %d\n", i, 64, b)
		c := alignedSize(uint64(i))
		fmt.Printf("alignedSize(%d) produced %d\n\n", i, c)
	}
}

func TestBitset_RealSize(t *testing.T) {
	bm := newBitset(127)
	fmt.Println(bm)
	fmt.Println("actual size in memory: ", bm.sizeof())
}