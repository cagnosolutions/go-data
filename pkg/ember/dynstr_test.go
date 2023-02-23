package ember

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func TestSizes(t *testing.T) {
	ds := dynstr{1, 2, []byte{'1', '2', '3'}}
	fmt.Printf("size of dynstr (in memory): %d\n", ds.sizeInMemory())
	fmt.Printf("size of dynstr (on disk): %d\n", ds.sizeOnDisk())
	fmt.Printf("%s\n", getKindAsString(dsKindBin+2))

	fmt.Println(max64, max32, max16, max8)

	n := (max16 >> 1) - 8
	b := make([]byte, 2)
	binary.PutVarint(b, int64(n))

	fmt.Printf("%d, [% x]\n", n, b)

	buf := make([]byte, binary.MaxVarintLen64)
	for _, x := range []int64{-65, -64, -2, -1, 0, 1, 2, 63, 64} {
		n := binary.PutVarint(buf, x)
		fmt.Printf("%x\n", buf[:n])
	}
}
