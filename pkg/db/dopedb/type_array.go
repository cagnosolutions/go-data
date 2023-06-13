package dopedb

import (
	"log"
)

func (e *Encoder) writeFixArray(v []any) {
	if len(v) > bitFix/2 { // 15
		panic("cannot encodeValue, type does not match expected encoding")
	}
	e.write1(FixArray, uint8(len(v)))
	for i := range v {
		err := e.encodeValue(v[i])
		if err != nil {
			log.Panicf("error encoding fix array element [%T]: %s\n", v[i], err)
		}
	}
}

func (e *Encoder) writeArray16(v []any) {
	if len(v) > bit16 {
		panic("cannot encodeValue, type does not match expected encoding")
	}
	e.write3(Array16, uint16(len(v)))
	for i := range v {
		err := e.encodeValue(v[i])
		if err != nil {
			log.Panicf("error encoding array 16 element [%T]: %s\n", v[i], err)
		}
	}
}

func (e *Encoder) writeArray32(v []any) {
	if len(v) > bit32 {
		panic("cannot encodeValue, type does not match expected encoding")
	}
	e.write5(Array32, uint32(len(v)))
	for i := range v {
		err := e.encodeValue(v[i])
		if err != nil {
			log.Panicf("error encoding array 32 element [%T]: %s\n", v[i], err)
		}
	}
}
