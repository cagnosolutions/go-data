package io

import (
	"fmt"
	"testing"
)

func TestDiskStore(t *testing.T) {

	// create new disk store
	ds, err := openDiskStore("diskstore-test")
	if err != nil {
		t.Errorf("error opening the disk store: " + err.Error())
	}

	// write something to the disk store
	pg := make([]byte, 64)
	copy(pg, "this is page 00")
	err = ds.write(0, pg)
	if err != nil {
		t.Errorf("error writing to the disk store / segment: " + err.Error())
	}

	copy(pg, "this is page 18")
	err = ds.write(18, pg)
	if err != nil {
		t.Errorf("error writing to the disk store / segment: " + err.Error())
	}

	// print out the disk store
	fmt.Println(ds)

	// don't forget to close the disk store
	err = ds.close()
	if err != nil {
		t.Errorf("error closing the disk store: " + err.Error())
	}

}
