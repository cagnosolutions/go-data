package dbms

import (
	"errors"
)

// error values
var (
	ErrDataStoreExists        = errors.New("data store exists")
	ErrDataStoreDoesNotExist  = errors.New("data store does not exist")
	ErrDataStoreClosed        = errors.New("data store is closed")
	ErrNamespaceExists        = errors.New("data store namespace exists")
	ErrNamespaceDoesNotExists = errors.New("data store namespace does not exists")
)

// StorageEngine is a high level data storage and management engine type that can
// be used to access the underlying types (those being a page cache, indexes and
// a disk manager) through the API the StorageEngine provides.
type StorageEngine struct {
	store      string
	namespaces []string
}

// Create creates a data store using the provided store name. If the store already
// exists an error will be returned. Otherwise, a new StorageEngine will be returned.
func Create(store string) (*StorageEngine, error) {
	return nil, nil
}

// Open opens a data store using the provided store name. If the store does not exist,
// an error will be returned. Otherwise, a new StorageEngine will be returned.
func Open(store string) (*StorageEngine, error) {
	return nil, nil
}

// Info will attempt to return a list of information about the store. If nothing can
// be found, it will simply return nil
func (se *StorageEngine) Info() []string {
	return nil
}

// Close closes the data store. If the store is already closes, and error is returned.
func (se *StorageEngine) Close() error {
	return nil
}

// CreateNamespace creates a new namespace within this store using the provided name.
// If the namespace already exists an error will be returned.
func (se *StorageEngine) CreateNamespace(name string) error {
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
	return 0, nil
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
