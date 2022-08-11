package dbms

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/dbms/page"
)

func TestBitsetIndex(t *testing.T) {
	bs := NewBitsetIndex()
	fmt.Println(bs)
	for i := uint(0); i < 32; i++ {
		bs.SetBit(i)
	}
	fmt.Println(bs)
	for i := uint(32); i < 64; i++ {
		bs.SetBit(i)
	}
	fmt.Println(bs)
	for i := uint(64); i < 92; i++ {
		bs.SetBit(i)
	}
	fmt.Println(bs)
	for i := uint(92); i < 128; i++ {
		bs.SetBit(i)
	}
	fmt.Println(bs)
	for i := uint(768); i < 1024; i++ {
		bs.SetBit(i)
	}
	fmt.Println(bs)
}

func BenchmarkBitsetIndex_GetFree(b *testing.B) {
	bs := NewBitsetIndex()
	// for i := 0; i < 16; i++ {
	// 	(*bs)[i] = ^uint64(0)
	// }
	for i := uint(0); i < 1024; i++ {
		bs.SetBit(i)
	}
	bs.UnsetBit(1023)
	// fmt.Println(bs)
	for i := 0; i < b.N; i++ {
		free := bs.GetFree()
		if free != 1023 {
			b.Error("did not find the correct free bit")
		}
	}
}

func TestFileManager(t *testing.T) {

	// open a file manager instance
	fm, err := OpenFileManager("testing/file-manager")
	if err != nil {
		t.Error(err)
	}

	// var before, after int64
	// for i := 0; i < 4; i++ {
	// 	after, err = allocateExtent(fm.file)
	// 	fmt.Printf("allocated another extent: before=%d, after=%d\n", before, after)
	// 	before = after
	// }

	// allocate some pages
	var pages []page.PageID
	for i := 0; i < 64; i++ {
		pid := fm.AllocatePage()
		pages = append(pages, pid)
		fmt.Printf("allocated page %d (pages=%d, file_size=%d)\n", pid, len(pages), fm.size)
	}

	pid := fm.AllocatePage()
	fm.WritePage(pid, page.Page("hello world"))

	// close our file manager instance
	err = fm.Close()
}

// allocateExtent grows provided file by an extent size until it reaches
// the maximum file size, at which point an error will be returned.
func allocateExtent(fd *os.File) (int64, error) {
	// get the current size of the file
	fi, err := fd.Stat()
	if err != nil {
		return -1, err
	}
	size := fi.Size()
	// check to make sure we are not at the max file segment size
	if size == maxSegmentSize {
		return size, errors.New("file has reached the max size")
	}
	// we are below the max file size, so we should have room.
	err = fd.Truncate(size + extentSize)
	if err != nil {
		return size, err
	}
	// successfully allocated an extent, now we can return the
	// updated (current) file size, and a nil error
	fi, err = fd.Stat()
	if err != nil {
		return size, err
	}
	return fi.Size(), nil
}
