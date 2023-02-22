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

func (s Span) String() string {
	return fmt.Sprintf("id=%d, beg=%d, end=%d", s.ID, s.Beg, s.End)
}

// Reference for this span type can be found at the location below.
// [https://cs.opensource.google/go/go/+/master:src/bytes/bytes.go;l=477]

// IndexData reads the data from the reader and returns a list of span offsets
// for each delimited section of data. It uses ReadLine which has been tested
// against ReadBytes, ReadSlice and Scanner, and it is as fast (or faster) and
// uses the smallest buffer of all of them.
// func (dr *DelimReader) IndexData() ([]Span, error) {
// 	// get a new buffered reader
// 	br := bufio.NewReader(dr.r)
// 	// Declare span offsets
// 	var beg, end int
// 	var id int
// 	var spans []Span
// 	var buffers int
// 	for {
// 		if beg < end {
// 			beg = end
// 		}
// 		data, isPrefix, err := br.ReadLine()
// 		if err != nil {
// 			if err == io.EOF {
// 				isPrefix = false
// 				break
// 			}
// 			return nil, err
// 		}
// 		// if beg < end && !isPrefix {
// 		//	beg = end
// 		// }
// 		if isPrefix {
// 			// Just the prefix of a segment, more to follow
// 			buffers++
// 			fmt.Printf(">>> [T-PRE]: buff=%d, beg=%d, end=%d, id=%d\n", buffers, beg, end, id)
// 			continue
// 		} else {
// 			// No more to the segment, simply add to span set
// 			end = beg + len(data) + len(newline)
// 			id++
// 			spans = append(
// 				spans, Span{
// 					ID:  id,
// 					Beg: beg,
// 					End: end,
// 				},
// 			)
// 			beg = end
// 			buffers = 0
// 			fmt.Printf(">>> [F-PRE]: buff=%d, beg=%d, end=%d, id=%d\n", buffers, beg, end, id)
// 		}
// 	}
// 	return spans, nil
// }

// func (dr *DelimReader) IndexData2() ([]Span, error) {
// 	br := bufio.NewReader(dr.r)
// 	var spans []Span
// 	var id int
// 	var beg, end int
// 	var offset int
// 	for {
// 		data, more, err := br.ReadLine()
// 		if err != nil {
// 			if err == io.EOF {
// 				break
// 			}
// 			return nil, err
// 		}
// 		// update offset
// 		offset += len(data) + 2
// 		// More to process, update end offset
// 		if more {
// 			end = offset
// 		}
// 		// Have a full span, so we can append it
// 		if !more {
// 			if end > (beg + len(data) + 2) {
// 				spans = append(spans, Span{id, beg, end})
// 			} else {
// 				end = beg + len(data) + 2
// 				spans = append(spans, Span{id, beg, end})
// 			}
// 			id++
// 			beg = offset
// 		}
// 		if data == nil && !more {
// 			fmt.Println(">>>", id)
// 		}
// 		fmt.Printf("(id=%d, more=%t, offset=%d) read %d bytes, beg=%d, end=%d\n", id, more, offset, len(data), beg, end)
// 		// end = beg + len(data)
// 		//line = append(line, data...)
// 		// if !more {
// 		//	break
// 		// }
// 	}
// 	return spans, nil
// }

// func readLine(r *bufio.Reader) ([]byte, error) {
// 	var line []byte
// 	var beg, end int
// 	for {
// 		l, more, err := r.ReadLine()
// 		if err != nil {
// 			return nil, err
// 		}
// 		end += len(l)
// 		// Avoid the copy if the first call produced a full line.
// 		if line == nil && !more {
// 			return l, nil
// 		}
// 		line = append(line, l...)
// 		if !more {
// 			break
// 		}
// 		fmt.Printf("beg=%d, end=%d\n", beg, end)
// 		beg = end
// 	}
// 	return line, nil
// }

