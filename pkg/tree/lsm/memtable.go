package lsm

import (
	"encoding/binary"
	"io"

	"github.com/cagnosolutions/go-data/pkg/tree/rbt/generic"
)

type Memtable struct {
	tree     *generic.RBTree[string, string]
	keySpace int
	valSpace int
	max      int
}

func NewMemtable(max int) *Memtable {
	return &Memtable{
		tree: generic.NewTree[string, string](),
		max:  max,
	}
}

func (m *Memtable) isFull() (full bool) {
	return m.tree.Len() >= m.max
}

func (m *Memtable) Flush(w *io.OffsetWriter) (off int64, err error) {
	// seek to the end of the file
	// off, err = w.Seek(0, io.SeekEnd)
	// if err != nil {
	// 	return off, err
	// }
	valOffs := m.keySpace
	// write the contents of each entry in the table
	m.tree.Scan(
		func(key string, val string) bool {
			// write key index
			buf := make([]byte, 2)
			binary.BigEndian.PutUint16(buf, uint16(len(key)))
			_, err = w.Write(buf)
			if err != nil {
				return false
			}
			_, err = w.Write([]byte(key))
			if err != nil {
				return false
			}
			// write value pointer for key
			binary.BigEndian.PutUint16(buf, uint16(valOffs))
			_, err = w.Write(buf)
			if err != nil {
				return false
			}
			// write value
			binary.BigEndian.PutUint16(buf, uint16(len(val)))
			_, err = w.WriteAt(buf, int64(valOffs))
			if err != nil {
				return false
			}
			valOffs += len(buf)
			_, err = w.WriteAt([]byte(val), int64(valOffs))
			if err != nil {
				return false
			}
			valOffs += len(val)

			return true
		},
	)
	return
}

func (m *Memtable) Put(key string, val string) bool {
	if m.isFull() {
		return true
	}
	m.keySpace += len(key) + 4
	m.valSpace += len(val) + 4
	m.tree.Put(key, val)
	return false
}

func (m *Memtable) Get(key string) (val string, found bool) {
	return "", false
}

func (m *Memtable) Del(key string) (val string, removed bool) {
	return "", false
}

// type Entry struct {
// 	Key
// 	Val
// }

type Key struct {
	KeyLen uint16
	KeyStr string
	OffVal uint64
}

type Val struct {
	ValLen uint32
	ValStr string
}

// type SSTable struct {
// 	NumEntries  uint64
// 	OffFirstKey uint64
// 	OffLastKey  uint64
// 	OffValues   uint64
// 	Keys        []Key
// 	Vals        []Val
// }
//
// func sizesForKeyspace(entries []Entry) (uint64, uint64) {
// 	var lastKey, valOff uint64
// 	for i, entry := range entries {
// 		if i < len(entries)-1 {
// 			lastKey += uint64(2 + len(entry.KeyStr) + 8)
// 		}
// 		valOff += uint64(2 + len(entry.ValStr) + 8)
// 	}
// 	return lastKey, valOff
// }
//
// func MakeSSTable(entries []Entry) *SSTable {
// 	offLastKey, offVals := sizesForKeyspace(entries)
// 	sst := &SSTable{
// 		NumEntries:  uint64(len(entries)),
// 		OffFirstKey: 8,
// 		OffLastKey:  offLastKey,
// 		OffValues:   offVals,
// 		Keys:        make([]Key, len(entries)),
// 		Vals:        make([]Val, len(entries)),
// 	}
// 	for i, entry := range entries {
// 		sst.Keys[i] = entry.Key
// 		sst.Vals[i] = entry.Val
// 	}
// 	sort.Stable(sst)
// 	return sst
// }
//
// func (sst *SSTable) Read(p []byte) (n int, err error) {
// 	// Create new SSTable
// 	newSST := new(SSTable)
//
// 	// Read number of entries
//
// 	// Read offset of first key
//
// 	// Read offset of last key
//
// 	// Read keys in sorted order
//
// 	// Read values
//
// 	// Assign new SSTable to current one
// 	*sst = *newSST
//
// 	return -1, nil
// }
//
// func (sst *SSTable) Write(p []byte) (n int, err error) {
// 	// Ensure SSTable is not empty
// 	if sst == nil || (len(sst.Keys) == len(sst.Vals) && (len(sst.Keys) < 1)) {
// 		return -1, errors.New("sstable is empty")
// 	}
//
// 	// Write number of entries
//
// 	// Write offset of first key
//
// 	// Write offset of last key
//
// 	// Write keys in sorted order
//
// 	// Write values
//
// 	z := ^uint16(0)
//
// 	return -1, nil
// }
//
// func (sst *SSTable) Len() int {
// 	return len(sst.Keys)
// }
//
// func (sst *SSTable) Less(i, j int) bool {
// 	return sst.Keys[i].KeyStr < sst.Keys[j].KeyStr
// }
//
// func (sst *SSTable) Swap(i, j int) {
// 	sst.Keys[i], sst.Keys[j] = sst.Keys[j], sst.Keys[i]
// 	sst.Vals[i], sst.Vals[j] = sst.Vals[j], sst.Vals[i]
// }

// SSTable Index
// [KeyLen][Key][ValLen][ValPtr],[KeyLen][Key][ValLen][ValPtr],[KeyLen][Key][ValLen][ValPtr]
//  uint16   *   uint16  uint32

// SSTable
// [KeyLen][Key][ValLen][Val],[KeyLen][Key][ValLen][Val],[KeyLen][Key][ValLen][Val]
