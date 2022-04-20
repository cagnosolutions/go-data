package util

import (
	"bytes"
	"io"
	"os"
)

const bufSize = 4096

type span struct {
	beg int
	end int
}

type Document struct {
	spans []span
	buf   []byte
	cur   int
	fp    *os.File
}

func NewDocument(fp *os.File) (*Document, error) {
	d := &Document{
		buf: make([]byte, bufSize),
		cur: 0,
		fp:  fp,
	}
	if err := d.index(); err != nil {
		return nil, err
	}
	return d, nil
}

func IndexAll(data, find []byte, chars int) []span {
	var j, i int
	left := len(data)
	spans := make([]span, 0, 32)
	for {
		i = bytes.Index(data[j:], find)
		if i == -1 {
			break
		}
		left -= i
		if chars < left {
			spans = append(spans, span{beg: i + j, end: i + j + chars})
			i += chars
		}
		j += i
	}
	return spans
}

// spansFunc indexes the slice p at each run of code points satisfying f(b)
// and returns a slice of indexed spans of p. If all code points in p
// satisfy f(b), or len(p) == 0, an empty slice is returned.
func spansFunc(p []byte, f func(byte) bool) []span {
	// A span is used to record a slice of s of the form s[beg:end].
	// The start index is inclusive and the end index is exclusive.
	spans := make([]span, 0, 32)
	// Find the field start and end indices.
	// Doing this in a separate pass (rather than slicing the string s
	// and collecting the result substrings right away) is significantly
	// more efficient, possibly due to cache effects.
	beg := -1 // valid span start if >= 0
	for i := 0; i < len(p); i++ {
		if f(p[i]) {
			if beg >= 0 {
				spans = append(spans, span{beg, i})
				beg = -1
				continue
			}
		}
		if beg < 0 {
			beg = i
		}
	}
	// Return span slice
	return spans
}

// spansDelim indexes the slice p at each run of code points satisfying f(b)
// and returns a slice of indexed spans of p. If all code points in p
// satisfy f(b), or len(p) == 0, an empty slice is returned.
func spansDelim(p []byte, delim byte) []span {
	// A span is used to record a slice of s of the form s[beg:end].
	// The start index is inclusive and the end index is exclusive.
	spans := make([]span, 0, 32)
	// Find the field start and end indices.
	// Doing this in a separate pass (rather than slicing the string s
	// and collecting the result substrings right away) is significantly
	// more efficient, possibly due to cache effects.
	beg := -1 // valid span start if >= 0
	for i := 0; i < len(p); i++ {
		if p[i] == delim {
			if beg >= 0 {
				spans = append(spans, span{beg, i})
				beg = -1
				continue
			}
		}
		if beg < 0 {
			beg = i
		}
	}
	// Return span slice
	return spans
}

// index reads through the entire file and indexes by
// storing span offsets
func (d *Document) index() error {
	// start reading
	for {
		// Fill up out buffer.
		_, err := d.fp.Read(d.buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		// Initialize our document spans using the newline delimiter.
		d.spans = append(d.spans, spansDelim(d.buf, '\n')...)
	}
	return nil
}
