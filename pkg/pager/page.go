package pager

// the following constants cannot be OR
const (
	statusFree = 0x0000
	statusUsed = 0x0001
	statusSpec = 0x0002
	typeA      = 0x0010
	typeB      = 0x0020
	typeC      = 0x0040
	typeD      = 0x0080
)

type header struct {
	pid   pageID // page id
	kind  uint16 // kind & status of page
	slots uint16 // number of slots in page
	lower uint16 // free space lower offset
	upper uint16 // free space upper offset
	data  []byte // raw page data
}

type page struct {
	header
	data []byte
}

func newPage(pid pageID) *page {
	return &page{
		header: header{
			pid:   pid,
			kind:  0,
			slots: 0,
			lower: 0,
			upper: 0,
		},
		data: make([]byte, pageSize),
	}
}

func (p *page) getPageID() pageID {
	return p.pid
}
