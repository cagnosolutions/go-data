package page

import (
	"fmt"
)

const (
	u64mask0to2 = 0x000000000000ffff
	u64mask2to4 = 0x00000000ffff0000
	u64mask4to6 = 0x0000ffff00000000
	u64mask6to8 = 0xffff000000000000
	shift2B     = 16
	shift4B     = 32
	shift6B     = 48
)

type cell uint64

func newCell(id, offset, length uint16) cell {
	var c cell
	c |= cell(id)
	c |= cell(C_USED) << shift2B
	c |= cell(offset) << shift4B
	c |= cell(length) << shift6B
	return c
}

func (c cell) getID() uint16 {
	return uint16(c)
}

func (c cell) getFlags() uint16 {
	return uint16(c >> shift2B)
}

func (c cell) hasFlag(flag uint16) bool {
	return uint16(c>>shift2B)&flag != 0
}

func (c cell) getOffset() uint16 {
	return uint16(c >> shift4B)
}

func (c cell) getLength() uint16 {
	return uint16(c >> shift6B)
}

func (c cell) getBounds() (uint16, uint16) {
	offset := uint16(c >> shift4B)
	length := uint16(c >> shift6B)
	return offset, offset + length
}

func (c *cell) setID(n uint16) {
	*c &^= u64mask0to2
	*c |= cell(n)
}

func (c *cell) setFlags(n uint16) {
	*c &^= u64mask2to4
	*c |= cell(n) << shift2B
}

func (c *cell) setOffset(n uint16) {
	*c &^= u64mask4to6
	*c |= cell(n) << shift4B
}

func (c *cell) setLength(n uint16) {
	*c &^= u64mask6to8
	*c |= cell(n) << shift6B
}

func (c cell) String() string {
	return fmt.Sprintf(
		"id=%d, flags=0x%.4x, offset=%d, length=%d", c.getID(), c.getFlags(), c.getOffset(),
		c.getLength(),
	)
}

type cells []cell

func (c cells) Len() int      { return len(c) }
func (c cells) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