func (dr *DelimReader) LineScanner() ([]Span, error) {
	var spans []Span
	var id, beg, end int
	scanner := bufio.NewScanner(dr.r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		data := scanner.Bytes()
		end += len(data) + 2 // add two to skip the trailing \r\n
		spans = append(
			spans, Span{
				ID:  id,
				Beg: beg,
				End: end - 2,
			},
		)
		beg = end
		id++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return spans, nil
}

func (dr *DelimReader) LineReader() ([]Span, error) {
	// Setup initial variables for the function
	var spans []Span
	var id, beg, end int
	// This lf variable is our system dependent number of characters for a linefeed.
	lf := newLineBytes()
	// get a new buffered reader set to our determined buffer size.
	br := bufio.NewReaderSize(dr.r, 4096)
	for {
		// Read up to buffer size length of data and look for the delimiter. If we fill
		// up the buffer and do not find the delimiter we are looking for, we will just
		// keep reading, one buffer length at a time, until we find it.
		data, err := br.ReadSlice('\n')
		if err != nil {
			if err == io.EOF {
				// We have reached the end--we are going to check for any remaining data.
				if len(data) > 0 {
					// We have some leftover data, which means the stream was not newline
					// terminated. Add the remaining data to one last span before breaking.
					spans = append(spans, Span{id + 1, beg, end + len(data)})
				}
				// Otherwise, the stream is indeed newline terminated, so we can just break.
				break
			}
			if err == bufio.ErrBufferFull {
				// Our buffer seems to be full, so at this point we will simply update the
				// ending offset and continue reading (skipping all the stuff below, and
				// restarting the loop from the next iteration.)
				end += len(data)
				continue
			}
			// Uh oh, we have some other issue going on.
			return nil, err
		}
		// We were able to locate a newline without filling the buffer, so we should update
		// our ending offset; then add our span data to our set.
		end += len(data)
		// We subtract two from the ending offset because we don't want to include any of
		// that new-line characters that we read in our span. On Windows this number will be
		// two, and on Mac (darwin), Linux, and Unix's it will be 1 character. Win="\r\n",
		// *Nix="\n"
		spans = append(spans, Span{id, beg, end - lf})
		// We will grow the beginning offset up to where the end is, and increment the id.
		beg = end
		id++
	}
	return spans, nil
}

// ALMOST FINAL RESULT --> [https://go.dev/play/p/o9DYbf6xA7G]

// IndexSpans can be used to index spans of text based around a delimiter of your
// choice. The size argument allows you to tune it a bit and have some control over
// the overhead used by the function. The delimiter is not included in the returned
// span bound set and empty lines are not ignored.
func IndexSpans(r io.Reader, delim byte, size int) ([]Span, error) {
	// Setup initial variables for the function
	var id, beg, end int
	// Drop is a helper func used to drop the correct number of bytes. Right now it
	// is mostly used to handle the special case of \n and \r\n
	drop := func(p []byte, c byte) int {
		if c == '\n' {
			if len(p) > 1 && p[len(p)-2] == '\r' {
				return 2
			}
		}
		return 1
	}
	// Initialize our spans
	spans := make([]Span, 0, 8)
	// get a new buffered reader set to our determined buffer size.
	br := bufio.NewReaderSize(r, size)
	for {
		// Read up to buffer size length of data and look for the delimiter. If we fill
		// up the buffer and do not find the delimiter we are looking for, we will just
		// keep reading, one buffer length at a time, until we find it. Note: ReadSlice
		// attempts to re-use the same buffer internally, so that helps a lot.
		data, err := br.ReadSlice(delim)
		if err != nil {
			if err == io.EOF {
				// We have reached the end--we are going to check for any remaining data.
				if len(data) > 0 {
					// We have some leftover data, which means the stream was not delimiter
					// terminated. Add the remaining data to one last span before breaking.
					spans = append(spans, Span{id + 1, beg, end + len(data)})
				}
				// Otherwise, the stream is indeed delimiter terminated, so we can just break.
				break
			}
			if err == bufio.ErrBufferFull {
				// Our buffer seems to be full, so at this point we will simply update the
				// ending offset and continue reading (skipping all the stuff below, and
				// restarting the loop from the next iteration.)
				end += len(data)
				continue
			}
			// Uh oh, we have some other issue going on.
			return nil, err
		}
		// We were able to locate a delimiter without filling the buffer, so we should update
		// our ending offset; then add our span data to our set.
		end += len(data)
		// Calculate number of bytes to drop
		n := drop(data, delim)
		// Add our span to our set and adjust the beginning, ending and id variables
		spans = append(spans, Span{id, beg, end - n})
		// We will grow the beginning offset up to where the end is, and increment the id.
		beg = end
		id++
	}
	return spans, nil
}
