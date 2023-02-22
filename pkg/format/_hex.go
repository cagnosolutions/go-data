package format

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
	"log"
)

// from: https://cs.opensource.google/go/go/+/refs/tags/go1.20.1:src/encoding/hex/hex.go;l=224

type Dumper struct {
	w      io.Writer
	line   [16]byte
	right  [18]byte
	buf    [14]byte
	used   int  // number of bytes in the current line
	n      uint // number of bytes total
	closed bool
}

func NewDumper(w io.Writer) io.WriteCloser {
	return &Dumper{w: w}
}

func toChar(b byte) byte {
	if b < 32 || b > 126 {
		return '.'
	}
	return b
}

func (d *Dumper) Write(data []byte) (n int, err error) {
	if d.closed {
		return 0, errors.New("hex dumper closed")
	}
	// var same bool
	var beg, end uint
	// Output lines look like:
	// 00000010  2e 2f 30 31 32 33 34 35  36 37 38 39 3a 3b 3c 3d  |./0123456789:;<=|
	// ^ offset                          ^ extra space              ^ ASCII of line.
	for i := range data {
		if bytes.Equal(d.right[0:16], d.line[0:16]) {
			//		same = true
			log.Printf(">> (%d)[%d-%d] line=%q\n", d.n/16, beg, end, d.line)
		}
		if d.used == 0 {
			beg = d.n
		}
		if d.used == 0 {
			// At the beginning of a line we print the current
			// offset in hex.
			d.buf[0] = byte(d.n >> 24)
			d.buf[1] = byte(d.n >> 16)
			d.buf[2] = byte(d.n >> 8)
			d.buf[3] = byte(d.n)
			hex.Encode(d.buf[4:], d.buf[:4])
			d.buf[12] = ' '
			d.buf[13] = ' '
			_, err = d.w.Write(d.buf[4:])
			if err != nil {
				return
			}
		}
		hex.Encode(d.buf[:], data[i:i+1])
		d.buf[2] = ' '
		l := 3
		if d.used == 7 {
			// There's an additional space after the 8th byte.
			d.buf[3] = ' '
			l = 4
		} else if d.used == 15 {
			// At the end of the line there's an extra space and
			// the bar for the right column.
			d.buf[3] = ' '
			d.buf[4] = '|'
			l = 5
		}
		_, err = d.w.Write(d.buf[:l])
		if err != nil {
			return
		}
		n++
		d.right[d.used] = toChar(data[i])
		d.used++
		d.n++
		if d.used == 16 {
			end = d.n
		}
		if d.used == 16 {
			// if bytes.Equal(d.right[0:16], d.line[0:16]) {
			// 	same = true
			// 	log.Printf(">> (%d)[%d-%d] line=%q\n", d.n/16, beg, end, d.line)
			// }
			// copy(d.line[0:16], d.right[0:16])
		}
		if d.used == 16 {
			d.right[16] = '|'
			d.right[17] = '\n'
			_, err = d.w.Write(d.right[:])
			if err != nil {
				return
			}
			d.used = 0
		}
	}
	return
}

func (d *Dumper) Close() (err error) {
	// See the comments in Write() for the details of this format.
	if d.closed {
		return
	}
	d.closed = true
	if d.used == 0 {
		return
	}
	d.buf[0] = ' '
	d.buf[1] = ' '
	d.buf[2] = ' '
	d.buf[3] = ' '
	d.buf[4] = '|'
	nBytes := d.used
	for d.used < 16 {
		l := 3
		if d.used == 7 {
			l = 4
		} else if d.used == 15 {
			l = 5
		}
		_, err = d.w.Write(d.buf[:l])
		if err != nil {
			return
		}
		d.used++
	}
	d.right[nBytes] = '|'
	d.right[nBytes+1] = '\n'
	_, err = d.w.Write(d.right[:nBytes+2])
	return
}
