package ssstore

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/util"
)

func TestIndex(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "index_test")
	util.AssertNoError(t, err)
	defer os.Remove(f.Name())

	c := Config{}
	c.Segment.MaxIndexBytes = 1024
	idx, err := newIndex(f, c)
	util.AssertNoError(t, err)
	_, _, err = idx.Read(-1)
	util.AssertError(t, err)
	util.AssertEqual(t, f.Name(), idx.Name())

	entries := []struct {
		Off uint32
		Pos uint64
	}{
		{Off: 0, Pos: 0},
		{Off: 1, Pos: 10},
	}
	// END: intro

	// START: end
	for _, want := range entries {
		err = idx.Write(want.Off, want.Pos)
		util.AssertNoError(t, err)

		_, pos, err := idx.Read(int64(want.Off))
		util.AssertNoError(t, err)
		if pos != want.Pos {
			t.Errorf("bad position (%d != %d)", pos, want.Pos)
		}
		// util.AssertEqual(t, want.Pos, pos)
	}

	// index and scanner should error when reading past existing entries
	_, _, err = idx.Read(int64(len(entries)))
	if err != io.EOF {
		t.Errorf("bad error value (%v != %v)", err, io.EOF)
	}
	// util.AssertEqual(t, io.EOF, err)
	_ = idx.Close()

	// index should build its state from the existing file
	f, _ = os.OpenFile(f.Name(), os.O_RDWR, 0600)
	idx, err = newIndex(f, c)
	util.AssertNoError(t, err)
	off, pos, err := idx.Read(-1)
	util.AssertNoError(t, err)
	if off != uint32(1) {
		t.Errorf("bad offset (%d != %d)", off, uint32(1))
	}
	// util.AssertEqual(t, uint32(1), off)
	if pos != entries[1].Pos {
		t.Errorf("bad position (%d != %d)", pos, entries[1].Pos)
	}
	// util.AssertEqual(t, entries[1].Pos, pos)
}

// END: end
