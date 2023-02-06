package bytes

import (
	"fmt"
	"testing"
)

var q = new(BytesQueue)

var tableData = []struct {
	ID   int
	Data []byte
	Off  int
}{
	{
		ID:   1,
		Data: []byte("this is the first record"),
		Off:  0,
	},
	{
		ID:   2,
		Data: []byte("record-2"),
		Off:  32,
	},
	{
		ID:   3,
		Data: []byte(`{"id":3,"buf":"record number three","pos":56}`),
		Off:  77,
	},
	{
		ID:   4,
		Data: []byte("record number four is cool"),
		Off:  103,
	},
	{
		ID:   5,
		Data: []byte("this-is-rec-5"),
		Off:  5,
	},
}

func TestBytesQueue_Insert(t *testing.T) {

	q.write([]byte("foo"))
	fmt.Printf("size=%d, pos=%d, buf=%q\n", q.size, q.pos, q.buf)
	q.write([]byte("bar"))
	fmt.Printf("size=%d, pos=%d, buf=%q\n", q.size, q.pos, q.buf)
	q.write([]byte("baz"))
	fmt.Printf("size=%d, pos=%d, buf=%q\n", q.size, q.pos, q.buf)

	// for _, data := range tableData {
	// 	if err := q.encode(data.Data, data.Off); err != nil {
	// 		t.Error(err)
	// 	}
	// 	fmt.Printf("size=%d, pos=%d, buf=%q\n", q.size, q.pos, q.buf)
	// }

}
