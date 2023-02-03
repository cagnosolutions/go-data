package bytes

import (
	"encoding/binary"
	"fmt"
)

type BytesQueue struct {
	data    []byte
	entries int
	size    int
}

func (q *BytesQueue) String() string {
	ss := fmt.Sprintf("entries=%d\n", q.entries)
	ss += fmt.Sprintf("size=%d\n", q.size)
	ss += fmt.Sprintf("data=%q\n", q.data)
	return ss
}

func (q *BytesQueue) Push(val []byte) {
	if q.entries*9 >= len(q.data) {
		q.grow()
	}

	size := make([]byte, 8)
	binary.LittleEndian.PutUint64(size, uint64(len(val)))
	q.data = append(q.data[:q.entries*9], append(size, val...)...)
	q.entries++
}

func (q *BytesQueue) Pop() []byte {
	if q.entries == 0 {
		return nil
	}

	size := binary.LittleEndian.Uint64(q.data[:8])
	val := q.data[8 : 8+size]
	q.data = append(q.data[:0], q.data[9+size:]...)
	q.entries--
	return val
}

func (q *BytesQueue) Insert(val []byte, pos int) {
	if q.entries*9 >= len(q.data) {
		q.grow()
	}

	size := make([]byte, 8)
	binary.LittleEndian.PutUint64(size, uint64(len(val)))
	q.data = append(q.data[:pos], append(size, val...)...)
	q.entries++
}

func (q *BytesQueue) InsertAt(val []byte, pos int) {
	q.Insert(val, pos)
}

func (q *BytesQueue) Remove(pos int, size int) {
	q.data = append(q.data[:pos], q.data[pos+size:]...)
	q.entries--
}

func (q *BytesQueue) RemoveAt(pos int, size int) {
	q.Remove(pos, size)
}

func (q *BytesQueue) grow() {
	newData := make([]byte, len(q.data)*2)
	copy(newData, q.data)
	q.data = newData
}
