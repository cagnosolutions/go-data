package ember

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func TestEmberDB_items(t *testing.T) {

	i1 := &item{
		k: "item1",
		v: []byte("item1-val"),
	}

	i1b := encode(i1)
	fmt.Println(i1b)

	i2, err := decode(i1b)
	if err != nil {
		t.Errorf("decode: %s", err)
	}

	fmt.Println(i1, "\n", i2)
}

func TestEmberDB(t *testing.T) {

	// open
	fmt.Println("Opening db...")
	db, err := Open(DefaultEmberConfig)
	if err != nil {
		t.Errorf("open: %s", err)
	}

	// write some stuff
	fmt.Println("Writing some data...")
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%.4d", i)
		val := fmt.Sprintf("val-%.64d", i)

		err = db.Set(key, []byte(val))
		if err != nil {
			t.Errorf("set (%s): %s", key, err)
		}
	}

	// close
	fmt.Println("Closing...")
	err = db.Close()
	if err != nil {
		t.Errorf("close: %s", err)
	}
	fmt.Println("Closed.")

	fmt.Println("Waiting a few seconds...")
	time.Sleep(30 * time.Second)

	// open again
	fmt.Println("Opening db again...")
	db, err = Open(DefaultEmberConfig)
	if err != nil {
		t.Errorf("open: %s", err)
	}

	// read some data
	fmt.Println("Reading some data...")
	for i := 0; i < 1000; i += 2 {
		key := fmt.Sprintf("key-%.4d", i)
		v, err := db.Get(key)
		if err != nil {
			t.Errorf("get (%s): %s", key, err)
		}
		val := []byte(fmt.Sprintf("val-%.64d", i))
		if !bytes.Equal(v, val) {
			t.Errorf("got=%q, wanted=%q\n", v, val)
		}
	}

	// close again
	fmt.Println("Closing again...")
	err = db.Close()
	if err != nil {
		t.Errorf("close: %s", err)
	}
	fmt.Println("Closed.")
}
