package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func main() {
	// index := ReadIndex("cmd/lsm/data/0004")
	// for _, i := range index {
	//	fmt.Printf("%#v\n", i)
	// }

	// checkMT(3, 2, 5, 10, 17, 15, 6)
	// checkMT(7, 1, 8, 4, 11, 9, 12)
	// checkMT(19, 16, 13, 18, 14, 20)

	indexes := OpenIndexes("cmd/lsm/data")
	for i, index := range indexes {
		fmt.Printf("data in index %d:\n", i)
		for _, data := range index.Data {
			fmt.Printf("%v\n", data)
		}
		fmt.Println()
	}
	res := SearchIndexes("key-018", indexes...)
	fmt.Println(res)

}

func fanIn[T any](input1, input2 <-chan T) <-chan T {
	c := make(chan T)
	go func() {
		for {
			select {
			case v := <-input1:
				c <- v
			case v := <-input2:
				c <- v
			}

		}
	}()
	return c
}

// func ReadKeyAt(fp *os.File, offset uint32) string {
// 	// get current (what will be the original) offset
// 	original, err := fp.Seek(0, io.SeekCurrent)
// 	if err != nil {
// 		panic(err)
// 	}
// 	// seek to the provided offset
// 	_, err = fp.Seek(int64(offset), io.SeekStart)
// 	if err != nil {
// 		panic(err)
// 	}
// 	// decode the key length
// 	var klen uint32
// 	err = binary.Read(fp, binary.BigEndian, &klen)
// 	if err != nil {
// 		panic(err)
// 	}
// 	// skip the val length
// 	_, err = fp.Seek(4, io.SeekCurrent)
// 	if err != nil {
// 		panic(err)
// 	}
// 	// create a buffer for the key, and read the key
// 	key := make([]byte, klen)
// 	_, err = fp.Read(key)
// 	if err != nil {
// 		panic(err)
// 	}
// 	// seek back to the original "current" location
// 	_, err = fp.Seek(original, io.SeekStart)
// 	if err != nil {
// 		panic(err)
// 	}
// 	// return the key
// 	return string(key)
// }

// func openSST() {
//
// 	sst, err := OpenSSTable("cmd/lsm/data/0001")
// 	if err != nil {
// 		panic(err)
// 	}
// 	for _, idx := range sst.Keys {
// 		fmt.Println(idx)
// 	}
// 	fmt.Println("getting 4: ", sst.Get("4"))
// 	fmt.Println("getting 7: ", sst.Get("7"))
// 	fmt.Println("getting 2: ", sst.Get("2"))
// 	fmt.Println("getting 1: ", sst.Get("1"))
// }

func checkMT(ss ...int) {
	mt := NewMemtable()
	for _, i := range ss {
		mt.Put(fmt.Sprintf("key-%.3d", i), fmt.Sprintf("entry value #%d", i))
	}
	err := mt.Flush("cmd/lsm/data")
	if err != nil {
		panic(err)
	}
}

// func readKeyIndex(fp *os.File, off uint32) (*KeyIndex, error) {
// 	// seek to offset
// 	_, err := fp.Seek(int64(off), io.SeekStart)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// read key length
// 	var klen uint32
// 	err = binary.Read(fp, binary.BigEndian, &klen)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// skip val length
// 	_, err = fp.Seek(4, io.SeekCurrent)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// create buffer
// 	buf := make([]byte, klen)
// 	// read key
// 	_, err = fp.Read(buf)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// return key index
// 	return &KeyIndex{
// 		Key: string(buf),
// 		Off: off,
// 	}, nil
// }

func readEntryAt(fp *os.File, off uint32) (*Entry, error) {
	// seek to offset
	_, err := fp.Seek(int64(off), io.SeekStart)
	if err != nil {
		return nil, err
	}
	// read key length
	var klen, vlen uint32
	err = binary.Read(fp, binary.BigEndian, &klen)
	if err != nil {
		return nil, err
	}
	// read val length
	err = binary.Read(fp, binary.BigEndian, &vlen)
	if err != nil {
		return nil, err
	}
	// create buffer
	buf := make([]byte, klen+vlen)
	// read key and value
	_, err = fp.Read(buf)
	if err != nil {
		return nil, err
	}

	// create and return entry
	return &Entry{
		Key: string(buf[0:klen]),
		Val: string(buf[klen : klen+vlen]),
	}, nil
}
