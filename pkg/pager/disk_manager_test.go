package pager

import (
	"fmt"
	"os"
	"testing"
)

var testFile = "testing/diskmanager.db"
var err error

func TestDiskManager(t *testing.T) {
	dm := newDiskManager(testFile)
	fmt.Println(dm.getFreePageIDs())
	for i := 0; i < 15; i++ {
		fmt.Println(dm.allocate())
	}
	fmt.Println(dm.getFreePageIDs())
	for i := 0; i < 15; i++ {
		fmt.Println(dm.allocate())
	}
	fmt.Println(dm.getFreePageIDs())
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
