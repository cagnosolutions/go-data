package wal2

import (
	"bytes"
	"reflect"
	"testing"
)

func TestEntry_MakeEntry(t *testing.T) {
	var e *entry
	e = makeEntry([]byte("foo"), nil)
	if e == nil {
		t.Fatal("could not make entry, but should have been able to")
	}
	k := e.getKey()
	if k == nil || len(k) != 3 {
		t.Fatal("get key failed")
	}
	e = makeEntry(nil, []byte("bar"))
	if e == nil {
		t.Fatal("could not make entry, but should have been able to")
	}
	v := e.getVal()
	if v == nil || len(v) != 3 {
		t.Fatal("get val failed")
	}
}

func TestEntry_EncodeDecode(t *testing.T) {
	// create key and value
	key := []byte("entry-001")
	val := []byte("this is some value")
	// create entry 1
	e1 := makeEntry(key, val)
	if e1 == nil {
		t.Fatal("could not make entry, but should have been able to")
	}
	// make a new buffer to encode entry 1 into
	buf := bytes.NewBuffer(nil)
	err := e1.writeTo(buf)
	if err != nil {
		t.Fatal("something went wrong while encoding")
	}
	// check encoded entry
	if !bytes.ContainsAny(append(key, val...), buf.String()) {
		t.Fatal("encoding failed to write data")
	}
	if buf.Len() != e1.size() {
		t.Fatal("encoding size is incorrect", buf.Len(), e1.size())
	}

	// create entry 2 to decode into
	var e2 entry
	err = e2.readFrom(buf)
	if err != nil {
		t.Fatal("encountered error while decoding:", err)
	}
	// check the entries against eachother
	if !reflect.DeepEqual(e1, &e2) || e1.String() != (&e2).String() {
		t.Fatal("entries are not equal")
	}
}
