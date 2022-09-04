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
	*c = cell(n) << 0
}

func (c *cell) setFlags(n uint16) {
	*c = cell(n) << 16
}

func (c *cell) setOffset(n uint16) {
	*c = cell(n) << 32
}

func (c *cell) setLength(n uint16) {
	*c = cell(n) << 48
}

func (c cell) String() string {
	return fmt.Sprintf(
		"id=%d, flags=0x%.4x, offset=%d, length=%d", c.getID(), c.getFlags(), c.getOffset(),
		c.getLength(),
	)
}
