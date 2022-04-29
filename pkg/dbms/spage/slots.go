package spage

import (
	"bytes"
	"sort"
)

// interesting.

// pageSlot is a single index
// entry for a record (held in
// the *Page as a []*pageSlot)
type pageSlot struct {
	itemID     uint16
	itemStatus uint16
	itemOffset uint16
	itemLength uint16
}

// itemBounds returns the beginning and ending offset
// positions for the location of this item within the Page
func (s *pageSlot) itemBounds() (uint16, uint16) {
	return s.itemOffset, s.itemOffset + s.itemLength
}

// Len is here to satisfy the sort interface for
// sorting the Page slots by the record prefix
func (p *Page) Len() int {
	return len(p.slots)
}

// Swap is here to satisfy the sort interface for
// sorting the Page slots by the record prefix
func (p *Page) Swap(i, j int) {
	p.slots[i], p.slots[j] = p.slots[j], p.slots[i]
}

// Less is here to satisfy the sort interface for
// sorting the Page slots by the record prefix
func (p *Page) Less(i, j int) bool {
	ipre, _ := p.slots[i].itemBounds()
	jpre, _ := p.slots[j].itemBounds()
	return bytes.Compare(p.data[ipre:ipre+8], p.data[jpre:jpre+8]) < 0
}

// recordPrefixBySlot returns the record prefix for
// the given slot index (mainly here for sorting)
// **the Less call has been refactored since writing
// this method, so it might no longer be needed
func (p *Page) recordPrefixBySlot(n int) []byte {
	beg, _ := p.slots[n].itemBounds()
	return p.data[beg : beg+8]
}

// sortSlotsByRecordPrefix is a wrapper for sorting
// the Page slots by the record prefix
func (p *Page) sortSlotsByRecordPrefix() {
	sort.Stable(p)
}
