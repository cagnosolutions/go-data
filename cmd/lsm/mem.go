package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type Memtable struct {
	mu   sync.Mutex
	data []Entry
}

func NewMemtable() *Memtable {
	return &Memtable{
		data: make([]Entry, 0),
	}
}

func (m *Memtable) Len() int {
	return len(m.data)
}

func (m *Memtable) Less(i, j int) bool {
	return m.data[i].Key < m.data[j].Key
}

func (m *Memtable) Swap(i, j int) {
	m.data[i], m.data[j] = m.data[j], m.data[i]
}

func (m *Memtable) find(k string) (int, bool) {
	return sort.Find(
		len(m.data), func(i int) int {
			return strings.Compare(k, m.data[i].Key)
		},
	)
}

func (m *Memtable) Get(k string) string {
	i, found := m.find(k)
	if !found {
		return ""
	}
	return m.data[i].Val
}

func (m *Memtable) Put(k string, v string) {
	i, found := m.find(k)
	if !found {
		// insert
		m.data = append(m.data, Entry{k, v})
		sort.Stable(m)
		return
	}
	// update
	m.data[i].Val = v
}

const dead = `0xdead`

func (m *Memtable) Del(k string) {
	i, found := m.find(k)
	if !found {
		m.data = append(m.data, Entry{k, dead})
		return
	}
	m.data[i].Val = dead
}

func writeEntry(fp *os.File, e *Entry) error {
	// write key length
	err := binary.Write(fp, binary.BigEndian, uint32(len(e.Key)))
	if err != nil {
		return err
	}
	// write val length
	err = binary.Write(fp, binary.BigEndian, uint32(len(e.Val)))
	if err != nil {
		return err
	}
	// write key
	_, err = fp.WriteString(e.Key)
	if err != nil {
		return err
	}
	// write value
	_, err = fp.WriteString(e.Val)
	if err != nil {
		return err
	}
	return nil
}

func ReadSequence(base string) int {
	// check to see if file exists
	fname := filepath.ToSlash(filepath.Join(base, "seq.db"))
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		// file needs to be created
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(1))
		err = os.WriteFile(fname, buf, 0655)
		if err != nil {
			panic(err)
		}
		return 1
	}
	// open file
	fp, err := os.OpenFile(fname, os.O_RDWR, 0655)
	if err != nil {
		panic(err)
	}
	// read sequence number
	var seq uint32
	err = binary.Read(fp, binary.BigEndian, &seq)
	if err != nil {
		panic(err)
	}
	// seek to beginning
	_, err = fp.Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}
	// update sequence number
	err = binary.Write(fp, binary.BigEndian, seq+1)
	if err != nil {
		panic(err)
	}
	err = fp.Sync()
	if err != nil {
		panic(err)
	}
	err = fp.Close()
	if err != nil {
		panic(err)
	}
	return int(seq)
}

func (m *Memtable) Flush(base string) error {
	// lock the memtable
	m.mu.Lock()
	defer m.mu.Unlock()

	// create any base directories
	err := os.MkdirAll(base, 0655)
	if err != nil {
		return err
	}

	// base = filepath.Join(base, fmt.Sprintf("%.16d", time.Now().Unix()))

	// read the current sequence number
	seq := ReadSequence(base)

	// update base path
	base = filepath.Join(base, fmt.Sprintf("%.4d", seq))

	// create any further base directories
	err = os.MkdirAll(base, 0655)
	if err != nil {
		return err
	}

	// Write ss table file
	index, err := WriteSSTable(base, m.data)
	if err != nil {
		return err
	}

	// Write the index file
	err = WriteIndex(base, index)
	if err != nil {
		return err
	}

	// shrink mem table
	m.data = m.data[:0]
	return nil
}
