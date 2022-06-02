package pager

import (
	"fmt"
	"os"
	"testing"
)

var testFile = "testing/diskmanager.db"

func TestDiskManager(t *testing.T) {
	var err error
	dm := newDiskManager(testFile)
	err = dm.load()
	if err != nil {
		panic(err)
	}
	free := dm.getFreePageIDs()
	fmt.Println(free)
	// close and clean
	clean(dm)
}

func clean(dm *diskManager) {
	err := dm.close()
	if err != nil {
		panic(err)
	}
	err = os.Remove(testFile)
	if err != nil {
		panic(err)
	}
}
