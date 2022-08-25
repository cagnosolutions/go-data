package engine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// error values
var (
	ErrDataStoreExists        = errors.New("data store exists")
	ErrDataStoreDoesNotExist  = errors.New("data store does not exist")
	ErrDataStoreClosed        = errors.New("data store is closed")
	ErrNamespaceExists        = errors.New("data store namespace exists")
	ErrNamespaceDoesNotExists = errors.New("data store namespace does not exists")
)

type store struct {
	cache  *PageCache
	recent *recID
}

// StorageEngine is a high level data storage and management engine type that can
// be used to access the underlying types (those being a page cache, indexes and
// a io manager) through the API the StorageEngine provides.
type StorageEngine struct {
	base        string
	journalPath string
	dataPath    string
	mu          sync.Mutex
	stores      map[string]store
	recents     map[string]*RecID
	journal     *WAL
}

// Open opens a data store using the provided store name. If the store does not exist,
// it will create a data store at the base path provided.
func Open(path string) (*StorageEngine, error) {
	// Create a new base path
	path = filepath.ToSlash(strings.Replace(path, filepath.Ext(path), "", -1))
	err := os.MkdirAll(path, 0644|os.ModeDir)
	if err != nil {
		return nil, err
	}
	// Setup other paths
	jp := filepath.Join(path, "journal")
	dp := filepath.Join(path, "data")
	// Enable the journal
	journal, err := OpenWAL(&WALConfig{
		BasePath:    jp,
		MaxFileSize: 256 << 10,
		SyncOnWrite: true,
	})
	if err != nil {
		return nil, err
	}
	// Set up a StorageEngine instance
	se := &StorageEngine{
		base:        path,
		journalPath: jp,
		dataPath:    dp,
		stores:      make(map[string]store),
		journal:     journal,
	}
	// Return storage engine
	return se, nil
}

// Info will attempt to return a list of information about the store. If nothing can
// be found, it will simply return nil
func (se *StorageEngine) Info() []string {
	se.mu.Lock()
	defer se.mu.Unlock()
	var ss []string
	for _, st := range se.stores {
		ss = append(ss, st.cache.JSON())
	}
	return ss
}

// Close closes the data store. If the store is already closes, and error is returned.
func (se *StorageEngine) Close() error {
	se.mu.Lock()
	defer se.mu.Unlock()
	var err error
	for _, st := range se.stores {
		err = st.cache.Close()
		if err != nil {
			return err
		}
	}
	err = se.journal.Close()
	if err != nil {
		return err
	}
	return nil
}

// CreateNamespace creates a new namespace within this store using the provided name.
// If the namespace already exists an error will be returned.
func (se *StorageEngine) CreateNamespace(name string) error {
	se.mu.Lock()
	defer se.mu.Unlock()
	_, found := se.stores[name]
	if found {
		return ErrDataStoreExists
	}
	pc, err := OpenPageCache(filepath.Join(se.dataPath), 64)
	if err != nil {
		return err
	}
	se.stores[name] = store{
		cache:  pc,
		recent: new(recID),
	}
	return nil
}

// NamespaceInfo will attempt to return a list of the namespaces and information this
// store currently has about the namespace provided using name. If no matching namespaces,
// can be found, it simply returns nil.
func (se *StorageEngine) NamespaceInfo(name string) []string {
	return nil
}

// DropNamespace will remove a namespace within this store using the provided name.
// If the namespace does not exist an error will be returned.
func (se *StorageEngine) DropNamespace(name string) error {
	return nil
}

// Insert will attempt to insert the provided data p into the provided namespace ns.
// It will return a record ID, or any potential errors encountered.
func (se *StorageEngine) Insert(ns string, p []byte) (uint32, error) {
	// Lock
	se.mu.Lock()
	defer se.mu.Unlock()
	// Try and locate namespace
	st, found := se.stores[ns]
	if !found {
		// Store not found
		return 0, ErrDataStoreDoesNotExist
	}
	// Got it; do our thing
	pid := st.recent.pid
	pg := st.cache.FetchPage(pid)
	if pg == nil {
		// Page not found
		return 0, ErrPageNotFound
	}
	// Add the record to the page
	rid, err := pg.addRecord(p)
	if err != nil {
		return 0, err
	}
	// Update the most recent record ID
	st.recent = rid
	// Unpin the page
	err = st.cache.UnpinPage(pid, true)
	if err != nil {
		return 0, err
	}
	// Finish and return
	return uint32(rid.sid), nil
}

// Return will attempt to return the record data matching the provided namespace ns,
// and record ID, id. If nothing can be found, nil data is returned along with any
// potential errors.
func (se *StorageEngine) Return(ns string, rid uint32) ([]byte, error) {
	return nil, nil
}

// ReturnAll will attempt to return all the record data located within the provided
// namespace ns, beginning at the record ID provided. If a positive limit is provided,
// it will limit the records returned to the number provided in limit. If limit is -1
// there will be no limit on the number of records that will be returned.
func (se *StorageEngine) ReturnAll(ns string, start uint32, limit int64) ([][]byte, error) {
	return nil, nil
}

// Update will attempt to update the record data matching the provided namespace ns.
// and record ID, rid, using the provided data p. If the record cannot be located or
// if the provided data is nil or any other error occurs, the error will be returned.
// On success, the error will be nil, and the boolean will indicate true.
func (se *StorageEngine) Update(ns string, rid uint32, p []byte) (bool, error) {
	return false, nil
}

// Delete will attempt to remove any record data matching the provided namespace ns,
// and record ID, rid. If nothing can be found, an error will be returned.
func (se *StorageEngine) Delete(ns string, rid uint32) error {
	return nil
}
