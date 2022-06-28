package wal2

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

// https://github.com/scottcagno/storage/blob/main/pkg/swal
// https://github.com/scottcagno/storage/tree/e54fed254539aebc0588d0a736faa69dc1bfbf27/pkg/lsmt/wal

type entry struct {
	szkey uint16
	szval uint16
	data  []byte
}

func makeEntry(key []byte, val []byte) *entry {
	if key == nil && val == nil {
		panic("cannot make entry with nil data; at least one of the values must be filled out")
	}
	szkey, szval := len(key), len(val)
	if szkey > math.MaxUint16 || szval > math.MaxUint16 {
		panic("cannot make entry; key or value exceeds uint16 size in length")
	}
	// data := make([]byte, szkey+szval)
	// copy(data[0:szkey], key)
	// copy(data[szval:], val)
	return &entry{
		szkey: uint16(szkey),
		szval: uint16(szval),
		data:  append(key, val...),
	}
}

func (e *entry) getKey() []byte {
	if e.szkey < 1 {
		return nil
	}
	key := e.data[:e.szkey]
	return key
}

func (e *entry) getVal() []byte {
	if e.szval < 1 {
		return nil
	}
	val := e.data[:e.szval]
	return val
}

func (e *entry) writeTo(w io.Writer) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint16(buf[0:2], e.szkey)
	binary.LittleEndian.PutUint16(buf[2:4], e.szval)
	_, err := w.Write(buf)
	if err != nil {
		return err
	}
	_, err = w.Write(e.data)
	if err != nil {
		return err
	}
	return nil
}

func newEntryForCallerInCallee(pent **entry) {
	ent := *pent
	ent = &entry{}
	**pent = *ent
}

func (e *entry) readFrom(r io.Reader) error {
	if e == nil {
		return errors.New("entry is nil, cannot assign to it")
	}
	buf := make([]byte, 4)
	_, err := r.Read(buf)
	if err != nil {
		return err
	}
	e.szkey = binary.LittleEndian.Uint16(buf[0:2])
	e.szval = binary.LittleEndian.Uint16(buf[2:4])
	e.data = make([]byte, e.szkey+e.szval)
	_, err = r.Read(e.data)
	if err != nil {
		return err
	}
	return nil
}

func (e *entry) size() int {
	return 4 + len(e.data)
}

func (e *entry) String() string {
	return fmt.Sprintf("entry:\n  szkey=%v\n  szval=%v\n  data=%q\n", e.szkey, e.szval, e.data)
}
