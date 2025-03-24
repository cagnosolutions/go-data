package main

const (
	// constants for the page, headers, cellptrs and record sizes
	pageSize         = 16 << 10
	pageHeaderSize   = 24
	pageCellPtrSize  = 8
	recordHeaderSize = 6
)

type RecordID struct {
	PageID uint32
	CellID uint16
}

type CellPtr struct {
	ID     uint16
	Offset uint16
	Length uint16
}

type PageHeader struct {
	ID    uint32 // page id
	Prev  uint32 // previous page id
	Next  uint32 // next page id
	Flags uint32 // flags and meta data for the page
	Cells uint16 // number of cells allocated in this page
	Free  uint16 // number of cells that are free (empty) to use
	Lower uint16 // lower bound of numFree space
	Upper uint16 // upper bound of numFree space
}

type Page struct {
	PageHeader
	Cells []CellPtr
}

func NewPage(id uint32, flags uint32) *Page {
	// allocate a new page
	p := new(Page)

	// set up the page header
	p.PageHeader = PageHeader{
		ID:    id,
		Prev:  0,
		Next:  0,
		Flags: flags,
		Cells: 0,
		Free:  0,
		Lower: pageHeaderSize,
		Upper: pageSize,
	}

	// initialize the page cells
	p.Cells = make([]CellPtr, 0)

	// return the new page
	return p
}

// Record is a binary data type
type Record []byte

// AddRecord adds a new Record to the page, and returns a *RecordID, or an error
func (p *Page) AddRecord(r Record) (*RecordID, error) {
	return nil, nil
}

// GetRecord takes a *RecordID and attempts to returnthe associated Record. If it
// cannot be located, then an error is returned.
func (p *Page) GetRecord(id *RecordID) (Record, error) {
	return nil, nil
}

// GetRecordByKey attempts to locate and return a Record using the provided
// record key. It performs a binary search, since the record cellptrs are
// always kept in a sorted order, and attempts to return a matching Record. If
// a there is more than one Record in the Page that has the same key then
// it will return the first one it locates. If it cannot locate a matching
// Record then a nil Record will be returned.
func (p *Page) GetRecordByKey(key []byte) *Record {
	return nil
}

// DelRecord attempts to delete a record using the provided record ID. The
// associated cellptr will be marked as free to re-use, and the record data
// will be overwritten. Any errors will be returned.
func (p *Page) DelRecord(id *RecordID) error {
	return nil
}

// RangeRecords takes a range function and enumerates the records in the page.
func (p *Page) RangeRecords(fn func(r *Record) error) error {
	return nil
}
