package puredb

import (
	"log"
	"os"
	"path/filepath"
	"sync"
)

type Info struct {
	tables map[string]CollectionInfo
}

func (i *Info) GetCollections() map[string]CollectionInfo {
	return i.tables
}

// PureDB is a simple embedded database engine that
// stores records encoded in NDJSON or JSONL format.
type PureDB struct {
	base   string
	tables map[string]*table
	lock   sync.RWMutex
}

// Open opens a database.
func Open(path string) (*PureDB, error) {
	path = filepath.ToSlash(filepath.Dir(path))
	err := os.MkdirAll(path, 0666)
	if err != nil {
		return nil, err
	}
	db := &PureDB{
		base:   path,
		tables: make(map[string]*table),
	}
	// load fills out the table index for
	// the database engine.
	err = db.load()
	if err != nil {
		return nil, err
	}
	return db, nil
}

// load is an initialization method for the database.
// It initializes the base path if it is not there
// and fills out the table index.
func (db *PureDB) load() error {
	// lock for reading
	db.lock.RLock()
	defer db.lock.RUnlock()
	// read the base directory
	files, err := os.ReadDir(db.base)
	if err != nil {
		return err
	}
	// load tables
	for _, file := range files {
		// check the file
		if !fileIsTable(file) {
			// skip non table files
			continue
		}
		// otherwise, we get the /name
		name := file.Name()
		// open table
		t, err := openTable(db.base, name)
		if err != nil {
			return err
		}
		// add to open tables
		db.tables[name] = t
	}
	// finished
	return nil
}

// MakeCollection takes a table name and creates and initializes a
// new table. It returns an error if the table is already created.
// CreateTable should only be used in rare cases if you wish to
// initialize a table, but you are not going to use it right away.
func (db *PureDB) MakeCollection(name string) error {
	// lock for reading and writing
	db.lock.Lock()
	defer db.lock.Unlock()
	// open the table
	t, err := openTable(db.base, name)
	if err != nil {
		return err
	}
	// don't forget to close!
	err = t.close()
	if err != nil {
		return err
	}
	return nil
}

// DropCollection takes a table name and completely deletes all data
// within the table and removes the table file itself. It returns
// an error if the table is not found, or if there was trouble
// deleting it.
func (db *PureDB) DropCollection(name string) error {
	// lock for read and write
	db.lock.Lock()
	defer db.lock.Unlock()
	// get a sanitized path first
	path := filepath.Clean(filepath.ToSlash(filepath.Join(db.base, name)))
	// check to see if the table exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// table is not there, do nothing
		return nil
	}
	// get table from tables map
	t, ok := db.tables[name]
	if !ok {
		// table not found
		return ErrTableNotFound
	}
	// otherwise, it's there, so remove it
	err = t.drop()
	if err != nil {
		return err
	}
	// remove it from the tables map
	delete(db.tables, name)

	return nil
}

// Insert takes a table name and a struct that implements the Record
// interface and adds a new record to the table. It will add duplicate
// records. It returns the ID of the record that was just added.
func (db *PureDB) Insert(name string, rec Record) (int, error) {
	// lock for reading and writing
	db.lock.Lock()
	defer db.lock.Unlock()
	// check to see if table is in tables map
	t, ok := db.tables[name]
	if !ok {
		// open table with name provided
		t, err := openTable(db.base, name)
		if err != nil {
			return -1, err
		}
		// add table to open tables
		db.tables[name] = t
	}
	// make sure we are using the right table
	t = db.tables[name]
	// call the insert method on the table
	id, err := t.insertRecord(rec)
	if err != nil {
		return -1, err
	}
	// return id of record inserted
	return id, nil
}

func (db *PureDB) InsertMany(name string, recs []Record) ([]int, error) {
	return nil, nil
}

// Return takes a table name and a pointer to a struct that implements
// the Record interface. The found record is unmarshalled into the
// provided pointer. If the provided id cannot be found or does not
// match a record in the table an ErrRecordNotFound error is returned.
func (db *PureDB) Return(name string, id int, ptr Record) error {
	// lock for reading and writing
	db.lock.Lock()
	defer db.lock.Unlock()
	// check to see if table is in tables map
	t, ok := db.tables[name]
	if !ok {
		// open table with name provided
		t, err := openTable(db.base, name)
		if err != nil {
			return err
		}
		// add table to open tables
		db.tables[name] = t
	}
	// make sure we are using the right table
	t = db.tables[name]
	// call the return method on the table
	err := t.returnRecord(id, ptr)
	if err != nil {
		return err
	}
	return nil
}

