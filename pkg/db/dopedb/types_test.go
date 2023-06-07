package dopedb

import (
	"fmt"
	"testing"
)

var testdb *DB

func init() {
	var err error
	testdb, err = NewDB(nil)
	if err != nil {
		panic(err)
	}
}

func TestStrType(t *testing.T) {

	key := "k1"

	in := Str("this is my first string")
	err := SetAs(testdb, key, &in)
	if err != nil {
		t.Errorf("SetAs failed: %s", err)
	}

	var out Str
	err = GetAs(testdb, key, &out)
	if err != nil {
		t.Errorf("GetAs failed: %s", err)
	}
	fmt.Printf("%#v\n", out)
}

func TestNumType(t *testing.T) {

	key := "k2"

	in := Num("1234.35")
	err := SetAs(testdb, key, &in)
	if err != nil {
		t.Errorf("SetAs failed: %s", err)
	}

	var out Num
	err = GetAs(testdb, key, &out)
	if err != nil {
		t.Errorf("GetAs failed: %s", err)
	}
	fmt.Printf("%#v\n", out)
}
