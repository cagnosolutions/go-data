package engine

import (
	"github.com/cagnosolutions/go-data/pkg/engine/buffer"
	"github.com/cagnosolutions/go-data/pkg/engine/page"
	"github.com/cagnosolutions/go-data/pkg/engine/storage"
)

type Config struct {
	BasePath   string
	PageFrames uint16
}

var defaultConfig = &Config{
	BasePath:   "db",
	PageFrames: 64,
}

type StorageEngine struct {
	conf  *Config
	store *storage.DiskStore
	pool  *buffer.BufferPoolManager
}

func OpenStorageEngine(conf *Config) (*StorageEngine, error) {
	if conf == nil {
		conf = defaultConfig
	}
	store, err := storage.Open(conf.BasePath)
	if err != nil {
		return nil, err
	}
	pool, err := buffer.New(store, conf.PageFrames)
	if err != nil {
		return nil, err
	}
	s := &StorageEngine{
		conf:  conf,
		store: store,
		pool:  pool,
	}
	return s, nil
}

// Allocate creates and returns a new page instance
func (s *StorageEngine) Allocate() (page.Page, error) {
	return s.pool.NewPage()
}

// Fetch fetches an already existing page by the page ID
func (s *StorageEngine) Fetch(pid page.PageID) (page.Page, error) {
	return s.pool.FetchPage(pid)
}

// Flush flushes a page to disk using the page ID
func (s *StorageEngine) Flush(pid page.PageID) error {
	return s.pool.FlushPage(pid)
}

func (s *StorageEngine) Close() error {
	var err error
	err = s.pool.Close()
	if err != nil {
		return err
	}
	return err
}
