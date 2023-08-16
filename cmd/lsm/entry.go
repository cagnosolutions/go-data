package main

import (
	"encoding/binary"
	"io"
)

type Entry struct {
	Key string
	Val string
}

func (e Entry) WriteTo(w io.Writer) (n int64, err error) {

	// write key length
	err = binary.Write(w, binary.BigEndian, uint32(len(e.Key)))
	if err != nil {
		return n, err
	}
	n += 4

	// write val length
	err = binary.Write(w, binary.BigEndian, uint32(len(e.Val)))
	if err != nil {
		return n, err
	}
	n += 4

	// write key
	_, err = w.Write([]byte(e.Key))
	if err != nil {
		return n, err
	}
	n += int64(len(e.Key))

	// write val
	_, err = w.Write([]byte(e.Val))
	if err != nil {
		return n, err
	}
	n += int64(len(e.Val))

	return n, nil
}

func (e Entry) Write(p []byte) (int, error) {
	if len(p) < 8+len(e.Key)+len(e.Val) {
		return -1, io.ErrShortWrite
	}
	var n int
	// write key len
	binary.BigEndian.PutUint32(p[n:n+4], uint32(len(e.Key)))
	n += 4
	// write val len
	binary.BigEndian.PutUint32(p[n:n+4], uint32(len(e.Val)))
	n += 4
	// write key
	n += copy(p[n:], e.Key)
	// write val
	n += copy(p[n:], e.Val)
	return n, nil
}
