package puredb

import (
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrTableNotFound  = errors.New("table not found")
)

// Record is an interface representing a single
// record able to be used by a collection.
type Record interface {

	// GetID should return the id of the record
	GetID() int

	// SetID should set the id of the record by
	// using the id provided
	SetID(id int)
}

// RecordSet is an interface representing a set of
// records able to be used by a collection.
type RecordSet interface {

	// Len should return the number of records
	// in the record set.
	Len() int
}

// CollectionInfo is an interface representing a collection
// and is responsible for returning information about records
// located within the collection.
type CollectionInfo interface {

	// GetRecords should return the total record count
	// along with the record id's that are currently in
	// this collection
	GetRecords() ([]int, int)

	// GetLastID should return the id of the record that
	// was last inserted
	GetLastID() int
}

// Collection is an interface representing a
// collection of records. It is synonymous with
// the notion of a table in a structured database.
type Collection interface {

	// insertRecord inserts a single record and returns the
	// id of the record that was just inserted.
	insertRecord(rec Record) (int, error)

	// insertRecords inserts multiple records and returns a
	// set of id's for the records that we just inserted.
	insertRecords(recs []Record) ([]int, error)

	// returnRecord returns a single record using the id provided.
	// The returned record is loaded into the pointer provided.
	returnRecord(id int, ptr Record) error

	// returnRecords returns all the records found in the
	// collection. The returned records are loaded into the set
	// of pointers provided.
	returnRecords(ptrs []Record) (int, error)

	// updateRecord updates the selected record using the id
	// and record provided.
	updateRecord(id int, rec Record) error

	// deleteRecord removes the selected record using
	// the id provided.
	deleteRecord(id int) error

	// search is a general purpose query method used to find a
	// set of records based on the query criteria. All records
	// that match the provided query should be returned into the
	// set of record pointers provided. It returns the number of
	// records that matched the query and a nil error on success.
	search(query string, ptrs []Record) (int, error)

	// getInfo returns an instance of CollectionInfo. A new
	// call to getInfo is required to get the most recent
	// changes if any mutations have been made to the collection.
	getInfo() CollectionInfo

	// drop is used to remove a collection and its data.
	drop() error

	// close is used to close (any optionally synchronize
	// before closing) the collection.
	close() error
}

// DBInfo is an interface representing a database and is responsible
// for returning information about the database, collections and
// records.
type DBInfo interface {

	// GetCollections returns a map where the map key is the
	// collection name, and the map value is an instance of
	// CollectionInfo for the named collection.
	GetCollections() map[string]CollectionInfo
}

// DB is an interface representing an embedded database
// engine that is able to perform all the basic operations
// necessary on collections and records.
type DB interface {

	// MakeCollection is used to instantiate a new collection.
	// It should only be used when you with to make a collection,
	// but will not be using it straight away.
	MakeCollection(collection string) error

	// DropCollection removes the named collection and all the
	// data it contains permanently.
	DropCollection(collection string) error

	// Insert inserts a single record into the named collection
	// using the record id provided. The returned record is loaded
	// into the record pointer provided.
	Insert(collection string, rec Record) (int, error)

	// InsertMany inserts multiple records and returns a set of id's
	// for the records that we just inserted.
	InsertMany(collection string, recs []Record) ([]int, error)

	// Return returns a single record using the id provided. The
	// returned record is loaded into the pointer provided.
	Return(collection string, id int, ptr Record) error

	// ReturnAll returns all the records found in the named
	// collection. The returned records are loaded into the
	// set of pointers provided. It will return the number of
	// records returned and a nil error on success.
	ReturnAll(collection string, ptrs []Record) (int, error)

	// Update updates the selected record in the named collection
	// using the id and record provided.
	Update(collection string, id int, rec Record) error

	// Delete removes a record from a collection using the
	// collection name and record id provided.
	Delete(collection string, id int) error

	// Search is a general purpose query method used to find
	// and return a set of records within a certain collection
	// based on the query criteria. All records in the collection
	// that match the provided query should be returned into the
	// set of record pointers provided. It returns the number of
	// records that matched the query and a nil error on success.
	Search(collection string, query string, ptrs []Record) (int, error)

	// GetInfo returns an instance of DBInfo which provides information
	// about the database. Any mutations made on the database invalidates
	// the last call to GetInfo. A new call to GetInfo is required to get
	// the most recent information.
	GetInfo() DBInfo

	// GetCollectionInfo returns an instance of CollectionInfo which
	// provides information for the selected collection. Any mutations
	// made to a collection invalidates the last call to GetInfo. A new
	// call to GetInfo is required to get the most recent information.
	GetCollectionInfo(collection string) CollectionInfo

	// Close synchronizes and closes the database and all collections.
	Close() error
}
