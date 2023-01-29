package dbms

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/util"
)

func createFile(name string, b *testing.B) *os.File {
	b.ReportAllocs()
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		b.Fatalf("creating file: %s", err)
	}
	return f
}

func openFile(name string, b *testing.B) *os.File {
	b.ReportAllocs()
	f, err := os.OpenFile(name, os.O_RDWR, 0644)
	if err != nil {
		b.Fatalf("creating file: %s", err)
	}
	return f
}

func closeFile(f *os.File, b *testing.B) {
	b.ReportAllocs()
	err := f.Close()
	if err != nil {
		b.Fatalf("closing file: %s", err)
	}
}

// Basic stats for file creation: speed: 221,135 ns/op; 672 B/op; 3 allocs/op
func BenchmarkOSFileCreate(b *testing.B) {
	for i := 0; i < b.N; i++ {

		b.StartTimer()
		f := createFile("file.txt", b)
		b.StopTimer()

		err := f.Close()
		if err != nil {
			b.Fatalf("closing file: %s", err)
		}
		err = os.Remove("file.txt")
		if err != nil {
			b.Fatalf("removing file: %s", err)
		}
	}
}

// Basic stats for file opening: speed: 34,148 ns/op; 672 B/op; 3 allocs/op
func BenchmarkOSFileOpen(b *testing.B) {
	// create a new file
	f, err := os.OpenFile("file.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		b.Fatalf("creating file: %s", err)
	}
	// close file
	err = f.Close()
	if err != nil {
		b.Fatalf("closing file: %s", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		b.StartTimer()
		f = openFile("file.txt", b)
		b.StopTimer()

		err = f.Close()
		if err != nil {
			b.Fatalf("closing file: %s", err)
		}
	}
	// remove file
	err = os.Remove("file.txt")
	if err != nil {
		b.Fatalf("removing file: %s", err)
	}
}

func TestOSFileHandles(t *testing.T) {

	// grab memory profile and print stats
	m := new(runtime.MemStats)
	util.PrintStatsTab(*m)

	// create a new file handler
	fh := NewFileHandler()

	// print file handler status
	fmt.Println(fh)

	// print memory usage
	util.PrintStatsTab(*m)

	// create files
	fmt.Println(">> creating files...")
	for i := 1; i <= 16; i++ {
		name := fmt.Sprintf("file-%.4d.txt", i)
		fh.Open(name)
	}

	// print memory usage
	util.PrintStatsTab(*m)

	// print file handler status
	fmt.Println(fh)

	// write data in all files
	fmt.Println(">> writing data to files...")
	for i := 1; i <= 16; i++ {
		name := fmt.Sprintf("file-%.4d.txt", i)
		data := fmt.Sprintf("file is file %.4d data", i)
		fh.Write(name, []byte(data))
	}

	// print memory usage
	util.PrintStatsTab(*m)

	// close files
	fmt.Println(">> closing files...")
	for i := 1; i <= 16; i++ {
		name := fmt.Sprintf("file-%.4d.txt", i)
		fh.Close(name)
	}

	// print memory usage
	util.PrintStatsTab(*m)

	// opening files
	fmt.Println(">> re-opening files...")
	for i := 1; i <= 16; i++ {
		name := fmt.Sprintf("file-%.4d.txt", i)
		fh.Open(name)
	}

	// print memory usage
	util.PrintStatsTab(*m)

	// reading data in all files
	data := make([]byte, 24)
	fmt.Println(">> reading data from all files...")
	for i := 1; i <= 16; i++ {
		name := fmt.Sprintf("file-%.4d.txt", i)
		fh.Read(name, data)
		fmt.Printf("read from %q: %q\n", name, data)
	}

	// print memory usage
	util.PrintStatsTab(*m)

	// remove 16 files
	fmt.Println(">> removing files...")
	for i := 1; i <= 16; i++ {
		name := fmt.Sprintf("file-%.4d.txt", i)
		fh.Remove(name)
	}

	// print memory usage
	util.PrintStatsTab(*m)
	fmt.Println(fh)

	runtime.GC()
	util.PrintStatsTab(*m)
}

type FileHandler struct {
	sync.RWMutex
	files map[string]*os.File
}

func NewFileHandler() *FileHandler {
	return &FileHandler{
		files: make(map[string]*os.File),
	}
}

func (fh *FileHandler) Open(name string) {
	// write lock
	fh.Lock()
	defer fh.Unlock()
	var err error
	// look for file in map
	fp, found := fh.files[name]
	if found {
		// ensure it is open
		if fp != nil {
			// it is open, just return
			return
		}
		// it exists, but is closed
	}
	// check to see if we should create or open
	_, err = os.Stat(name)
	if os.IsNotExist(err) {
		// create file
		fp, err = os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_RDWR|os.O_SYNC, 0644)
		if err != nil {
			log.Panicf("open: error creating file: %s", err)
		}
	} else {
		// open file
		fp, err = os.OpenFile(name, os.O_RDWR|os.O_SYNC, 0644)
		if err != nil {
			log.Panicf("open: error creating file: %s", err)
		}
	}
	// add the fp to the map
	fh.files[name] = fp
}

