package page

import (
	"fmt"
)

type cell uint64

func newCell(id, offset, length uint16) cell {
	var c cell
	c |= cell(id) << 0
	c |= cell(C_USED) << 16
	c |= cell(offset) << 32
	c |= cell(length) << 48
	return c
}

func (c cell) getID() uint16 {
	return uint16(c >> 0)
}

func (c cell) getFlags() uint16 {
	return uint16(c >> 16)
}

func (c cell) getOffset() uint16 {
	return uint16(c >> 32)
}

func (c cell) getLength() uint16 {
	return uint16(c >> 48)
}

func (c *cell) setID(n uint16) {
	*c &^= 0x000000000000ffff // mask bytes 0-2 and clear
	*c |= cell(n) << 0
}

func (c *cell) setFlags(n uint16) {
	*c &^= 0x00000000ffff0000 // mask bytes 2-4 and clear
	*c |= cell(n) << 16
}

func (c *cell) setOffset(n uint16) {
	*c &^= 0x0000ffff00000000 // mask bytes 4-6 and clear
	*c |= cell(n) << 32
}

func (c *cell) setLength(n uint16) {
	*c &^= 0xffff000000000000 // mask bytes 6-8 and clear
	*c |= cell(n) << 48
}

// (*bs)[i>>3] |= 1 << (i & (7))
//
// val = (val &^ mask) | (newval & mask)
//
// val &^= 0xfffffff0 // clear lower 4 bits
// val |= lower4bits & 0xfffffff0

func (c cell) String() string {
	return fmt.Sprintf(
		"id=%d, flags=0x%.4x, offset=%d, length=%d", c.getID(), c.getFlags(), c.getOffset(),
		c.getLength(),
	)
}
