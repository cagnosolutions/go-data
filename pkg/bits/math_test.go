package bits

import (
	"fmt"
	"testing"
)

const debug = false

func TestAdd(t *testing.T) {
	for _, e := range entrySet {
		c := addFn(e.a, e.b)
		v, ok := e.c[add]
		if !ok || c != v {
			t.Error("add: something went wrong")
		}
		if debug {
			fmt.Printf("add: %d + %d = %d\n", e.a, e.b, v)
		}
	}
}

func TestBitAdd(t *testing.T) {
	for _, e := range entrySet {
		c := BiAdd(e.a, e.b)
		v, ok := e.c[add]
		if !ok || c != v {
			t.Error("binary add: something went wrong")
		}
		if debug {
			fmt.Printf("binary add: %d + %d = %d\n", e.a, e.b, v)
		}
	}
}

func TestSub(t *testing.T) {
	for _, e := range entrySet {
		c := subFn(e.a, e.b)
		v, ok := e.c[sub]
		if !ok || c != v {
			t.Error("sub: something went wrong")
		}
		if debug {
			fmt.Printf("sub: %d - %d = %d\n", e.a, e.b, v)
		}
	}
}

func TestBitSub(t *testing.T) {
	for _, e := range entrySet {
		c := BiSub(e.a, e.b)
		v, ok := e.c[sub]
		if !ok || c != v {
			t.Error("binary sub: something went wrong")
		}
		if debug {
			fmt.Printf("binary sub: %d - %d = %d\n", e.a, e.b, v)
		}
	}
}

func TestMul(t *testing.T) {
	for _, e := range entrySet {
		c := mulFn(e.a, e.b)
		v, ok := e.c[mul]
		if !ok || c != v {
			t.Error("mul: something went wrong")
		}
		if debug {
			fmt.Printf("mul: %d * %d = %d\n", e.a, e.b, v)
		}
	}
}

func TestBitMul(t *testing.T) {
	for _, e := range entrySet {
		c := BiMul(e.a, e.b)
		v, ok := e.c[mul]
		if !ok || c != v {
			t.Errorf("binary mul: something went wrong (want=%d, got=%d)\n", v, c)
		}
		if debug {
			fmt.Printf("binary mul: %d * %d = %d\n", e.a, e.b, v)
		}
	}
}

func TestDiv(t *testing.T) {
	for _, e := range entrySet {
		c := divFn(e.a, e.b)
		v, ok := e.c[div]
		if !ok || c != v {
			t.Error("div: something went wrong")
		}
		if debug {
			fmt.Printf("div: %d / %d = %d\n", e.a, e.b, v)
		}
	}
}

func TestBitDiv(t *testing.T) {
	for _, e := range entrySet {
		c := BiDiv(e.a, e.b)
		v, ok := e.c[div]
		if !ok || c != v {
			t.Error("binary div: something went wrong")
		}
		if debug {
			fmt.Printf("binary div: %d / %d = %d\n", e.a, e.b, v)
		}
	}
}

func BenchmarkAdd(b *testing.B) {
	// b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, e := range entrySet {
			c := addFn(e.a, e.b)
			v, ok := e.c[add]
			if !ok || c != v {
				b.Error("add: something went wrong")
			}
		}
	}
}

func BenchmarkBinaryAdd(b *testing.B) {
	// b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, e := range entrySet {
			c := BiAdd(e.a, e.b)
			v, ok := e.c[add]
			if !ok || c != v {
				b.Error("binary add: something went wrong")
			}
		}
	}
}

func BenchmarkSub(b *testing.B) {
	// b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, e := range entrySet {
			c := subFn(e.a, e.b)
			v, ok := e.c[sub]
			if !ok || c != v {
				b.Error("sub: something went wrong")
			}
		}
	}
}

func BenchmarkBinarySub(b *testing.B) {
	// b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, e := range entrySet {
			c := BiSub(e.a, e.b)
			v, ok := e.c[sub]
			if !ok || c != v {
				b.Error("binary sub: something went wrong")
			}
		}
	}
}

func BenchmarkMul(b *testing.B) {
	// b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, e := range entrySet {
			c := mulFn(e.a, e.b)
			v, ok := e.c[mul]
			if !ok || c != v {
				b.Error("mul: something went wrong")
			}
		}
	}

}

func BenchmarkBinaryMul(b *testing.B) {
	// b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, e := range entrySet {
			c := BiMul(e.a, e.b)
			v, ok := e.c[mul]
			if !ok || c != v {
				b.Error("binary mul: something went wrong")
			}
		}
	}
}

func BenchmarkDiv(b *testing.B) {
	// b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, e := range entrySet {
			c := divFn(e.a, e.b)
			v, ok := e.c[div]
			if !ok || c != v {
				b.Error("div: something went wrong")
			}
		}
	}
}

func BenchmarkBinaryDiv(b *testing.B) {
	// b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, e := range entrySet {
			c := BiDiv(e.a, e.b)
			v, ok := e.c[div]
			if !ok || c != v {
				b.Error("binary div: something went wrong")
			}
		}
	}
}

// EVERYTHING BELOW THIS IS TEST DATA

const add, sub, mul, div int = 0x0a, 0x0b, 0x0c, 0x0d

type ans map[int]uint

type numEntry struct {
	a, b uint
	c    ans
}

func newNumEntry(a, b uint) numEntry {
	return numEntry{
		a: a, b: b, c: ans{
			add: a + b,
			sub: a - b,
			mul: a * b,
			div: a / b,
		},
	}
}

var entrySet = []numEntry{
	newNumEntry(4, 2),
	newNumEntry(20, 10),
	newNumEntry(15, 5),
	newNumEntry(6, 3),
	newNumEntry(8, 4),
	newNumEntry(45, 15),
	newNumEntry(2, 1),
	newNumEntry(2, 2),
	newNumEntry(36, 18),
}

func addFn(a, b uint) uint {
	return a + b
}

func subFn(a, b uint) uint {
	return a - b
}

func mulFn(a, b uint) uint {
	return a * b
}

func divFn(a, b uint) uint {
	return a / b
}
