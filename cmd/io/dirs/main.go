package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() {

	dir := "cmd/io/dirs/files/test"

	setupFakeFileStructure(dir)

	index, err := GetSegmentIDs(dir)
	if err != nil {
		panic(err)
	}
	fmt.Println("index info gathered:")
	for _, info := range index {
		fmt.Println(info)
	}

	// teardownFakeFileStructure(dir)
}

const (
	currentSegment = "dat-current.seg" // "seg-current.db"
	segmentPrefix  = "dat-"
	segmentSuffix  = ".seg"
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

func setupFakeFileStructure(path string) {
	// clean path
	root := clean(path)
	// make dirs
	err := os.MkdirAll(root, 0644)
	if err != nil {
		panic(err)
	}
	var fd *os.File
	// create files
	for i := 0; i < 4096; i += 512 {
		fd, err = os.Create(filepath.Join(root, MakeFileNameFromID(int64(i))))
		if err != nil {
			panic(err)
		}
		err = fd.Close()
		if err != nil {
			panic(err)
		}
	}
	// make a "current" segment
	fd, err = os.Create(filepath.Join(root, currentSegment))
	if err != nil {
		panic(err)
	}
	err = fd.Close()
	if err != nil {
		panic(err)
	}
}

func clean(path string) string {
	full, err := filepath.Abs(filepath.ToSlash(path))
	if err != nil {
		panic("clean: " + err.Error())
	}
	return full
}

func setup() {
	f, err := os.CreateTemp("", "foo")
	if err != nil {
		panic(err)
	}
	f.WriteString("foo bar")
	name := f.Name()
	fi, err := f.Stat()
	if err != nil {
		panic(err)
	}
	size := fi.Size()
	err = f.Close()
	if err != nil {
		panic(err)
	}
	newSize := int64(200)
	err = os.Truncate(name, newSize)
	if err != nil {
		panic(err)
	}
	fi, err = os.Stat(name)
	if err != nil {
		panic(err)
	}
	check := fi.Size()
	data, err := os.ReadFile(name)
	if err != nil {
		panic(err)
	}
	fmt.Printf("old size=%d, new size=%d, check=%d\ndata=%q\n", size, newSize, check, data)
}
