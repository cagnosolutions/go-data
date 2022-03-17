package ndjson

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"reflect"
)

var (
	ErrPtrRequired        = errors.New("pointer required")
	ErrPtrToSliceRequired = errors.New("pointer is not pointing to a slice")
)

// LineReader reads one complete JSON document from each line.
type LineReader struct {
	br *bufio.Scanner
}

// NewLineReader returns a reader that reads newline delimited JSON documents.
func NewLineReader(r io.Reader) *LineReader {
	return &LineReader{
		br: bufio.NewScanner(r),
	}
}

func (lr *LineReader) ReadRaw() ([]byte, error) {
	// read next line delimited chunk of data.
	if lr.br.Scan() {
		// get the data we just scanned.
		b := lr.br.Bytes()
		// make a new slice of size len(b)
		dat := make([]byte, len(b))
		// and copy the data into it (so that if modified)
		// the underlying slice will not be altered
		copy(dat, b)
		// finally, return the data
		return dat, nil
	}
	// see if we produced any errors
	err := lr.br.Err()
	if err == nil {
		// for details on checking nil here, see below
		// https://cs.opensource.google/go/go/+/master:src/bufio/scan.go;l=97
		return nil, io.EOF
	}
	// otherwise, return
	return nil, err
}

// Decode decodes the next JSON document from the reader. It will return
// an error of io.EOF when the end of the file or stream is reached. The
// reader can continue to read lines even if the previous line returned
// and error (as long as the error is not io.EOF of course.)
func (lr *LineReader) Read(ptr interface{}) error {
	// read next line delimited chunk of data.
	if lr.br.Scan() {
		// get the data we just scanned.
		dat := lr.br.Bytes()
		// unmarshal the data into the provided pointer. If v is nil or
		// not a pointer, an InvalidUnmarshalError will be returned.
		return json.Unmarshal(dat, &ptr)
	}
	// see if we produced any errors
	err := lr.br.Err()
	if err == nil {
		// for details on checking nil here, see below
		// https://cs.opensource.google/go/go/+/master:src/bufio/scan.go;l=97
		return io.EOF
	}
	// otherwise, return
	return err
}

func (lr *LineReader) ReadAllRawMatch(pattern string) ([][]byte, error) {
	var datas [][]byte
	for {
		// read raw byte data
		data, err := lr.ReadRaw()
		if err != nil {
			// check for eof
			if err == io.EOF {
				break
			}
			// or error
			return datas, err
		}
		// check data against pattern
		if !match(data, []byte(pattern), matchAND) {
			// if not match was found, simply continue
			// on to the next data record...
			continue
		}
		// otherwise, we found one, so we should go
		// ahead and append it to our data set
		datas = append(datas, data)
	}
	// and we are done
	return datas, nil
}

func (lr *LineReader) ReadAllRaw() ([][]byte, error) {
	datas := make([][]byte, 0)
	for {
		// read raw byte data
		data, err := lr.ReadRaw()
		if err != nil {
			// check for eof
			if err == io.EOF {
				break
			}
			// or error
			return datas, err
		}
		// append newly read document into set
		datas = append(datas, data)
	}
	// and we are done
	return datas, nil
}

// ReadAllMatch reads and unmarshalls all the json entries that
// match the provided patter
func (lr *LineReader) ReadAllMatch(pattern string, v interface{}) (int, error) {
	// check to ensure we received a pointer value
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Ptr {
		return -1, ErrPtrRequired
	}
	// check to make sure that pointer value is a slice
	ptrElem := ptr.Type().Elem()
	if ptrElem.Kind() != reflect.Slice {
		return -1, ErrPtrToSliceRequired
	}
	// setup counter
	var n int
	// get the type of the element from the pointer
	elemType := ptrElem.Elem()
	for {
		// read raw byte data
		data, err := lr.ReadRaw()
		if err != nil {
			// check for eof
			if err == io.EOF {
				break
			}
			// or error
			return -1, err
		}
		// check data against pattern
		if !match(data, []byte(pattern), matchAND) {
			// if not match was found, simply continue
			// on to the next data record...
			continue
		}
		// construct a new instance of the type
		typ := reflect.New(elemType)
		// unmarshal into an interface of the underlying type
		err = json.Unmarshal(data, typ.Interface())
		if err != nil {
			return -1, err
		}
		// finally, take the pointer to the slice we passed in
		// and append a new (fully filled out) element to the
		// slice pointer, and continue...
		ptr.Elem().Set(reflect.Append(ptr.Elem(), typ.Elem()))
		// increment counter
		n++
	}
	// once finished, return a nil error
	return n, nil
}

// ReadAll reads and unmarshalls all the json entries
func (lr *LineReader) ReadAll(v interface{}) (int, error) {
	// check to ensure we received a pointer value
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Ptr {
		return -1, ErrPtrRequired
	}
	// check to make sure that pointer value is a slice
	ptrElem := ptr.Type().Elem()
	if ptrElem.Kind() != reflect.Slice {
		return -1, ErrPtrToSliceRequired
	}
	// setup counter
	var n int
	// get the type of the element from the pointer
	elemType := ptrElem.Elem()
	for {
		// read raw byte data
		data, err := lr.ReadRaw()
		if err != nil {
			// check for eof
			if err == io.EOF {
				break
			}
			// or error
			return -1, err
		}
		// construct a new instance of the type
		typ := reflect.New(elemType)
		// unmarshal into an interface of the underlying type
		err = json.Unmarshal(data, typ.Interface())
		if err != nil {
			return -1, err
		}
		// finally, take the pointer to the slice we passed in
		// and append a new (fully filled out) element to the
		// slice pointer, and continue...
		ptr.Elem().Set(reflect.Append(ptr.Elem(), typ.Elem()))
		// increment counter
		n++
	}
	// once finished, return a nil error
	return n, nil
}

type LineWriter struct {
	w io.Writer
}

func NewLineWriter(w io.Writer) *LineWriter {
	return &LineWriter{
		w: w,
	}
}

// WriteRaw writes data followed by a newline delimiter. It sanitizes the
// data and removes any newline characters before it finally writes.
func (lw *LineWriter) WriteRaw(data []byte) (int, error) {
	// get a buffered writer
	bw := bufio.NewWriter(lw.w)
	// range the data char for char and clean up and newlines
	for _, ch := range data {
		// skip any '\n' (0x0A) or '\r' (0x0D)
		if ch == 0x0A || ch == 0x0D {
			continue
		}
		// otherwise, we write the char out
		err := bw.WriteByte(ch)
		if err != nil {
			return -1, err
		}
	}
	// make sure we write the trailing newline char
	err := bw.WriteByte('\n')
	if err != nil {
		return -1, err
	}
	// get size in bytes
	n := bw.Buffered()
	// and make sure we flush the buffered writer
	err = bw.Flush()
	if err != nil {
		return -1, err
	}
	return n, nil
}

// Write writes data as a JSON document followed by a newline delimiter.
func (lw *LineWriter) Write(v interface{}) (int, error) {
	// marshal the data into a raw byte form
	dat, err := json.Marshal(v)
	if err != nil {
		return -1, err
	}
	// now call our WriteRaw method for appending our delimiter.
	n, err := lw.WriteRaw(dat)
	if err != nil {
		return -1, err
	}
	// success
	return n, nil
}