func (fh *FileHandler) Write(name string, p []byte) {
	// write lock
	fh.Lock()
	defer fh.Unlock()
	var err error
	// look for file in map
	fp, found := fh.files[name]
	if found {
		// ensure it is open
		if fp == nil {
			// it is not open
			log.Panicf("write: file found, but not open")
			return
		}
		// it is open, continue
	}
	// write data to fp
	_, err = fp.Write(p)
	if err != nil {
		log.Panicf("write: error writing: %s", err)
	}
	// sync data to fp
	err = fp.Sync()
	if err != nil {
		log.Panicf("write: error syncing: %s", err)
	}
	// done, return
	return
}

func (fh *FileHandler) Read(name string, p []byte) {
	// write lock
	fh.Lock()
	defer fh.Unlock()
	var err error
	// look for file in map
	fp, found := fh.files[name]
	if found {
		// ensure it is open
		if fp == nil {
			// it is not open
			log.Panicf("read: file found, but not open")
			return
		}
		// it is open, continue
	}
	// read data from fp
	_, err = fp.Read(p)
	if err != nil {
		log.Panicf("read: error reading: %s", err)
	}
	// done, return
	return
}

func (fh *FileHandler) Close(name string) {
	// write lock
	fh.Lock()
	defer fh.Unlock()
	var err error
	// look for file in map
	fp, found := fh.files[name]
	if !found {
		// not found
		return
	}
	// ensure it is closed
	if fp == nil {
		// it is closed, just return
		return
	}
	// close file
	err = fp.Close()
	if err != nil {
		log.Panicf("close: error closing: %s", err)
	}
	// set the fp to nil
	fh.files[name] = nil
}

func (fh *FileHandler) Remove(name string) {
	// write lock
	fh.Lock()
	defer fh.Unlock()
	var err error
	// look for file in map
	fp, found := fh.files[name]
	if !found {
		// not found
		return
	}
	// ensure it is closed
	if fp != nil {
		// not closed, close it
		err = fp.Close()
		if err != nil {
			log.Panicf("close: error closing: %s", err)
		}
		// set the fp to nil
		fh.files[name] = nil
	}
	// remove from map
	delete(fh.files, name)
	// delete it from the filesystem
	err = os.Remove(name)
	if err != nil {
		log.Panicf("close: removing file: %s", err)
	}
}

func (fh *FileHandler) ListFiles() []string {
	// read lock
	fh.RLock()
	defer fh.RUnlock()
	var ss []string
	var open bool
	var size int64
	var fd uintptr
	// iterate over map
	for name, fp := range fh.files {
		if fp != nil {
			open = true
			fi, err := fp.Stat()
			if err != nil {
				log.Panicf("list: error getting file stats: %s", err)
			}
			size = fi.Size()
			fd = fp.Fd()
		} else {
			open = false
			size = -1
			fd = 0
		}
		if fd != 0 {
			ss = append(ss, fmt.Sprintf("name=%q, size=%d, closed=%v, fd=%v", name, size, !open, fd))
		} else {
			ss = append(ss, fmt.Sprintf("name=%q, size=%d, closed=%v, fd=%v", name, size, !open, "??"))
		}
	}
	// return file list
	return ss
}

func (fh *FileHandler) String() string {
	ss := fmt.Sprintf("files:\n")
	for _, s := range fh.ListFiles() {
		ss += fmt.Sprintf("\t%s\n", s)
	}
	return ss
}
