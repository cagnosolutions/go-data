package puredb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"

	"github.com/cagnosolutions/go-data/pkg/json/ndjson"
)

type TableInfo struct {
	LastID      int   `json:"last_id"`
	Records     []int `json:"records"`
	RecordCount int   `json:"record_count"`
}

func (i *TableInfo) GetRecords() ([]int, int) {
	return i.Records, i.RecordCount
}

func (i *TableInfo) GetLastID() int {
	return i.LastID
}

type table struct {
	base        string        // base is the database root base
	name        string        // name is the name of the table
	fp          *os.File      // fp is the table file pointer
	offs        map[int]int64 // offs stores the record offsets
	recordCount int           // recordCount is the number of records
	lastID      int           // lastID is the last id added
}

func openTable(base, name string) (*table, error) {
	// open or create the file
	fp, err := openFile(base, name)
	if err != nil {
		return nil, err
	}
	// initialize our new table
	t := &table{
		base: base,
		name: name,
		fp:   fp,
		offs: make(map[int]int64),
	}
	// run our load method
	err = t.load()
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (t *table) load() error {
	// setup our record map
	var m map[string]interface{}
	// setup our line reader
	lr := ndjson.NewLineReader(t.fp)
	// get the current offset
	off, err := t.fp.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	// and start reading
	for {
		// read (raw) record
		dat, err := lr.ReadRaw()
		if err != nil {
			// check for end of file
			if err == io.EOF {
				break
			}
			return err
		}
		// unmarshal data into map
		err = json.Unmarshal(dat, &m)
		if err != nil {
			return err
		}
		// get id from data
		id, ok := m["_id"].(float64)
		if !ok {
			// log.Printf("type assertion failed! (off=%d, m=%v, id=%v)\recordCount", off, m, id)
			return err
		}
		// add to offset map
		t.offs[int(id)] = off
		// increment record count
		t.recordCount++
		// update lastID
		t.lastID = int(id)
		// update the offset (+1 is for the '\recordCount' delimiter)
		off += int64(len(dat) + 1)
	}
	return nil
}

func (t *table) insertRecord(rec Record) (int, error) {
	// go to the end of the file
	off, err := t.fp.Seek(0, io.SeekEnd)
	if err != nil {
		return -1, err
	}
	// open line writer
	lw := ndjson.NewLineWriter(t.fp)
	// set the record id
	rec.SetID(t.lastID + 1)
	// write record data
	_, err = lw.Write(rec)
	if err != nil {
		return -1, err
	}
	// after record is successfully written
	// increment record count, lastID and
	// add record to the offset map
	t.recordCount++
	t.lastID++
	t.offs[t.lastID] = off
	// return record id or record inserted
	return t.lastID, nil
}

func (t *table) insertRecords(recs []Record) ([]int, error) {
	// go to the end of the file
	off, err := t.fp.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	// open line writer
	lw := ndjson.NewLineWriter(t.fp)
	// init record counter
	var rc int
	// init record id set to return
	var rids []int
	// loop through the records, setting
	// record id's and writing all in one
	// shot
	for i := range recs {
		// set id
		rc++
		recs[i].SetID(t.lastID + rc)
		// write record data
		n, err := lw.Write(recs[i])
		if err != nil {
			return nil, err
		}
		// update offset map
		t.offs[t.lastID+rc] = off
		// increment offset using n
		off += int64(n)
		// add id to recoid id set
		rids = append(rids, t.lastID+rc)
	}
	// after all records have successfully been
	// written, increment record count, lastID and
	// add records to the offset map.
	t.recordCount += rc
	t.lastID += recs[len(recs)-1].GetID()
	// return record id set
	return rids, nil
}

func (t *table) returnRecord(id int, ptr Record) error {
	// find the record offset using the provided id
	off, ok := t.offs[id]
	if !ok {
		// not found
		return ErrRecordNotFound
	}
	// open new line reader
	lr := ndjson.NewLineReader(t.fp)
	// move the file pointer to the record offset
	_, err := t.fp.Seek(off, io.SeekStart)
	if err != nil {
		return err
	}
	// read record
	err = lr.Read(ptr)
	if err != nil {
		return err
	}
	return nil
}

func (t *table) returnRecords(ptrs interface{}) (int, error) {
	// go to the start of the file
	_, err := t.fp.Seek(0, io.SeekStart)
	if err != nil {
		return -1, err
	}
	// open our line reader
	lr := ndjson.NewLineReader(t.fp)
	// read all the records
	n, err := lr.ReadAll(ptrs)
	if err != nil {
		return -1, err
	}
	// success
	return n, nil
}

func (t *table) updateRecord(id int, rec Record) error {
	// find the record offset using the provided id
	off, ok := t.offs[id]
	if !ok {
		// not found
		return ErrRecordNotFound
	}
	// open new line reader
	lr := ndjson.NewLineReader(t.fp)
	// move the file pointer to the record offset
	_, err := t.fp.Seek(off, io.SeekStart)
	if err != nil {
		return err
	}
	// read the raw record
	dat, err := lr.ReadRaw()
	if err != nil {
		return err
	}
	// unmarshal the record
	var tmp map[string]interface{}
	err = json.Unmarshal(dat, &tmp)
	if err != nil {
		return err
	}
	// update the tmp with the changes
	for fld, val := range tmp {
		v := getField(rec, fld)
		if val != v && v != nil {
			val = v
		}
	}
	// create a tombstone that is len(dat) sized
	tomb := makeTombstone(len(dat))
	// seek back to the start of the record
	// that we just read and made a copy of
	_, err = t.fp.Seek(off, io.SeekStart)
	if err != nil {
		return err
	}
	// open a new line writer
	lw := ndjson.NewLineWriter(t.fp)
	// write tombstone in place
	_, err = lw.WriteRaw(tomb)
	if err != nil {
		return err
	}
	// go to the end of the file
	off, err = t.fp.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	// open line writer once more
	lw = ndjson.NewLineWriter(t.fp)
	// set the record id
	rec.SetID(id)
	// write record data
	_, err = lw.Write(rec)
	if err != nil {
		return err
	}
	return nil
}

func (t *table) deleteRecord(id int) error {
	// find the record offset using the provided id
	off, ok := t.offs[id]
	if !ok {
		// not found
		return ErrRecordNotFound
	}
	// open new line reader
	lr := ndjson.NewLineReader(t.fp)
	// move the file pointer to the record offset
	_, err := t.fp.Seek(off, io.SeekStart)
	if err != nil {
		return err
	}
	// read the raw record
	dat, err := lr.ReadRaw()
	if err != nil {
		return err
	}
	// create a tombstone that is len(dat) sized
	tomb := makeTombstone(len(dat))
	// seek back to the start of the record
	_, err = t.fp.Seek(off, io.SeekStart)
	if err != nil {
		return err
	}
	// open a new line writer
	lw := ndjson.NewLineWriter(t.fp)
	// write tombstone in place
	_, err = lw.WriteRaw(tomb)
	if err != nil {
		return err
	}
	return nil
}

func (t *table) search(query string, ptrs interface{}) (int, error) {
	// go to the start of the file
	_, err := t.fp.Seek(0, io.SeekStart)
	if err != nil {
		return -1, err
	}
	// open our line reader
	lr := ndjson.NewLineReader(t.fp)
	// call our read all matcher method
	return lr.ReadAllMatch(query, ptrs)
}

// search finds things; for now, just do some simple matching
func (t *table) _search(query string, ptrs interface{}) (int, error) {
	// go to the start of the file
	_, err := t.fp.Seek(0, io.SeekStart)
	if err != nil {
		return -1, err
	}
	// we'll need this later
	// var buf bytes.Buffer
	// split query
	qq := bytes.Split([]byte(query), []byte{';'})
	// open our line reader
	lr := ndjson.NewLineReader(t.fp)
	// also open our line writer
	// lw := ndjson.NewLineWriter(&buf)
	var matches [][]byte
	// and start reading
	for {
		// read (raw) record
		dat, err := lr.ReadRaw()
		if err != nil {
			// check for end of file
			if err == io.EOF {
				break
			}
			return -1, err
		}
		// range "query" parts and check for simple match
		for _, q := range qq {
			// check for a simple match
			ok := bytes.Contains(dat, q) && q != nil && len(q) > 0
			if ok {
				// if we have a match, write record to the buffer
				// log.Printf("rec: %q, match=%v (%+v)\n", dat, ok, q)
				// _, err = lw.WriteRaw(dat)
				// if err != nil {
				//	return -1, err
				// }
				matches = append(matches, dat)
			}
		}
		// and we continue to find matches...
	}
	// once we are done finding out matches, lets unmarshall the
	// data, but first we must open a new line reader on our buffer
	lr = ndjson.NewLineReader(bytes.NewReader(bytes.Join(matches, []byte{'\n'})))
	// lr = ndjson.NewLineReader(&buf)
	// and read and unmarshall all our matches records
	n, err := lr.ReadAll(ptrs)
	if err != nil {
		return -1, err
	}
	// success
	return n, nil
}

func (t *table) getInfo() CollectionInfo {
	// assemble record ID's
	var records []int
	for id, _ := range t.offs {
		records = append(records, id)
	}
	sort.Ints(records)
	// return table info
	return &TableInfo{
		LastID:      t.lastID,
		Records:     records,
		RecordCount: t.recordCount,
	}
}

func (t *table) drop() error {
	// close file pointer
	err := t.fp.Close()
	if err != nil {
		return err
	}
	// remove table
	err = removeFile(t.base, t.name)
	if err != nil {
		return err
	}
	return nil
}

func (t *table) close() error {
	// make sure everything is synced
	err := t.fp.Sync()
	if err != nil {
		return err
	}
	err = t.fp.Close()
	if err != nil {
		return err
	}
	return nil
}

func (t *table) _offsetMap() map[int]int64 {
	return t.offs
}

func (t *table) String() string {
	return fmt.Sprintf(`{"name":%q,"last_id":%d,"record_count":%d}`, t.name, t.lastID, t.recordCount)
}

func makeTombstone(size int) []byte {
	// create a tombstone that is len(dat) sized
	tomb := make([]byte, size)
	for i := range tomb {
		switch i {
		case 0:
			tomb[i] = '{'
		case 1:
			tomb[i] = '"'
		case 2:
			tomb[i] = 'T'
		case 3:
			tomb[i] = '"'
		case 4:
			tomb[i] = ':'
		case 5:
			tomb[i] = '"'
		case size - 2:
			tomb[i] = '"'
		case size - 1:
			tomb[i] = '}'
		default:
			tomb[i] = 'X'
		}
	}
	return tomb
}

func getField(v Record, field string) interface{} {
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	if !f.IsValid() {
		return nil
	}
	return f.Interface()
}
