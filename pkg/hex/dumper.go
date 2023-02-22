package hex

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// 00000000  30 31 32 33 34 35 36 37  38 39 41 42 43 44 45 46  |0123456789ABCDEF|
// 00000010  2e 2f 30 31 32 33 34 35  36 37 38 39 3a 3b 3c 3d  |./0123456789:;<=|
// 00000010  30 31 32 33 34 35 36 37  38 39 41 42 43 44 45 46  |
// 00000010  30 31 32 33 34 35 36 37  38 39 41 42 43 44 45 46  |0123456789ABCDEF|
// ^ offset                          ^ extra space              ^ ASCII of line.

// Dumper returns a WriteCloser that writes a hex dump of all written data to
// w. The format of the dump matches the output of `hexdump -C` on the command
// line.
func NewDumper(w io.Writer) io.WriteCloser {
	return &dumper{w: w}
}

type dumper struct {
	w          io.Writer
	rightChars [18]byte
	buf        [14]byte
	used       int  // number of bytes in the current line
	n          uint // number of bytes, total
	closed     bool
}

func toChar(b byte) byte {
	if b < 32 || b > 126 {
		return '.'
	}
	return b
}

func (h *dumper) Write(data []byte) (n int, err error) {
	if h.closed {
		return 0, errors.New("encoding/hex: dumper closed")
	}

	// Output lines look like:
	// 00000010  2e 2f 30 31 32 33 34 35  36 37 38 39 3a 3b 3c 3d  |./0123456789:;<=|
	// ^ offset                          ^ extra space              ^ ASCII of line.
	for i := range data {
		if h.used == 0 {
			// At the beginning of a line we print the current
			// offset in hex.
			h.buf[0] = byte(h.n >> 24)
			h.buf[1] = byte(h.n >> 16)
			h.buf[2] = byte(h.n >> 8)
			h.buf[3] = byte(h.n)
			hex.Encode(h.buf[4:], h.buf[:4])
			h.buf[12] = ' '
			h.buf[13] = ' '
			_, err = h.w.Write(h.buf[4:])
			if err != nil {
				return
			}
		}
		hex.Encode(h.buf[:], data[i:i+1])
		h.buf[2] = ' '
		l := 3
		if h.used == 7 {
			// There's an additional space after the 8th byte.
			h.buf[3] = ' '
			l = 4
		} else if h.used == 15 {
			// At the end of the line there's an extra space and
			// the bar for the right column.
			h.buf[3] = ' '
			h.buf[4] = '|'
			l = 5
		}
		_, err = h.w.Write(h.buf[:l])
		if err != nil {
			return
		}
		n++
		h.rightChars[h.used] = toChar(data[i])
		h.used++
		h.n++
		if h.used == 16 {
			h.rightChars[16] = '|'
			h.rightChars[17] = '\n'
			_, err = h.w.Write(h.rightChars[:])
			if err != nil {
				return
			}
			h.used = 0
		}
	}
	return
}

func (h *dumper) Close() (err error) {
	// See the comments in Write() for the details of this format.
	if h.closed {
		return
	}
	h.closed = true
	if h.used == 0 {
		return
	}
	h.buf[0] = ' '
	h.buf[1] = ' '
	h.buf[2] = ' '
	h.buf[3] = ' '
	h.buf[4] = '|'
	nBytes := h.used
	for h.used < 16 {
		l := 3
		if h.used == 7 {
			l = 4
		} else if h.used == 15 {
			l = 5
		}
		_, err = h.w.Write(h.buf[:l])
		if err != nil {
			return
		}
		h.used++
	}
	h.rightChars[nBytes] = '|'
	h.rightChars[nBytes+1] = '\n'
	_, err = h.w.Write(h.rightChars[:nBytes+2])
	return
}

func writeColorHex(w io.Writer, c byte) {
	fmt.Fprintf(w, "%s%.2X%s ", colorchar2(c), c, reset)
}

func writeColorChar(w io.Writer, c byte) {
	fmt.Fprintf(w, "%s%c%s", colorchar2(c), c, reset)
}

func colorchar2(c byte) string {
	switch {
	case c == 0x00:
		return grey
	case isalpha(c):
		return orange
	case isdigit(c):
		return yellow
	case !isprint(c):
		return red
	case isspace(c):
		return green
	case ispunct(c):
		return pruple
	case c < 16:
		return "\033[0;32m"
	case c >= 16 && c <= 31:
		return "\033[1;35m"
	}
	return string(c)
}

func colorchar(c byte) {
	switch {
	case c == 0x00:
		fmt.Printf(grey)
	case isalpha(c):
		fmt.Printf(orange)
	case isdigit(c):
		fmt.Printf(yellow)
	case !isprint(c):
		fmt.Printf(red)
	case isspace(c):
		fmt.Printf(green)
	case ispunct(c):
		fmt.Printf(pruple)
	case c < 16:
		fmt.Printf("\033[0;32m")
	case c >= 16 && c <= 31:
		fmt.Printf("\033[1;35m")
	}
}

func resetcolor() {
	fmt.Printf(reset)
}

func isalpha(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

func ispunct(c byte) bool {
	return 33 <= c && c <= 47 || 58 <= c && c <= 64 || 91 <= c && c <= 96 || 123 <= c && c <= 126
}

func isdigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func isprint(c byte) bool {
	return 32 <= c && c <= 126
}

func isspace(c byte) bool {
	return 9 <= c && c <= 13 || c == 32
}
