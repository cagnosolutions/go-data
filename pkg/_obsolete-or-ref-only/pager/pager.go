package pager

import (
	"errors"
)

var ()

type Pager struct {
	bp *bufferPool
}

// Open attempts to open and return, or create and return an instance of a Pager.
// It takes a file name or path ending with a file name, strips the suffix and
// writes to that file. It also creates a meta file with information in it that the
// pager can use.
func Open(path string, pageSize, pageCount uint16) (*Pager, error) {
	// First, check the page size to ensure it is acceptable.
	if pageSize&(szPg-1) != 0 {
		return nil, ErrBadPageSize
	}
	// Next, we want to sanitize the provided path, and trim the provided file suffix.
	full, err := pathCleanAndTrimSUffix(path)
	if err != nil {
		return nil, err
	}
	// Then, we can pass the path to the disk disk instantiation instance.
	// dm, err := newDiskManager(full, pageSize, pageCount)
	// if err != nil {
	//	return nil, err
	// }
	// Finally, we can instantiate and return a new *Pager instance.
	return &Pager{
		bp: newBufferPool(full, pageSize, pageCount),
	}, nil
}

// NewPage allocates and returns a new page of data.
func (p *Pager) NewPage() (Page, error) {
	pg := p.bp.newPage()
	if pg == nil {
		return nil, errors.New("page frames are all full, and non could be evicted")
	}
	return pg, nil
}

// FetchPage locates and returns the page matching the supplied page ID.
func (p *Pager) FetchPage(pid PageID) (Page, error) {
	pg := p.bp.fetchPage(pid)
	if pg == nil {
		return nil, errors.New(
			"got an error: either reading the page off of the disk or" +
				" the frames are all full, and non could be evicted",
		)
	}
	return pg, nil
}

// UnpinPage locates the page matching the supplied page ID and lets the
// buffer pool that we are finished using it. A page should be flushed
// before UnpinPage is call in most situations.
func (p *Pager) UnpinPage(pid PageID, isDirty bool) error {
	return p.bp.unpinPage(pid, isDirty)
}

// FlushPage flushes any changes made to the page back to the underlying
// storage disk.
func (p *Pager) FlushPage(pid PageID) error {
	return p.bp.flushPage(pid)
}

// DeletePage locates the page matching the supplied page ID and removes
// it from the buffer pool. It also removes the page from the underlying
// storage disk.
func (p *Pager) DeletePage(pid PageID) error {
	return p.bp.deletePage(pid)
}

// FlushAll is just like FlushPage except for it will flush all the pages
// that are currently resident in the buffer pool.
func (p *Pager) FlushAll() error {
	return p.bp.flushAll()
}

// Close closes the pager
func (p *Pager) Close() error {
	return p.bp.close()
}
