package pager

const (
	statusUsed = 0xffff
	statusFree = 0x0000
)

type header struct {
	pid     pageID
	prevPid pageID
	nextPid pageID
	kind    uint16 // type of page
	status  uint16 // status of page
	free    uint16 // total free bytes in page
	slots   uint16 // number of slots in page
	freeLo  uint16 // free space lower offset
	freeHi  uint16 // free space upper offset
}

type page struct {
	header
	data []byte
}

func newPage(pid pageID) *page {
	return &page{
		header: header{
			pid:     pid,
			prevPid: 0,
			nextPid: 0,
			kind:    0,
			status:  statusUsed,
			free:    pageSize - hdrSize,
			slots:   0,
			freeLo:  hdrSize,
			freeHi:  pageSize,
		},
		data: make([]byte, pageSize),
	}
}

func (p *page) init() {

}

func (p *page) getPageID() pageID {
	return p.pid
}

func (p *page) aaa() {

}

func (p *page) bbb() {

}

func (p *page) ccc() {

}

func (p *page) ddd() {

}

func (p *page) eee() {

}

func (p *page) fff() {

}
