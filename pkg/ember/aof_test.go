package ember

import (
	"fmt"
	"testing"
	"time"
)

func TestAOFPrune(t *testing.T) {

	pause := 5 * time.Second

	// open an aof
	f, err := open("testing/data.aof")
	if err != nil {
		t.Errorf("open: %s", err)
	}

	printAndPause("writing some data...", pause)

	// write some data
	for i := 0; i < 1024; i++ {
		_, err = f.write([]byte(fmt.Sprintf("[%d] this is entry number %d", i, i)))
		if err != nil {
			t.Errorf("write: %s", err)
		}
	}

	printAndPause("reading some data...", pause)

	// read some data
	for i := 0; i < 1024; i += 8 {
		got, err := f.read(i)
		if err != nil {
			t.Errorf("read: %s", err)
		}
		fmt.Printf("read(%d)=%q\n", i, got)
	}

	printAndPause("deleting some data...", pause)

	// delete some data
	for i := 0; i < 1024; i += 2 {
		err = f.delete(i)
		if err != nil {
			t.Errorf("delete: %s", err)
		}
	}

	printAndPause("pruning file...", pause)

	// prune
	err = f.pruneFile()
	if err != nil {
		t.Errorf("pruneFile: %s", err)
	}

	printAndPause("reading some (more) data (after pruning)...", pause)

	// read some data
	for i := 0; i < 512; i++ {
		got, err := f.read(i)
		if err != nil {
			t.Errorf("read: %s", err)
		}
		fmt.Printf("read->(%d)=%q\n", i, got)
	}

	printAndPause("closing...", pause)

	// close the file
	err = f.close()
	if err != nil {
		t.Errorf("close: %s", err)
	}
}

func printAndPause(s string, n time.Duration) {
	fmt.Println(s)
	time.Sleep(n)
}
