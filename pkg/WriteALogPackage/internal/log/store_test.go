package log

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/util"
)

var (
	write = []byte("hello world")
	width = uint64(len(write)) + lenWidth
)

func TestStoreAppendRead(t *testing.T) {
	f, err := ioutil.TempFile("", "store_append_read_test")
	util.AssertNoError(t, err)
	defer os.Remove(f.Name())

	s, err := newStore(f)
	util.AssertNoError(t, err)

	testAppend(t, s)
	testRead(t, s)
	testReadAt(t, s)

	s, err = newStore(f)
	util.AssertNoError(t, err)
	testRead(t, s)
}

// END: intro

// START: end
func testAppend(t *testing.T, s *store) {
	t.Helper()
	for i := uint64(1); i < 4; i++ {
		n, pos, err := s.Append(write)
		util.AssertNoError(t, err)
		util.AssertEqual(t, pos+n, width*i)
	}
}

func testRead(t *testing.T, s *store) {
	t.Helper()
	var pos uint64
	for i := uint64(1); i < 4; i++ {
		read, err := s.Read(pos)
		util.AssertNoError(t, err)
		util.AssertEqual(t, write, read)
		pos += width
	}
}

func testReadAt(t *testing.T, s *store) {
	t.Helper()
	for i, off := uint64(1), int64(0); i < 4; i++ {
		b := make([]byte, lenWidth)
		n, err := s.ReadAt(b, off)
		util.AssertNoError(t, err)
		util.AssertEqual(t, lenWidth, n)
		off += int64(n)

		size := enc.Uint64(b)
		b = make([]byte, size)
		n, err = s.ReadAt(b, off)
		util.AssertNoError(t, err)
		util.AssertEqual(t, write, b)
		util.AssertEqual(t, int(size), n)
		off += int64(n)
	}
}

// END: end

// START: close
func TestStoreClose(t *testing.T) {
	f, err := ioutil.TempFile("", "store_close_test")
	util.AssertNoError(t, err)
	defer os.Remove(f.Name())
	s, err := newStore(f)
	util.AssertNoError(t, err)
	_, _, err = s.Append(write)
	util.AssertNoError(t, err)

	f, beforeSize, err := openFile(f.Name())
	util.AssertNoError(t, err)

	err = s.Close()
	util.AssertNoError(t, err)

	f, afterSize, err := openFile(f.Name())
	util.AssertTrue(t, afterSize > beforeSize)
}

func openFile(name string) (file *os.File, size int64, err error) {
	f, err := os.OpenFile(
		name,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, 0, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}
	return f, fi.Size(), nil
}
