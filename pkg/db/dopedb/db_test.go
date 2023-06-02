package dopedb

import (
	"fmt"
	"testing"
)

func TestDBHasher(t *testing.T) {
	db := NewDB()

	var n uint32

	for i := 0; i < 3; i++ {
		for _, ss := range []string{"foo", "foo1", "foo2", "bar", "baz"} {
			n = db.hasher(ss)
			fmt.Printf("%q=%d\n", ss, n)
		}
		fmt.Println()
	}

}
