package engine

import (
	"fmt"
	"os"
	"testing"
)

func TestDiskManager_OpenDiskManager(t *testing.T) {
	var fm *DiskManager
	if fm != nil {
		t.Errorf("open: io manager should be nil, got %v", fm)
	}
	fm, err := OpenDiskManager("my-test-io.txt")
	if err != nil {
		t.Errorf("open: io manager open error: %s", err)
	}
	defer func() {
		err := os.Remove("my-test-io.txt")
		if err != nil {
			t.Errorf("open: error removing io: %s", err)
		}
	}()
	if fm == nil {
		t.Errorf("open: io manager should NOT be nil, got %v", fm)
	}
	err = fm.Close()
	if err != nil {
		t.Errorf("open: error closing io: %s", err)
	}
}

func TestDiskManager_AllocatePage(t *testing.T) {
	fm, err := OpenDiskManager("my-test-io.txt")
	if err != nil {
		t.Errorf("allocate: io manager open error: %s", err)
	}
	defer func() {
		err := os.Remove("my-test-io.txt")
		if err != nil {
			t.Errorf("allocate: error removing io: %s", err)
		}
	}()
	if fm == nil {
		t.Errorf("allocate: io manager should NOT be nil, got %v", fm)
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
		t.Errorf("allocate: error closing io: %s", err)
	}
}

func TestDiskManager_WritePage(t *testing.T) {
	fm, err := OpenDiskManager("my-test-io.txt")
	if err != nil {
		t.Errorf("write: io manager open error: %s", err)
	}
	defer func() {
		err := os.Remove("my-test-io.txt")
		if err != nil {
			t.Errorf("write: error removing io: %s", err)
		}
	}()
	if fm == nil {
		t.Errorf("write: io manager should NOT be nil, got %v", fm)
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
		pg := newPage(uint32(pid), P_USED)
		rk := []byte(fmt.Sprintf("%.4d", pid))
		rv := []byte(fmt.Sprintf("some data for page #%.4d", pid))
		_, err = pg.addRecord(NewRecord(R_STR_STR, rk, rv))
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
		t.Errorf("write: error closing io: %s", err)
	}
}

func TestDiskManager_ReadPage(t *testing.T) {
	fm, err := OpenDiskManager("my-test-io.txt")
	if err != nil {
		t.Errorf("read: io manager open error: %s", err)
	}
	defer func() {
		err := os.Remove("my-test-io.txt")
		if err != nil {
			t.Errorf("read: error removing io: %s", err)
		}
	}()
	if fm == nil {
		t.Errorf("read: io manager should NOT be nil, got %v", fm)
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
		pg := newPage(uint32(pid), P_USED)
		rk := []byte(fmt.Sprintf("%.4d", pid))
		rv := []byte(fmt.Sprintf("some data for page #%.4d", pid))
		_, err = pg.addRecord(NewRecord(R_STR_STR, rk, rv))
		if err != nil {
			t.Errorf("read: error writing page record: %s", err)
		}
		err = fm.WritePage(pid, pg)
		if err != nil {
			t.Errorf("read: error writing page: %s", err)
		}
	}

	for _, pid := range pages {
		pg := newPage(uint32(pid), P_USED)
		err = fm.ReadPage(pid, pg)
		if err != nil {
			t.Errorf("read: error writing page: %s", err)
		}
		if pg == nil {
			t.Errorf("read: page should not be nil")
		}
		fmt.Printf("page header: %+v\n", pg.getPageHeader())
	}

	err = fm.Close()
	if err != nil {
		t.Errorf("read: error closing io: %s", err)
	}
}
