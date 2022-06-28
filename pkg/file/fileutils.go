package file

import (
	"bufio"
	"io"
	"runtime"
)

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
	case "linux", "unix":
		return 1 // \n
	case "windows":
		return 2 // \r\n
	}
	return 0
}

type Span struct {
	Line     int // Line number of this span
	Beg, End int // Beginning and ending offsets for this span
}

// IndexData reads the data from the reader and returns a list of span offsets
// for each delimited section of data. It uses ReadLine which has been tested
// against ReadBytes, ReadSlice and Scanner, and it is as fast (or faster) and
// uses the smallest buffer of all of them.
func (dr *DelimReader) IndexData() ([]Span, error) {
	// Get a new buffered reader
	br := bufio.NewReader(dr.r)
	// Declare span offsets
	var beg, end int
	var line int
	var spans []Span
	for {
		if beg < end {
			beg = end
		}
		data, prefix, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				prefix = false
				break
			}
			return nil, err
		}
		// Not the same line as the previous one, so we will add a span
		if !prefix {
			end = beg + len(data) + newLineBytes()
			line++
			spans = append(
				spans, Span{
					Line: line,
					Beg:  beg,
					End:  end,
				},
			)
		}
	}
	return spans, nil
}