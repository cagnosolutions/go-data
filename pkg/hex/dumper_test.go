package hex

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"testing"
)

var testData = `This line is A+
AAAAAAAAAAAAAAA
AAAAAAAAAAAAAAA
AAAAAAAAAAAAAAA
AAAAAAAAAAAAAAA

This is a test. This is only a test. Please do not be alarmed...

Further instructions to follow, but in the meantime--please
try to stay calm. We will contact you at 1345 hours, not
before, and not after. 

Thank you, and have a good day!
~You know who we are
`

var benchmarks = []struct {
	name string
	fn   func(s string) (string, error)
}{
	{"stdLibDumper", stdLibDumper},
	{"dumperV1", dumperV1},
	{"dumperV2", dumperV2},
	{"Control", func(s string) (string, error) { return "", nil }},
}

func stdLibDumper(s string) (string, error) {
	var sb strings.Builder
	d := hex.Dumper(&sb)
	_, err := d.Write([]byte(s))
	if err != nil {
		return "", err
	}
	err = d.Close()
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

func dumperV1(s string) (string, error) {
	var sb strings.Builder
	err := encode([]byte(s), &sb)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

func dumperV2(s string) (string, error) {
	var sb strings.Builder
	err := encodev2([]byte(s), &sb)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

func BenchmarkDumperVsStdLib(b *testing.B) {
	for _, tt := range benchmarks {
		b.Run(
			tt.name, func(b *testing.B) {
				b.ResetTimer()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					_, err := tt.fn(testData)
					if err != nil {
						b.Errorf("%s failed: %s", tt.name, err)
					}
				}
			},
		)
	}
}

func TestDumper_Encode(t *testing.T) {

	data := []byte("Wikipedia, the free encyclopedia that anyone can edit")

	rangeBy(
		16, data, func(b []byte, pad int) {
			fmt.Printf("%q (padding=%d)\n", b, pad)
		},
	)

}

func rangeBy(n int, data []byte, fn func(b []byte, pad int)) {
	var i int
	var size = len(data)
	for i = 0; i < size; i += n {
		if i+n > size {
			fn(data[i:], (i+n)-size)
			break
		}
		fn(data[i:i+n], 0)
	}
}

func BenchmarkRangeBy(b *testing.B) {
	bench := []struct {
		name string
		fn   func(n int, data []byte, fn func(b []byte, pad int))
		n    int
	}{
		{"rangeBy1", rangeBy, 1},
		{"rangeBy2", rangeBy, 2},
		{"rangeBy4", rangeBy, 4},
		{"rangeBy8", rangeBy, 8},
		{"rangeBy16", rangeBy, 16},
		{"rangeBy32", rangeBy, 32},
	}

	data := []byte(testData)

	for _, tt := range bench {
		b.Run(
			tt.name, func(b *testing.B) {
				b.ResetTimer()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					tt.fn(
						tt.n, data, func(p []byte, pad int) {
							if p == nil || pad < 0 {
								b.Errorf("%s failed for some reason\n", tt.name)
							}
						},
					)
				}
			},
		)
	}
}

func encode(data []byte, w io.Writer) error {
	var n int
	buf := make([]byte, 83)

	var offset int

	for offset <= len(data) {

		// encode offset
		buf[0] = byte(offset >> 24)
		buf[1] = byte(offset >> 16)
		buf[2] = byte(offset >> 8)
		buf[3] = byte(offset)
		hex.Encode(buf[4:], buf[:4])
		buf[12] = ' '
		buf[13] = ' '
		n += 14

		// encode next 16 bytes into two rows
		// of 8 bytes encoded as hex
		for i := range data[offset : offset+16] {
			hex.Encode(buf[n:], data[offset+i:offset+i+1])
			n += 2
			buf[n] = ' '
			n++
			if i == 7 {
				// add an extra space
				buf[n] = ' '
				n++
			}
			if i == 15 {
				buf[n] = ' '
				n++
				break
			}
		}

		// write ascii
		buf[n] = '|'
		n++
		for _, c := range data[offset : offset+16] {
			buf[n] = toChar(c)
			n++
		}
		buf[n] = '|'
		n++
		buf[n] = '\n'

		_, err := w.Write(buf[4:])
		if err != nil {
			return err
		}

		offset += 16
		n = 0
	}
	return nil
}

func encodev2(data []byte, w io.Writer) error {
	var n int
	buf := make([]byte, 24)
	same := []byte{'*', '\n'}
	var err error

	var bl int

	for bl <= len(data) {
		// check to see if previous line
		// was the same as the current one
		if bl > 0 && bytes.Equal(data[bl-16:bl], data[bl:bl+16]) {
			_, err = w.Write(same)
			if err != nil {
				return err
			}
			bl += 16
			n = 0
			continue
		}
		// encode offset
		buf[0] = byte(bl >> 24)
		buf[1] = byte(bl >> 16)
		buf[2] = byte(bl >> 8)
		buf[3] = byte(bl)
		hex.Encode(buf[4:], buf[:4])
		buf[12] = ' '
		buf[13] = ' '
		_, err = w.Write(buf[4:])
		if err != nil {
			return err
		}

		// encode next 16 bytes into two rows
		// of 8 bytes encoded as hex
		n = 14
		for i := range data[bl : bl+16] {
			hex.Encode(buf[n:], data[bl+i:bl+i+1])
			n += 2
			buf[n] = ' '
			n++
			if i == 7 {
				// add an extra space
				buf[n] = ' '
				n++
			}
			if i == 15 {
				buf[n] = ' '
				n++
				break
			}
		}

		// write ascii
		buf[n] = '|'
		n++
		for _, c := range data[bl : bl+16] {
			buf[n] = toChar(c)
			n++
		}
		buf[n] = '|'
		n++
		buf[n] = '\n'

		_, err = w.Write(buf[4:])
		if err != nil {
			return err
		}

		bl += 16
		n = 0
	}

	// fmt.Printf("%s\n", buf)
	// fmt.Printf("offset: %q\n", buf[0:8])
	// fmt.Printf("2 spaces: %q\n", buf[8:10])
	// fmt.Printf("16 byte hex: %q\n", buf[10:33])
	// fmt.Printf("2 spaces: %q\n", buf[33:35])
	// fmt.Printf("16 byte hex: %q\n", buf[35:58])
	// fmt.Printf("2 spaces: %q\n", buf[58:60])
	// fmt.Printf("ascii text: %q\n", buf[60:78])
	// fmt.Printf("newline at end: %q\n", buf[79])
	return nil
}

func TestEncode(t *testing.T) {
	var sb strings.Builder
	err := encode([]byte(testData), &sb)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(sb.String())
}

func TestDumper_EncodeAddr(t *testing.T) {

	offset := 16
	// data := []byte("0123456789ABCDEFzzz")
	data := []byte(testData)

	var n int
	buf := make([]byte, 84)

	// encode offset
	buf[0] = byte(offset >> 24)
	buf[1] = byte(offset >> 16)
	buf[2] = byte(offset >> 8)
	buf[3] = byte(offset)
	hex.Encode(buf[4:], buf[:4])
	buf[12] = ' '
	buf[13] = ' '
	n += 14

	// encode next 16 bytes into two rows
	// of 8 bytes encoded as hex
	for i := range data {
		hex.Encode(buf[n:], data[i:i+1])
		n += 2
		buf[n] = ' '
		n++
		if i == 7 {
			// add an extra space
			buf[n] = ' '
			n++
		}
		if i == 15 {
			buf[n] = ' '
			n++
			break
		}
	}

	// write ascii
	buf[n] = '|'
	n++
	for i := range data[:16] {
		buf[n] = data[i]
		n++
	}
	buf[n] = '|'
	n++
	buf[n] = '\n'

	buf = buf[4:]

	// fmt.Printf("%s\n", buf)
	// fmt.Printf("offset: %q\n", buf[0:8])
	// fmt.Printf("2 spaces: %q\n", buf[8:10])
	// fmt.Printf("16 byte hex: %q\n", buf[10:33])
	// fmt.Printf("2 spaces: %q\n", buf[33:35])
	// fmt.Printf("16 byte hex: %q\n", buf[35:58])
	// fmt.Printf("2 spaces: %q\n", buf[58:60])
	// fmt.Printf("ascii text: %q\n", buf[60:78])
	// fmt.Printf("newline at end: %q\n", buf[79])
}

func TestDumper(t *testing.T) {
	var sb strings.Builder
	d := NewDumper(&sb)
	d.Write([]byte("aaaaaaaaaaaaaaaa"))
	d.Write([]byte("bbbbbbbbbbbbbbbb"))
	d.Write([]byte("cccccccccccccccc"))
	fmt.Println(sb.String())
}

func TestDumper_Colors(t *testing.T) {
	s := []byte("abcdefghijklmnopqrstuvwxyz 0123456789 ABCDEFGHIJKLMNOPQRSTUVWXYZ \t\t\nThis is just a test. \x00")
	for i := 0; i < len(s); i++ {
		colorchar(s[i])
		fmt.Printf("%c", s[i])
	}
	fmt.Println()
	for i := 0; i < len(s); i++ {
		colorchar(s[i])
		fmt.Printf("%x", s[i])
	}
	for i := 0; i < len(s); i++ {
		colorchar(s[i])
		fmt.Printf("char=%q, dec=%d, hex=%.2x\n", s[i], s[i], s[i])
	}
}

func TestDumper_IsPunct(t *testing.T) {
	fmt.Printf("All punctuation characters:\n")
	for i := 0; i < 256; i++ {
		if ispunct(byte(i)) {
			fmt.Printf("%c ", i)
		}
	}
}

func TestDumper_IsAlpha(t *testing.T) {
	fmt.Printf("All alphabetic characters:\n")
	for i := 0; i < 256; i++ {
		if isalpha(byte(i)) {
			fmt.Printf("%c ", i)
		}
	}
}

func TestDumper_IsDigit(t *testing.T) {
	fmt.Printf("All alphabetic characters:\n")
	for i := 0; i < 256; i++ {
		if isdigit(byte(i)) {
			fmt.Printf("%c ", i)
		}
	}
}

func TestDumper_IsPrint(t *testing.T) {
	fmt.Printf("All printable characters:\n")
	for i := 0; i < 256; i++ {
		if isprint(byte(i)) {
			fmt.Printf("%c ", i)
		}
	}
}

func TestDumper_IsSpace(t *testing.T) {
	fmt.Printf("All space characters:\n")
	for i := 0; i < 256; i++ {
		if isspace(byte(i)) {
			fmt.Printf("%c (dec=%d, hex=%x)\n", i, i, i)
		}
	}
}
