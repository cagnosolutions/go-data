package main

import (
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func WriteSSTable(base string, data []Entry) ([]IndexEntry, error) {
	// create our main ss table file
	fp, err := os.Create(filepath.Join(base, "data.db"))
	if err != nil {
		return nil, err
	}

	var index []IndexEntry
	var empty Entry

	// range our data
	for i, e := range data {

		// get the current offset
		current, err := fp.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, err
		}

		// add it to the index entries
		index = append(
			index, IndexEntry{
				Key: e.Key,
				Off: uint32(current),
			},
		)

		// write key length
		err = binary.Write(fp, binary.BigEndian, uint32(len(e.Key)))
		if err != nil {
			return nil, err
		}

		// write val length
		err = binary.Write(fp, binary.BigEndian, uint32(len(e.Val)))
		if err != nil {
			return nil, err
		}

		// write key
		_, err = fp.WriteString(e.Key)
		if err != nil {
			return nil, err
		}

		// write value
		_, err = fp.WriteString(e.Val)
		if err != nil {
			return nil, err
		}

		// clear out entry
		data[i] = empty
	}

	// flush and close the ss table file
	err = fp.Sync()
	if err != nil {
		return nil, err
	}

	err = fp.Close()
	if err != nil {
		return nil, err
	}

	return index, nil
}

type SSTableIndex struct {
	Path  string
	Index []IndexEntry
}

func OpenSSTableIndex(base string) *SSTableIndex {
	return &SSTableIndex{
		Path:  base,
		Index: ReadIndex(base),
	}
}

func (sst *SSTableIndex) Get(k string) string {
	i, found := sort.Find(
		len(sst.Index), func(i int) int {
			return strings.Compare(k, sst.Index[i].Key)
		},
	)
	if !found {
		return ""
	}
	fp, err := os.Open(filepath.Join(sst.Path, "data.db"))
	if err != nil {
		panic(err)
	}
	e, err := readEntryAt(fp, sst.Index[i].Off)
	if err != nil {
		panic(err)
	}
	err = fp.Close()
	if err != nil {
		panic(err)
	}
	return e.Val
}

// func OpenSSTable(path string) (*SSTable, error) {
// 	// open the index file
// 	fp, err := os.Open(filepath.Join(path, "index.db"))
// 	if err != nil {
// 		return nil, err
// 	}
// 	// create a new key index
// 	index := make([]uint32, 0)
//
// 	buf := make([]byte, 4)
// 	for {
// 		// read index entry
// 		_, err = fp.Read(buf)
// 		if err != nil {
// 			if err == io.EOF {
// 				break
// 			}
// 			log.Panicf("error reading: %s", err)
// 		}
// 		// add offset to key index
// 		index = append(index, binary.BigEndian.Uint32(buf))
// 	}
// 	// close the index file
// 	err = fp.Close()
// 	if err != nil {
// 		return nil, err
// 	}
// 	// open the data file
// 	fp, err = os.Open(filepath.Join(path, "data.db"))
// 	if err != nil {
// 		return nil, err
// 	}
// 	// create a new key index set
// 	keys := make([]IndexEntry, 0)
// 	// assemble the table
// 	for _, idx := range index {
// 		// read the key index
// 		ki, err := readKeyIndex(fp, idx)
// 		if err != nil {
// 			return nil, err
// 		}
// 		// and append to our list of keys
// 		keys = append(keys, *ki)
// 	}
// 	// close the data file
// 	err = fp.Close()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &SSTable{
// 		Path: path,
// 		Keys: keys,
// 	}, nil
//
// }

/*

type SSTable struct {
	file    string
	entries []Entry
}

func Open(file string) (*SSTable, error) {
	if !strings.HasSuffix(file, ".sst") {
		file += ".sst"
	}
	sst := &SSTable{
		file:    file,
		entries: nil,
	}
	if Exists(file) {
		// open file
		fp, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		// load entries
		err = ReadSSTable(fp, sst)
		if err != nil {
			return nil, err
		}
		// close file
		err = fp.Close()
		if err != nil {
			return nil, err
		}
	}
	// return new sstable
	return sst, nil
}

func ScanEntries(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if len(data) < 4 {
		return 0, nil, nil
	}
	// read length
	sz := binary.BigEndian.Uint32(data[0:4])
	if len(data[4:]) < int(sz) {
		return 0, nil, nil
	}
	// read value
	token = data[4 : 4+sz]
	return int(4 + sz), token, nil
}

func ReadSSTable(fp *os.File, entries *EntrySet) error {
	if !strings.HasSuffix(fp.Name(), ".sst") {
		return errors.New("cannot read sstable file, bad suffix")
	}
	s := bufio.NewScanner(fp)
	s.Split(ScanEntries)
	var ent Entry
	for s.Scan() {
		ent.Key = s.Text()

		entries = append(entries,
	}
	for {
		n, err := fp.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Panic("error reading: %s", err)
		}
		klen = binary.BigEndian.Uint32(buf[0:4])
		vlen = binary.BigEndian.Uint32(buf[4:8])
		blob := make([]byte, klen+vlen)
		_, err = fp.Read(blob)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Panic("error reading: %s", err)
		}

	}
	return nil
}

func WriteSSTable(fp *os.File, entries *EntrySet) error {
	if !strings.HasSuffix(fp.Name(), ".sst") {
		return errors.New("cannot write sstable file, bad suffix")
	}
	if !sort.IsSorted(entries) {
		sort.Stable(entries)
	}
	entries.Range(
		func(e Entry) bool {
			_, err := e.WriteTo(fp)
			return err == nil
		},
	)
	return fp.Sync()
}

*/
