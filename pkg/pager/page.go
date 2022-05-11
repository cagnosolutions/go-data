package pager

type page struct {
	pid pageID
}

func newPage(pid pageID) *page {
	return &page{
		pid: pid,
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
