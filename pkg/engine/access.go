package engine

import (
	"github.com/cagnosolutions/go-data/pkg/engine/buffer"
	"github.com/cagnosolutions/go-data/pkg/engine/page"
)

type Config struct {
	BasePath string
}

type StorageEngine struct {
	base string
	pool *buffer.BufferPoolManager
}

func OpenStorageEngine(base string) *StorageEngine {
	return new(StorageEngine)
}

// Allocate creates and returns a new page instance
func (s *StorageEngine) Allocate() (page.Page, error) {
	return s.pool.NewPage(), nil
}

// Fetch fetches an already existing page by the page ID
func (s *StorageEngine) Fetch(pid page.PageID) (page.Page, error) {
	return s.pool.FetchPage(pid), nil
}

// Flush flushes a page to disk using the page ID
func (s *StorageEngine) Flush(pid page.PageID) error {
	return s.pool.FlushPage(pid)
}

func (s *StorageEngine) Close() error {
	return s.pool.Close()
}
