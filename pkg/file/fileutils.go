package file

import (
	"bufio"
	"fmt"
	"io"
	"runtime"
)

var newline = []byte{'\r', '\n'}

func init() {
	if newLineBytes() == 1 {
		newline = []byte{'\n'}
	}
}

// DelimReader reads and returns one section of data with each successive call
// to ReadNext, or ScanNext as denoted by the delimiter provided at instantiation.
type DelimReader struct {
	r     io.Reader
	delim byte
}

func NewDelimReader(r io.Reader, delim byte) *DelimReader {
	return &DelimReader{
		r:     r,
		delim: delim,
	}
}

func newLineBytes() int {
	switch runtime.GOOS {
	case "linux", "unix", "darwin":
		return 1 // \n
	case "windows":
		return 2 // \r\n
	}
	return 0
}

type Span struct {
	ID       int // Index number of this span
	Beg, End int // Beginning and ending offsets for this span
}

// Reference for this span type can be found at the location below.
// [https://cs.opensource.google/go/go/+/master:src/bytes/bytes.go;l=477]

// IndexData reads the data from the reader and returns a list of span offsets
// for each delimited section of data. It uses ReadLine which has been tested
// against ReadBytes, ReadSlice and Scanner, and it is as fast (or faster) and
// uses the smallest buffer of all of them.
func (dr *DelimReader) IndexData() ([]Span, error) {
	// Get a new buffered reader
	br := bufio.NewReader(dr.r)
	// Declare span offsets
	var beg, end int
	var id int
	var spans []Span
	var buffers int
	for {
		if beg < end {
			beg = end
		}
		data, isPrefix, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				isPrefix = false
				break
			}
			return nil, err
		}
		// if beg < end && !isPrefix {
		//	beg = end
		// }
		if isPrefix {
			// Just the prefix of a segment, more to follow
			buffers++
			fmt.Printf(">>> [T-PRE]: buff=%d, beg=%d, end=%d, id=%d\n", buffers, beg, end, id)
			continue
		} else {
			// No more to the segment, simply add to span set
			end = beg + len(data) + len(newline)
			id++
			spans = append(
				spans, Span{
					ID:  id,
					Beg: beg,
					End: end,
				},
			)
			beg = end
			buffers = 0
			fmt.Printf(">>> [F-PRE]: buff=%d, beg=%d, end=%d, id=%d\n", buffers, beg, end, id)
		}
	}
	return spans, nil
}

func (dr *DelimReader) IndexData2() ([]Span, error) {
	br := bufio.NewReader(dr.r)
	var spans []Span
	var id int
	var beg, end int
	var offset int
	for {
		data, more, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		// update offset
		offset += len(data) + 2
		// More to process, update end offset
		if more {
			end = offset
		}
		// Have a full span, so we can append it
		if !more {
			if end > (beg + len(data) + 2) {
				spans = append(spans, Span{id, beg, end})
			} else {
				end = beg + len(data) + 2
				spans = append(spans, Span{id, beg, end})
			}
			id++
			beg = offset
		}
		if data == nil && !more {
			fmt.Println(">>>", id)
		}
		fmt.Printf("(id=%d, more=%t, offset=%d) read %d bytes, beg=%d, end=%d\n", id, more, offset, len(data), beg, end)
		// end = beg + len(data)
		//line = append(line, data...)
		// if !more {
		//	break
		// }
	}
	return spans, nil
}
