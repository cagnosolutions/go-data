package main

import (
	"encoding/binary"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func OpenIndexes(base string) []Index {
	var idx []Index
	err := filepath.WalkDir(
		base, func(path string, d fs.DirEntry, err error) error {
			if d.Name() == "index.db" {
				idx = append(idx, OpenIndex(filepath.Dir(path)))
			}
			return nil
		},
	)
	if err != nil {
		panic(err)
	}
	return idx
}

type Result struct {
	Path string
	Data IndexEntry
}

func SearchIndexes(key string, indexes ...Index) *Result {
	c := make(chan *Result)
	searchIndex := func(i int) { c <- indexes[i].Search(key) }
	for i := range indexes {
		go searchIndex(i)
	}
	return <-c
}

type Index struct {
	Path string
	Data []IndexEntry
}

func OpenIndex(base string) Index {
	return Index{
		Path: base,
		Data: ReadIndex(base),
	}
}

func (idx *Index) Search(key string) *Result {
	if !sort.IsSorted(idx) {
		sort.Stable(idx)
	}
	i, found := sort.Find(
		idx.Len(), func(i int) int {
			return strings.Compare(key, idx.Data[i].Key)
		},
	)
	if !found {
		return nil
	}
	return &Result{
		Path: idx.Path,
		Data: idx.Data[i],
	}
}

func (idx *Index) Len() int {
	return len(idx.Data)
}

func (idx *Index) Less(i, j int) bool {
	return idx.Data[i].Key < idx.Data[j].Key
}

func (idx *Index) Swap(i, j int) {
	idx.Data[i], idx.Data[j] = idx.Data[j], idx.Data[i]
}

type IndexEntry struct {
	Key string
	Off uint32
}

func ReadIndex(base string) (index []IndexEntry) {
	// open the index file
	fp, err := os.Open(filepath.Join(base, "index.db"))
	if err != nil {
		panic(err)
	}
	// create buffer
	buf := make([]byte, 8)
	var klen, offs uint32
	for {
		// read chunk
		_, err = fp.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		// decode key length
		klen = binary.BigEndian.Uint32(buf[0:4])
		// decode entry offset
		offs = binary.BigEndian.Uint32(buf[4:8])
		// read key
		key := make([]byte, klen)
		_, err = fp.Read(key)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		// add to index
		index = append(
			index, IndexEntry{
				Key: string(key),
				Off: offs,
			},
		)
	}
	err = fp.Close()
	if err != nil {
		panic(err)
	}
	return index
}

func WriteIndex(base string, index []IndexEntry) error {
	// open and write the index file
	fp, err := os.Create(filepath.Join(base, "index.db"))
	if err != nil {
		return err
	}

	// write key lengths and entry offsets
	for _, ie := range index {

		// write key length
		err = binary.Write(fp, binary.BigEndian, uint32(len(ie.Key)))
		if err != nil {
			return err
		}

		// write offset
		err = binary.Write(fp, binary.BigEndian, ie.Off)
		if err != nil {
			return err
		}

		// write key
		_, err = fp.WriteString(ie.Key)
		if err != nil {
			return err
		}

	}

	// flush and close the index file
	err = fp.Sync()
	if err != nil {
		return err
	}

	err = fp.Close()
	if err != nil {
		return err
	}

	return nil
}
