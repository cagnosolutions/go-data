package dopedb

import (
	"encoding/binary"
	"fmt"
)

var Set setTypes

type setTypes struct{}

func (s setTypes) EncFixArray(p []byte, v []any) {
	_ = p[len(v)+1]        // early bounds check to guarantee safety of writes below
	if len(v) > bitFix/2 { // 15
		panic("cannot encode, type does not match expected encoding")
	}
	var n int
	p[n] = byte(FixArray | len(v))
	n++
	for i := range v {
		copy(p[n:], v)
	}
}

func (s setTypes) EncArray16(p []byte, v []any) {
	// TODO implement me
	panic("implement me")
}

func (s setTypes) EncArray32(p []byte, v []any) {
	// TODO implement me
	panic("implement me")
}

func (s setTypes) EncFixMap(p []byte, v map[string]any) {
	// TODO implement me
	panic("implement me")
}

func (s setTypes) EncMap16(p []byte, v map[string]any) {
	// TODO implement me
	panic("implement me")
}

func (s setTypes) EncMap32(p []byte, v map[string]any) {
	// TODO implement me
	panic("implement me")
}

func (s setTypes) DecFixArray(p []byte) []any {
	// TODO implement me
	panic("implement me")
}

func (s setTypes) DecArray16(p []byte) []any {
	// TODO implement me
	panic("implement me")
}

func (s setTypes) DecArray32(p []byte) []any {
	// TODO implement me
	panic("implement me")
}

func (s setTypes) DecFixMap(p []byte) map[string]any {
	// TODO implement me
	panic("implement me")
}

func (s setTypes) DecMap16(p []byte) map[string]any {
	// TODO implement me
	panic("implement me")
}

func (s setTypes) DecMap32(p []byte) map[string]any {
	// TODO implement me
	panic("implement me")
}

func (s stringTypes) EncFixStr(p []byte, v string) {
	_ = p[len(v)+1]      // early bounds check to guarantee safety of writes below
	if len(v) > bitFix { // bitFix = 31
		panic("cannot encode, type does not match expected encoding")
	}
	p[0] = byte(FixStr | len(v))
	copy(p[1:], v)
}

func (s stringTypes) EncStr8(p []byte, v string) {
	_ = p[len(v)+2]    // early bounds check to guarantee safety of writes below
	if len(v) > bit8 { // bit8 = 255
		panic("cannot encode, type does not match expected encoding")
	}
	p[0] = Str8
	p[1] = byte(len(v))
	copy(p[2:], v)
}

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
