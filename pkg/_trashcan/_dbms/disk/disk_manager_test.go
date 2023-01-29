package disk

import (
	"testing"

	"github.com/cagnosolutions/go-data/pkg/dbms/page"
	"github.com/cagnosolutions/go-data/pkg/util"
)

func TestReadWritePage(t *testing.T) {
	dm := NewDiskManagerTest()
	defer dm.ShutDown()

	data := make([]byte, page.PageSize)
	buffer := make([]byte, page.PageSize)

	copy(data, "A test string.")

	dm.ReadPage(0, buffer) // tolerate empty read

	err := dm.WritePage(0, data)
	if err != nil {
		t.Errorf("error writing page 0: %s\n", err)
	}
	err = dm.ReadPage(0, buffer)
	if err != nil {
		t.Errorf("error reading page 0: %s\n", err)
	}
	util.Equals(t, int64(4096), dm.Size())
	util.Equals(t, data, buffer)

	memset(buffer, 0)
	copy(data, "Another test string.")

	err = dm.WritePage(5, data)
	if err != nil {
		t.Errorf("error writing page 5: %s\n", err)
	}
	err = dm.ReadPage(5, buffer)
	if err != nil {
		t.Errorf("error reading page 5: %s\n", err)
	}
	util.Equals(t, data, buffer)

	// the size of disk is 24576 bytes because we have 6 pages
	util.Equals(t, int64(24576), dm.Size())
}

func memset(buffer []byte, value int) {
	for i := range buffer {
		buffer[i] = 0
	}
}
