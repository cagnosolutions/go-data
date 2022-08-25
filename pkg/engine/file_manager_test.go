package engine

import (
	"fmt"
	"os"
	"testing"
)

func TestFileManager_OpenFileManager(t *testing.T) {
	var fm *FileManager
	if fm != nil {
		t.Errorf("open: file manager should be nil, got %v", fm)
	}
	fm, err := OpenFileManager("my-test-file.txt")
	if err != nil {
		t.Errorf("open: file manager open error: %s", err)
	}
	defer func() {
		err := os.Remove("my-test-file.txt")
		if err != nil {
			t.Errorf("open: error removing file: %s", err)
		}
	}()
	if fm == nil {
		t.Errorf("open: file manager should NOT be nil, got %v", fm)
	}
	err = fm.Close()
	if err != nil {
		t.Errorf("open: error closing file: %s", err)
	}
}

func TestFileManager_AllocatePage(t *testing.T) {
	fm, err := OpenFileManager("my-test-file.txt")
	if err != nil {
		t.Errorf("allocate: file manager open error: %s", err)
	}
	defer func() {
		err := os.Remove("my-test-file.txt")
		if err != nil {
			t.Errorf("allocate: error removing file: %s", err)
		}
	}()
	if fm == nil {
		t.Errorf("allocate: file manager should NOT be nil, got %v", fm)
	}

	var pages []PageID
	for i := 0; i < 8; i++ {
		pages = append(pages, fm.AllocatePage())
	}
	if len(pages) != 8 {
		t.Errorf("allocate: error did not allocated 8 pages, got %d", len(pages))
	}

	err = fm.Close()
	if err != nil {
		t.Errorf("allocate: error closing file: %s", err)
	}
}

func TestFileManager_WritePage(t *testing.T) {
	fm, err := OpenFileManager("my-test-file.txt")
	if err != nil {
		t.Errorf("write: file manager open error: %s", err)
	}
	defer func() {
		err := os.Remove("my-test-file.txt")
		if err != nil {
			t.Errorf("write: error removing file: %s", err)
		}
	}()
	if fm == nil {
		t.Errorf("write: file manager should NOT be nil, got %v", fm)
	}

	var pages []PageID
	for i := 0; i < 8; i++ {
		pages = append(pages, fm.AllocatePage())
	}
	if len(pages) != 8 {
		t.Errorf("write: error did not allocated 8 pages, got %d", len(pages))
	}
	fmt.Printf("page id's allocated: %v\n", pages)

	for _, pid := range pages {
		pg := newPage(pid)
		_, err = pg.addRecord([]byte(fmt.Sprintf("some data for page #%.4d", pid)))
		if err != nil {
			t.Errorf("write: error writing page record: %s", err)
		}
		err = fm.WritePage(pid, pg)
		if err != nil {
			t.Errorf("write: error writing page: %s", err)
		}
	}

	err = fm.Close()
	if err != nil {
		t.Errorf("write: error closing file: %s", err)
	}
}

func TestFileManager_ReadPage(t *testing.T) {
	fm, err := OpenFileManager("my-test-file.txt")
	if err != nil {
		t.Errorf("read: file manager open error: %s", err)
	}
	defer func() {
		err := os.Remove("my-test-file.txt")
		if err != nil {
			t.Errorf("read: error removing file: %s", err)
		}
	}()
	if fm == nil {
		t.Errorf("read: file manager should NOT be nil, got %v", fm)
	}

	var pages []PageID
	for i := 0; i < 8; i++ {
		pages = append(pages, fm.AllocatePage())
	}
	if len(pages) != 8 {
		t.Errorf("read: error did not allocated 8 pages, got %d", len(pages))
	}
	fmt.Printf("page id's allocated: %v\n", pages)

	for _, pid := range pages {
		pg := newPage(pid)
		_, err = pg.addRecord([]byte(fmt.Sprintf("some data for page #%.4d", pid)))
		if err != nil {
			t.Errorf("read: error writing page record: %s", err)
		}
		err = fm.WritePage(pid, pg)
		if err != nil {
			t.Errorf("read: error writing page: %s", err)
		}
	}

	for _, pid := range pages {
		pg := newPage(pid)
		err = fm.ReadPage(pid, pg)
		if err != nil {
			t.Errorf("read: error writing page: %s", err)
		}
		if pg == nil {
			t.Errorf("read: page should not be nil")
		}
		fmt.Printf("page header: %+v\n", pg.GetHeader())
	}

	err = fm.Close()
	if err != nil {
		t.Errorf("read: error closing file: %s", err)
	}
}
