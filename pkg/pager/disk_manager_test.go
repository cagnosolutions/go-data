package pager

import (
	"os"
	"testing"
)

var dManFile = "testing/dman.db"
var r1, r2, r3 *recID
var err error

func TestDiskManager(t *testing.T) {
	// create new dman
	dm := newDiskManager(dManFile)
	// fmt.Println("dman fileSize:", util.Sizeof(dm))

	// get a fresh page ID
	pid := dm.allocatePage()

	// create a new page using page ID
	pg := newPage(pid)

	// add some record data to the page
	r1, err = pg.addRecord([]byte("record-001"))
	if err != nil {
		panic(err)
	}
	r2, err = pg.addRecord([]byte("record-002"))
	if err != nil {
		panic(err)
	}
	r3, err = pg.addRecord([]byte("record-003"))
	if err != nil {
		panic(err)
	}

	// write the page
	err = dm.writePage(pid, pg)
	if err != nil {
		panic(err)
	}

	// close
	err = dm.close()
	if err != nil {
		panic(err)
	}

	// open again
	dm = newDiskManager(dManFile)

	// read the page we just wrote
	np := make(page, szPg)
	err = dm.readPage(pid, np)
	if err != nil {
		panic(err)
	}

	// print the page data
	// fmt.Println(np.DumpPage(false))

	// read the data from the page
	r1d, err := np.getRecord(r1)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("record 1, data: %q\n", r1d)
	_ = r1d

	r2d, err := np.getRecord(r2)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("record 2, data: %q\n", r2d)
	_ = r2d

	r3d, err := np.getRecord(r3)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("record 3, data: %q\n", r3d)
	_ = r3d

	// deallocate page
	err = dm.deallocatePage(pid)
	if err != nil {
		panic(err)
	}

	// read the page we just wrote (again, after deallocating)
	np2 := make(page, szPg)
	err = dm.readPage(pid, np2)
	if err != nil {
		panic(err)
	}

	// print the page data
	// fmt.Println(np2.DumpPage(false))

	// list any free pages
	// fmt.Println("free:", dm.getFreePages())

	// close and clean
	cnc(dm)
}

func cnc(dm *diskManager) {
	err = dm.close()
	if err != nil {
		panic(err)
	}
	err = os.Remove(dManFile)
	if err != nil {
		panic(err)
	}
}
