package storage

import (
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/engine/page"
)

func TestDiskStore_Open(t *testing.T) {
	var fm *DiskStore
	if fm != nil {
		t.Errorf("open: io manager should be nil, got %v", fm)
	}
	fm, err := Open("my-test-io.txt")
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

func TestDiskStore_AllocatePage(t *testing.T) {
	fm, err := Open("my-test-io.txt")
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

	var pages []page.PageID
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

func TestDiskStore_WritePage(t *testing.T) {
	fm, err := Open("my-test-io.txt")
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

	var pages []page.PageID
	for i := 0; i < 8; i++ {
		pages = append(pages, fm.AllocatePage())
	}
	if len(pages) != 8 {
		t.Errorf("write: error did not allocated 8 pages, got %d", len(pages))
	}
	fmt.Printf("page id's allocated: %v\n", pages)

	for _, pid := range pages {
		pg := page.NewPage(pid, page.P_USED)
		rk := []byte(fmt.Sprintf("%.4d", pid))
		rv := []byte(fmt.Sprintf("some data for page #%.4d", pid))
		_, err = pg.AddRecord(page.NewRecord(page.R_STR, page.R_STR, rk, rv))
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

func TestDiskStore_ReadPage(t *testing.T) {
	fm, err := Open("my-test-io.txt")
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

	var pages []page.PageID
	for i := 0; i < 8; i++ {
		pages = append(pages, fm.AllocatePage())
	}
	if len(pages) != 8 {
		t.Errorf("read: error did not allocated 8 pages, got %d", len(pages))
	}
	fmt.Printf("page id's allocated: %v\n", pages)

	for _, pid := range pages {
		pg := page.NewPage(pid, page.P_USED)
		rk := []byte(fmt.Sprintf("%.4d", pid))
		rv := []byte(fmt.Sprintf("some data for page #%.4d", pid))
		_, err = pg.AddRecord(page.NewRecord(page.R_STR, page.R_STR, rk, rv))
		if err != nil {
			t.Errorf("read: error writing page record: %s", err)
		}
		err = fm.WritePage(pid, pg)
		if err != nil {
			t.Errorf("read: error writing page: %s", err)
		}
	}

	for _, pid := range pages {
		pg := page.NewPage(pid, page.P_USED)
		err = fm.ReadPage(pid, pg)
		if err != nil {
			t.Errorf("read: error writing page: %s", err)
		}
		if pg == nil {
			t.Errorf("read: page should not be nil")
		}
		fmt.Printf("page header: %+v\n", pg.GetPageHeader())
	}

	err = fm.Close()
	if err != nil {
		t.Errorf("read: error closing io: %s", err)
	}
}

func TestDiskStore_DeallocatePage(t *testing.T) {
	fm, err := Open("my-test-io.txt")
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

	var pages []page.PageID
	for i := 0; i < 8; i++ {
		pages = append(pages, fm.AllocatePage())
	}
	if len(pages) != 8 {
		t.Errorf("read: error did not allocated 8 pages, got %d", len(pages))
	}
	fmt.Printf("page id's allocated: %v\n", pages)

	for _, pid := range pages {
		pg := page.NewPage(pid, page.P_USED)
		rk := []byte(fmt.Sprintf("%.4d", pid))
		rv := []byte(fmt.Sprintf("some data for page #%.4d", pid))
		_, err = pg.AddRecord(page.NewRecord(page.R_STR, page.R_STR, rk, rv))
		if err != nil {
			t.Errorf("read: error writing page record: %s", err)
		}
		err = fm.WritePage(pid, pg)
		if err != nil {
			t.Errorf("read: error writing page: %s", err)
		}
	}

	for _, pid := range pages {
		pg := page.NewPage(pid, page.P_USED)
		err = fm.ReadPage(pid, pg)
		if err != nil {
			t.Errorf("read: error reading page: %s", err)
		}
		if pg == nil {
			t.Errorf("read: page should not be nil")
		}
		fmt.Printf("Page %d:\n %s\n\n", pg.GetPageID(), hex.Dump(pg))
		// fmt.Printf("page header: %+v\n", pg.GetPageHeader())
	}

	for _, pid := range pages {
		err = fm.DeallocatePage(pid)
		if err != nil {
			t.Errorf("dealloc: error %s", err)
		}
	}

	for _, pid := range pages {
		pg := page.NewPage(pid, page.P_USED)
		err = fm.ReadPage(pid, pg)
		if err != nil {
			t.Errorf("read: error writing page: %s", err)
		}
		if pg == nil {
			t.Errorf("read: page should not be nil")
		}
		fmt.Printf("Page %d:\n %s\n\n", pg.GetPageID(), hex.Dump(pg))
	}

	err = fm.Close()
	if err != nil {
		t.Errorf("read: error closing io: %s", err)
	}
}
