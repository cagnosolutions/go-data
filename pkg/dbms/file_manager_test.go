package dbms

import (
	"fmt"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/dbms/page"
)

const basePath = "testing/current-manager"

func TestFileManager_All(t *testing.T) {

	// open a current manager instance
	fm, err := OpenFileManager(basePath)
	if err != nil {
		t.Error(err)
	}

	// allocate some pages
	var pages []page.PageID
	for i := 0; i < 64; i++ {
		pid := fm.AllocatePage()
		pages = append(pages, pid)
		fmt.Printf("allocated page %d (pages=%d, file_size=%d)\n", pid, len(pages), fm.size)
	}

	// close our current manager instance
	err = fm.Close()
}