// ReturnAll returns all the records found in the named collection. The
// returned records are loaded into the set of pointers provided. It will
// return the number of records returned and a nil error on success.
func (db *PureDB) ReturnAll(name string, ptrs interface{}) (int, error) {
	// lock for reading and writing
	db.lock.Lock()
	defer db.lock.Unlock()
	// check to see if table is in tables map
	t, ok := db.tables[name]
	if !ok {
		// open table with name provided
		t, err := openTable(db.base, name)
		if err != nil {
			return -1, err
		}
		// add table to open tables
		db.tables[name] = t
	}
	// make sure we are using the right table
	t = db.tables[name]
	// call the returnAll method on the table
	n, err := t.returnRecords(ptrs)
	if err != nil {
		return -1, err
	}
	return n, err
}

// Update takes a table name and a struct implementing the Record
// interface. It updates the record in the table that has that matches
// the provided Record's ID.
func (db *PureDB) Update(name string, id int, rec Record) error {
	// lock for reading and writing
	db.lock.Lock()
	defer db.lock.Unlock()
	// check to see if table is in tables map
	t, ok := db.tables[name]
	if !ok {
		// open table with name provided
		t, err := openTable(db.base, name)
		if err != nil {
			return err
		}
		// add table to open tables
		db.tables[name] = t
	}
	// make sure we are using the right table
	t = db.tables[name]
	// call the return method on the table
	err := t.updateRecord(id, rec)
	if err != nil {
		return err
	}
	return nil
}

// Delete takes a table name and a record id and removes the record
// that has a matching id.
func (db *PureDB) Delete(name string, id int) error {
	// lock for reading and writing
	db.lock.Lock()
	defer db.lock.Unlock()
	// check to see if table is in tables map
	t, ok := db.tables[name]
	if !ok {
		// open table with name provided
		t, err := openTable(db.base, name)
		if err != nil {
			return err
		}
		// add table to open tables
		db.tables[name] = t
	}
	// make sure we are using the right table
	t = db.tables[name]
	// call the return method on the table
	err := t.deleteRecord(id)
	if err != nil {
		return err
	}
	return nil
}

// Search takes a table name, a record id, and a pointer to a struct that
// implements the Record interface, finds the associated record from the
// table, and populates the struct.
func (db *PureDB) Search(name string, query string, ptrs interface{}) (int, error) {
	// lock for reading and writing
	db.lock.Lock()
	defer db.lock.Unlock()
	// check to see if table is in tables map
	t, ok := db.tables[name]
	if !ok {
		// open table with name provided
		t, err := openTable(db.base, name)
		if err != nil {
			return -1, err
		}
		// add table to open tables
		db.tables[name] = t
	}
	// make sure we are using the right table
	t = db.tables[name]
	// call the search method of the table
	n, err := t.search(query, ptrs)
	if err != nil {
		return -1, err
	}
	// done
	return n, nil
}

func (db *PureDB) GetInfo() DBInfo {
	// lock for reading and writing
	db.lock.Lock()
	defer db.lock.Unlock()
	// return
	return nil
}

func (db *PureDB) GetCollectionInfo(name string) CollectionInfo {
	// lock for reading and writing
	db.lock.Lock()
	defer db.lock.Unlock()
	// load or get table
	t, err := db.getTable(name)
	if err != nil {
		return nil
	}
	// and return table info
	return t.getInfo()
}

func (db *PureDB) Close() error {
	// lock for reading and writing
	db.lock.Lock()
	defer db.lock.Unlock()
	// make sure all the tables that are open
	// are closed before closing the db
	for _, t := range db.tables {
		// close table
		err := t.close()
		if err != nil {
			log.Panic(err)
		}
	}
	// now we can return
	return nil
}

// getTable is a helper method that returns a table
// from the table index.
func (db *PureDB) getTable(name string) (*table, error) {
	// check to see if table is in tables map
	_, ok := db.tables[name]
	if !ok {
		// open table with name provided
		t, err := openTable(db.base, name)
		if err != nil {
			return t, err
		}
		// add table to table index
		db.tables[name] = t
	}
	// otherwise, we have it in the table index
	// so, we simply just need to return it
	return db.tables[name], nil
}

func fileIsTable(de os.DirEntry) bool {
	if !de.IsDir() {
		switch filepath.Ext(de.Name()) {
		case ".json", ".jsonl", ".ndjson":
			return true
		}
	}
	return false
}

func (db *PureDB) IsOpen() bool {
	return db.base != ""
}
