package pager

import (
	"fmt"
	"os"
	"testing"

	"github.com/cagnosolutions/go-data/pkg/util"
)

var testFile = "testing/diskmanager.db"
var pagesInSeg = 1
var err error

func TestDiskManager(t *testing.T) {
	dm := newDiskManager(testFile, pagesInSeg)
	fmt.Println("disk manager size:", util.Sizeof(dm))
	var ids []pageID
	fmt.Println(dm.getFreePageIDs())
	for i := 0; i < 15; i++ {
		ids = append(ids, dm.allocate())
	}
	fmt.Println(dm.getFreePageIDs())
	for i := 0; i < 15; i++ {
		ids = append(ids, dm.allocate())
	}
	fmt.Println(dm.getFreePageIDs())
	fmt.Println(ids)
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
