package buffer

import (
	"github.com/cagnosolutions/go-data/pkg/engine/page"
)

type BufferPoolManager struct {
	*BufferCacheManager
}

func NewBufferPoolManager(path string, pagesize uint16) (*BufferPoolManager, error) {
	bc, err := OpenBufferCacheManager(path, pagesize)
	if err != nil {
		return nil, err
	}
	return &BufferPoolManager{
		BufferCacheManager: bc,
	}, nil
}

// Alloc creates and returns a new page instance
func (bp *BufferPoolManager) Alloc() (page.Page, error) {
	return bp.NewPage(), nil
}

// Fetch fetches an already existing page by the page ID
func (bp *BufferPoolManager) Fetch(pid page.PageID) (page.Page, error) {
	return bp.FetchPage(pid), nil
}

// Flush flushes a page to disk using the page ID
func (bp *BufferPoolManager) Flush(pid page.PageID) error {
	return bp.FlushPage(pid)
}

func (bp *BufferPoolManager) Close() error {
	return bp.Close()
}
