package pager

import (
	"fmt"
	"testing"
)

const size = 128

func TestBitmap_Size(t *testing.T) {
	// create new bitmap
	bm := newBitmap(size)
	// check the bitmap size
	if bm.length != size {
		t.Error("bitmap is the wrong size")
	}

	bm2 := newBitmap(32768)
	if bm2.length != 32768 {
		t.Error("bitmap is the wrong size")
	}
	if len(bm2.bits) != 512 {
		t.Error("bitmap is the wrong size")
	}
}

func TestBitmap_Has(t *testing.T) {
	// create new bitmap
	bm := newBitmap(size)
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

func TestBitmap_ToBytes(t *testing.T) {
	// create new bitmap
	bm := newBitmap(size)
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
	bm2 := newBitmap(size * 3)
	fmt.Println("before:", bm2)
	// print out
	n = bm2.read(b)
	fmt.Printf("read %d bytes\n", n)
	fmt.Println("after:", bm2)
}

func TestBitmap_FindFirst(t *testing.T) {
	// create new bitmap
	bm := newBitmap(size)
	// set the first 72 bits
	for i := 0; i < size; i++ {
		// except for bit 32 and 56
		if i < 72 && i != 32 && i != 56 {
			bm.set(uint(i))
		}
	}
	fmt.Println(bm)
	// find first
	at := bm.first()
	fmt.Printf("free space found at: %d\n", at)
	// set the found one
	bm.set(uint(at))
	fmt.Println(bm)
	// find the next free one
	at = bm.first()
	fmt.Printf("free space found at: %d\n", at)
	// set the found one
	bm.set(uint(at))
	fmt.Println(bm)
	// find the next free one
	at = bm.first()
	fmt.Printf("free space found at: %d\n", at)
	fmt.Println(bm)
}
