package dbms

import (
	"fmt"
	"path/filepath"
	"testing"
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
	// write to file
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
	// read from file
	err = bs.ReadFile(filepath.Join(basePath, "dat-current.idx"))
	if err != nil {
		t.Error(err)
	}
	for j := range bs {
		if (*bs)[j] != want[j] {
			t.Error("bitset failed to read file back in correctly")
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
