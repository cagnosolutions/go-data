package bytes

import (
	"testing"
	"unicode"
)

func tenBytes(r byte) []byte {
	bs := make([]byte, 10)
	for i := range bs {
		bs[i] = byte(r)
	}
	return []byte(string(bs))
}

// User-defined self-inverse mapping function
func rot13Bytes(r byte) (byte, bool) {
	const step = 13
	if r >= 'a' && r <= 'z' {
		return ((r - 'a' + step) % 26) + 'a', true
	}
	if r >= 'A' && r <= 'Z' {
		return ((r - 'A' + step) % 26) + 'A', true
	}
	return r, true
}

func TestMap(t *testing.T) {
	// Run a couple of awful growth/shrinkage tests
	a := string(tenBytes('a'))

	// 1.  Grow. This triggers two re-allocations in Map.
	maxRune := func(r byte) (byte, bool) {
		return unicode.MaxASCII, true
	}
	m := Map(maxRune, []byte(a))
	expect := string(tenBytes(unicode.MaxASCII))
	if string(m) != expect {
		t.Errorf("growing: expected %q got %q", expect, m)
	}

	// 2. Shrink
	minRune := func(r byte) (byte, bool) { return 'a', true }
	m = Map(minRune, []byte(tenBytes(unicode.MaxASCII)))
	expect = a
	if string(m) != expect {
		t.Errorf("shrinking: expected %q got %q", expect, m)
	}

	// 3. Rot13
	m = Map(rot13Bytes, []byte("a to zed"))
	expect = "n gb mrq"
	if string(m) != expect {
		t.Errorf("rot13: expected %q got %q", expect, m)
	}

	// 4. Rot13^2
	m = Map(rot13Bytes, Map(rot13Bytes, []byte("a to zed")))
	expect = "a to zed"
	if string(m) != expect {
		t.Errorf("rot13: expected %q got %q", expect, m)
	}

	// 5. Drop
	dropNotLatin := func(r byte) (byte, bool) {
		if 0x41 <= r && r <= 0x7a {
			return r, true
		}
		return 0, false
		// if unicode.Is(unicode.Latin, rune(r)) {
		//	return r, true
		// }
		// return 0, false
	}
	m = Map(dropNotLatin, []byte("Hello, 세계"))
	expect = "Hello"
	if string(m) != expect {
		t.Errorf("drop: expected %q got %q", expect, m)
	}

	// 6. Invalid rune
	invalidRune := func(r byte) (byte, bool) {
		return unicode.MaxASCII + 1, true
	}
	m = Map(invalidRune, []byte("x"))
	expect = "\x80"
	if string(m) != expect {
		t.Errorf("invalidRune: expected %q got %q", expect, m)
	}
}
