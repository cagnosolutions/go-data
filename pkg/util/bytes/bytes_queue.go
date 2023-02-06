package bytes

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

const (
	defaultSize = 64
	headerSize  = 8
)

type BytesQueue struct {
	buf     []byte
	entries int // entries is the number of entries in the queue
	pos     int // read at &buf[pos], write at &buf[len(buf)]
	size    int // size is like len
}

func (q *BytesQueue) insert(p []byte) {
	// check to see if we need to initialize or grow the buffer
	q.checkGrow()
	// copy p into our buffer at the specified position
	sz := copy(q.buf[q.pos:], p)
	if sz == 0 {
		panic("insert: did not copy any bytes")
	}
	q.pos += sz
}

func (q *BytesQueue) String() string {
	q.checkGrow()
	ss := fmt.Sprintf("entries=%d, pos=%d, buf:\n", q.entries, q.pos)
	var n, pos int
	for pos < cap(q.buf) {
		n++
		p, err := q.decode(pos)
		if err != nil && err == io.EOF {
			break
		}
		ss += fmt.Sprintf("entry[%d]=%q\n", n, p)
		pos = q.next(pos)
	}
	return ss
}

func (q *BytesQueue) Push(val []byte) {
	q.checkGrow()

	size := make([]byte, 8)
	binary.LittleEndian.PutUint64(size, uint64(len(val)))
	q.buf = append(q.buf[:q.entries*9], append(size, val...)...)

	// must increment the entry count and pos
	q.entries++
	q.pos += 8 + len(val)
}

func (q *BytesQueue) Pop() []byte {
	if q.entries == 0 {
		return nil
	}

	size := binary.LittleEndian.Uint64(q.buf[:8])
	val := q.buf[8 : 8+size]
	q.buf = append(q.buf[:0], q.buf[9+size:]...)

	// must decrement the entry count
	q.entries--
	return val
}

func (q *BytesQueue) Insert(val []byte, pos int) {
	q.checkGrow()
	size := make([]byte, 8)
	binary.LittleEndian.PutUint64(size, uint64(len(val)))
	q.buf = append(q.buf[:pos], append(size, val...)...)

	// must increment the entry count and pos
	q.entries++
	q.pos += 8 + len(val)
}

func (q *BytesQueue) InsertAt(val []byte, pos int) {
	q.Insert(val, pos)
}

func (q *BytesQueue) Remove(pos int, size int) {
	q.buf = append(q.buf[:pos], q.buf[pos+size:]...)

	// must decrement the entry count
	q.entries--
}

func (q *BytesQueue) RemoveAt(pos int, size int) {
	q.Remove(pos, size)
}

func (q *BytesQueue) Range(f func(pos int, p []byte) bool) {
	if q.entries == 0 {
		return
	}
	for pos := 0; pos < len(q.buf); pos = q.next(pos) {
		p, err := q.decode(pos)
		if err != nil && err == io.EOF {
			break
		}
		if !f(pos, p) {
			break
		}
	}
}

func (q *BytesQueue) checkGrow() {
	// initialize slice for the first time
	if q.buf == nil {
		log.Println("[BytesQueue] checkGrow: initializing slice")
		q.buf = make([]byte, defaultSize)
		q.entries = 0
		q.pos = 0
	}
	// if q.entries*9 >= len(q.buf) {
	// 	q.grow()
	// }
	if q.pos > 0 && q.pos+(q.pos/4) >= len(q.buf) {
		log.Println("[BytesQueue] checkGrow: calling grow")
		q.grow()
	}
}

func (q *BytesQueue) grow() {
	newData := make([]byte, cap(q.buf)*2)
	copy(newData, q.buf)
	q.buf = newData
}

func (q *BytesQueue) write(p []byte) error {
	q.checkGrow()
	if q.size+len(p)+8 >= cap(q.buf) {
		return io.EOF
	}
	n := q.size
	binary.LittleEndian.PutUint64(q.buf[n:n+8], uint64(len(p)))
	copy(q.buf[n+8:n+8+len(p)], p)
	q.size += 8 + len(p)
	q.pos = q.size + 8 + len(p)
	return nil
}

func (q *BytesQueue) encode(p []byte, at int) error {
	q.checkGrow()
	// early bounds check
	// _ = q.buf[pos]
	if q.pos >= cap(q.buf) {
		return io.EOF
	}
	// encode p's header into the buffer
	binary.LittleEndian.PutUint64(q.buf[at:at+8], uint64(len(p)))
	// copy p's contents into our bytes queue at the position
	copy(q.buf[at+8:at+8+len(p)], p)
	q.size += 8 + len(p)
	q.pos = at + 8 + len(p)
	return nil
}

func (q *BytesQueue) decode(pos int) ([]byte, error) {
	// early bounds check
	if pos >= len(q.buf) {
		return nil, io.EOF
	}

	// decode pos of the record
	size := binary.LittleEndian.Uint64(q.buf[pos : pos+8])
	return q.buf[pos+8 : pos+8+int(size)], nil
	// copy the buffer contents into the provided slice, p
	// copy(p, q.buf[pos+8:pos+8+int(pos)])
	// return nil
}

// next returns the position of the next entry
func (q *BytesQueue) next(pos int) int {
	// decode pos of the record
	n := int(binary.LittleEndian.Uint64(q.buf[pos : pos+8]))
	if n == 0 {
		return -1
	}
	return n + 8
}

// growSlice grows b by n, preserving the original content of b.
// If the allocation fails, it panics with ErrTooLarge.
func growSlice(b []byte, n int) []byte {
	defer func() {
		if recover() != nil {
			panic("bytes queue: too large")
		}
	}()
	// TODO(http://golang.org/issue/51462): We should rely on the append-make
	// pattern so that the compiler can call runtime.growslice. For example:
	//	return append(b, make([]byte, n)...)
	// This avoids unnecessary zero-ing of the first len(b) bytes of the
	// allocated slice, but this pattern causes b to escape onto the heap.
	//
	// Instead use the append-make pattern with a nil slice to ensure that
	// we allocate buffers rounded up to the closest pos class.
	c := len(b) + n // ensure enough space for n elements
	if c < 2*cap(b) {
		// The growth rate has historically always been 2x. In the future,
		// we could rely purely on append to determine the growth rate.
		c = 2 * cap(b)
	}
	b2 := append([]byte(nil), make([]byte, c)...)
	copy(b2, b)
	return b2[:len(b)]
}
