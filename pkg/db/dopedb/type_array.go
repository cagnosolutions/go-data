package dopedb

import (
	"encoding/binary"
	"fmt"
)

func encFixArray(p []byte, v []any) {
	if !hasRoom(p, 1) {
		panic(ErrWritingBuffer)
	}
	p[0] = byte(FixArray | len(v))

	// TODO: implement...
	for i := range v {
		fmt.Println(i)
	}
}

func encArray16(p []byte, v []any) {
	if !hasRoom(p, 3) {
		panic(ErrWritingBuffer)
	}
	p[0] = Array16
	binary.BigEndian.PutUint16(p[1:3], uint16(len(v)))

	// TODO: implement...
	for i := range v {
		fmt.Println(i)
	}
}

func encArray32(p []byte, v []any) {
	if !hasRoom(p, 5) {
		panic(ErrWritingBuffer)
	}
	p[0] = Array32
	binary.BigEndian.PutUint32(p[1:5], uint32(len(v)))

	// TODO: implement...
	for i := range v {
		fmt.Println(i)
	}
}

func decFixArray(p []byte) []any {

	// TODO: implement...
	return nil
}

func decArray16(p []byte) []any {

	// TODO: implement...
	return nil
}

func decArray32(p []byte) []any {

	// TODO: implement...
	return nil
}
