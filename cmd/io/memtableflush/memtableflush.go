package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/cagnosolutions/go-data/pkg/tree/lsm"
)

func main() {
	mt := lsm.NewMemtable(64)

	path := "cmd/io/memtableflush/sstables"

	for i := 0; i < 64; i++ {

		isFull := mt.Put(fmt.Sprintf("key-%.4d", i), fmt.Sprintf("my-value-%.8d", i))
		if isFull {
			WriteSSTable(path, mt)
		}

	}

	ReadSSTable(path)

}

func WriteSSTable(path string, mt *lsm.Memtable) {
	log.Println("Writing SSTable...")
	name := MakeFileNameFromID(1)
	fd, err := os.Create(filepath.Join(path, name))
	// fd, err := CreateFileFromID(path, 1)
	if err != nil {
		log.Panicf("error creating file: %s\n", err)
	}
	defer func(fd *os.File) {
		err := fd.Close()
		if err != nil {
			log.Panicf("error closing: %s\n", err)
		}
	}(fd)
	offs, err := mt.Flush(io.NewOffsetWriter(fd, 0))
	if err != nil {
		log.Panicf("error flushing: %s\n", err)
	}
	fmt.Printf("offset is now: %d\n", offs)
}

func ReadSSTable(path string) {
	log.Println("Reading SSTable...")
	name := MakeFileNameFromID(1)
	fd, err := os.Open(filepath.Join(path, name))
	if err != nil {
		log.Panicf("error opening file: %s", err)
	}
	readSSTable(fd)

}

func readKey(r io.ReaderAt, off int) (key string, ptr uint16, offs int) {
	buf := make([]byte, 2)
	n, err := r.ReadAt(buf, off)
	if err != nil {
		log.Panicf("error reading key size: %s", err)
	}
	off += n

	size := binary.BigEndian.Uint16(buf)
	buf = make([]byte, size)

	n, err = r.ReadAt(buf, off)
	if err != nil {
		log.Panicf("error reading key: %s", err)
	}
	key = string(buf)
	off += n

	n, err = r.ReadAt(buf[0:2], off)
	if err != nil {
		log.Panicf("error reading key: %s", err)
	}
	ptr = binary.BigEndian.Uint16(buf[0:2])
	off += n

	return key, ptr, off
}

func readSSTable(at io.ReaderAt) {
	// decode key
	var keys []string
	var values []string

	var offs int
	szBuf := make([]byte, 2)
	var size uint16
	buf := make([]byte, 512)

	// read key
	key, ptr, off := readKey(at, offs)
	n, err := at.ReadAt(szBuf, int64(offs))
	if err != nil {
		log.Panicf("error reading at offset: %d\n", offs)
	}
	offs += n
	size = binary.BigEndian.Uint16(szBuf)

	// read key
	if int(size) > len(buf) {
		// grow slice
		buf = append(buf, make([]byte, int(size)-len(buf))...)
	}
	n, err = at.ReadAt(buf[:size], int64(offs))
	offs += n

}

const (
	currentSegment = "dat-current.seg" // "seg-current.db"
	segmentPrefix  = "dat-"
	segmentSuffix  = ".sst"
)

func MakeFileNameFromID(index int64) string {
	hexa := strconv.FormatInt(index, 16)
	return fmt.Sprintf("%s%08s%s", segmentPrefix, hexa, segmentSuffix)
}

func GetIDFromFileName(name string) (int64, error) {
	hexa := name[len(segmentPrefix) : len(name)-len(segmentSuffix)]
	return strconv.ParseInt(hexa, 16, 32)
}

func GetSegmentIDs(dir string) ([]int, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var sids []int
	for _, file := range files {
		if file.IsDir() || !strings.HasPrefix(file.Name(), segmentPrefix) {
			continue
		}
		if strings.HasPrefix(file.Name(), segmentPrefix) {
			if strings.Contains(file.Name(), currentSegment) {
				sids = append(sids, -1)
				continue
			}
			sid, err := GetIDFromFileName(file.Name())
			if err != nil {
				return nil, err
			}
			sids = append(sids, int(sid))
		}
	}
	sort.Ints(sids)
	return sids, nil
}

func teardownFakeFileStructure(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		panic(err)
	}
}

func CreateFileFromID(path string, id int) (*os.File, error) {
	// clean path
	root := clean(path)
	// make dirs
	err := os.MkdirAll(root, 0644)
	if err != nil {
		panic(err)
	}
	var fd *os.File
	// create file
	fd, err = os.Create(filepath.Join(root, MakeFileNameFromID(int64(id))))
	if err != nil {
		panic(err)
	}
	return fd, err
}

func clean(path string) string {
	full, err := filepath.Abs(filepath.ToSlash(path))
	if err != nil {
		panic("clean: " + err.Error())
	}
	return full
}
