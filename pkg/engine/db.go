package engine

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/cagnosolutions/go-data/pkg/engine/buffer"
	"github.com/cagnosolutions/go-data/pkg/engine/page"
	"github.com/cagnosolutions/go-data/pkg/engine/storage"
)

const (
	defaultFrameCount = 64
	badID             = ^uint64(0) - 1
)

var (
	ErrNSNotFound = errors.New("namespace could not be found")
	ErrBadID      = errors.New("bad id")
)

type DB struct {
	base string
	data sync.Map
}

func OpenDB(base string) (*DB, error) {
	return &DB{
		base: filepath.ToSlash(base),
	}, nil
}

func (db *DB) getNS(base, name string) (Namespace, error) {
	v, exists := db.data.Load(nsPath(base, name))
	if !exists {
		return nil, ErrNSNotFound
	}
	ns, ok := v.(Namespace)
	if !ok {
		return nil, errors.New("something weird happened")
	}
	return ns, nil
}

func (db *DB) Select(name string) (Namespace, error) {
	// locate the namespace matching the provided name
	ns, err := db.getNS(db.base, name)
	if err != nil {
		return nil, err
	}
	return ns, nil
}

func (db *DB) Create(name string) (Namespace, error) {
	// check to see if this namespace already exists
	ns, err := db.getNS(db.base, name)
	if err == nil {
		// namespace was found, return it without an error
		return ns, nil
	}
	if !errors.Is(err, ErrNSNotFound) {
		// something weird happened, return an error
		return nil, err
	}
	// otherwise, namespace does not exist, so we will need
	// to create a new one.
	c, err := openCollection(db.base, name)
	if err != nil {
		return nil, err
	}
	// add to our map
	db.data.Store(c.path, c)
	return c, nil
}

func (db *DB) Drop(name string) error {
	// check to see if this namespace already exists
	ns, err := db.getNS(db.base, name)
	if err == nil {
		// namespace was found--time to destroy
		if err = ns.destroy(); err != nil {
			return err
		}
		// and remove from our mapping
		db.data.Delete(nsPath(db.base, name))
	}
	if !errors.Is(err, ErrNSNotFound) {
		// something weird happened, return an error
		return err
	}
	return nil
}

func (db *DB) Close() error {
	var err error
	fn := func(path, v any) bool {
		ns, ok := v.(Namespace)
		if !ok {
			return true
		}
		err = ns.close()
		if err != nil {
			return false
		}
		db.data.Delete(path)
		return true
	}
	db.data.Range(fn)
	return err
}

type Collection struct {
	path   string
	pool   *buffer.BufferPoolManager
	curr   uint32
	maxPID uint32
}

func nsPath(base, name string) string {
	return filepath.ToSlash(filepath.Join(base, name))
}

func openCollection(base, name string) (*Collection, error) {
	path := nsPath(base, name)
	fp, err := storage.Open(path)
	if err != nil {
		return nil, err
	}
	bp, err := buffer.New(fp, defaultFrameCount)
	if err != nil {
		return nil, err
	}
	return &Collection{
		path: path,
		pool: bp,
	}, nil
}

func encRecord(data Record) (page.Record, error) {
	id := make([]byte, 4)
	binary.LittleEndian.PutUint32(id, data.GetID())
	rec, err := data.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return page.NewRecord(page.R_NUM, page.R_STR, id, rec), nil
}

func decRecord(r page.Record, ptr Record) error {
	err := ptr.UnmarshalBinary(r.Val())
	if err != nil {
		return err
	}
	ptr.SetID(binary.LittleEndian.Uint32(r.Key()))
	return nil
}

func (c *Collection) Find(f func(rec Record) bool) ([]Record, error) {
	// fi, err := os.Stat(c.path)
	// if err != nil {
	// 	return nil, err
	// }
	// var pg page.Page
	// var err error
	// maxPID := uint32(fi.Size()) / page.PageSize
	//
	// fn := func(r *page.Record) error {
	//
	// 	recr.Val()
	// }
	//
	// for pid := uint32(0); pid < maxPID; pid++ {
	// 	// fetch the proper page
	// 	pg, err = c.pool.FetchPage(pid)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	err = pg.RangeRecords(fn)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	return nil, nil
}

func (c *Collection) FindOne(id uint64, ptr Record) error {
	if id == badID {
		return ErrBadID
	}
	// decode id to locate the page
	rid := page.DecodeRecordID(id)
	// fetch the proper page
	pg, err := c.pool.FetchPage(rid.PageID)
	if err != nil {
		return err
	}
	// attempt to locate the record
	rec, err := pg.GetRecord(rid)
	if err != nil {
		return err
	}
	// decode into the pointer provided
	err = decRecord(rec, ptr)
	// err = ptr.UnmarshalBinary(rec.Val())
	if err != nil {
		return err
	}
	// unpin the page (not dirty)
	err = c.pool.UnpinPage(rid.PageID, false)
	if err != nil {
		return err
	}
	return nil
}

func (c *Collection) FindAll() ([]Record, error) {
	return nil, nil
}

func (c *Collection) Insert(data Record) (uint64, error) {
	if c.curr == 0 {
		pg, err := c.pool.NewPage()
		if err != nil {
			return badID, err
		}
		err = c.pool.FlushPage(pg.GetPageID())
		if err != nil {
			return badID, err
		}
	}
fetch:
	// fetch the current page
	pg, err := c.pool.FetchPage(c.curr)
	if err != nil {
		return badID, err
	}
	// encode the data record
	rec, err := encRecord(data)
	if err != nil {
		return badID, err
	}
	// check to ensure it has room
	if !page.HasRoom(pg, rec) {
		// no room, so update the current page
		// and go back to fetch
		// update the current page
		atomic.SwapUint32(&c.curr, c.curr+1)
		goto fetch
	}
	// add the encoded record to the page
	rid, err := pg.AddRecord(rec)
	if err != nil {
		return badID, err
	}
	// unpin the page (make sure to mark dirty)
	err = c.pool.UnpinPage(c.curr, true)
	if err != nil {
		return badID, err
	}
	// return the encoded record id
	return page.EncodeRecordID(rid), nil
}

func (c *Collection) Update(id uint64, data Record) (uint64, error) {
	if id == badID {
		return badID, ErrBadID
	}
fetch:
	// decode id
	rid := page.DecodeRecordID(id)
	// fetch the correct page
	pg, err := c.pool.FetchPage(rid.PageID)
	if err != nil {
		return badID, err
	}
	// encode the data record
	rec, err := encRecord(data)
	if err != nil {
		return badID, err
	}
	// check to ensure it has room
	if !page.HasRoom(pg, rec) {
		// no room, so update the current page
		// and go back to fetch
		// update the current page
		atomic.SwapUint32(&c.curr, c.curr+1)
		goto fetch
	}
	// delete the current data record
	err = pg.DelRecord(rid)
	if err != nil {
		return badID, err
	}
	// add the encoded record to the page
	rid, err = pg.AddRecord(rec)
	if err != nil {
		return badID, err
	}
	// unpin the page (make sure to mark dirty)
	err = c.pool.UnpinPage(c.curr, true)
	if err != nil {
		return badID, err
	}
	// return the encoded record id
	return badID, nil
}

func (c *Collection) Delete(id uint64) error {
	if id == badID {
		return ErrBadID
	}
	rid := page.DecodeRecordID(id)
	pg, err := c.pool.FetchPage(rid.PageID)
	if err != nil {
		return err
	}
	err = pg.DelRecord(rid)
	if err != nil {
		return err
	}
	return nil
}

func (c *Collection) Commit() error {
	err := c.pool.FlushAll()
	if err != nil {
		return err
	}
	return nil
}

func (c *Collection) destroy() error {
	err := c.pool.Close()
	if err != nil {
		return err
	}
	err = os.Remove(c.path)
	if err != nil {
		return err
	}
	return nil
}

func (c *Collection) close() error {
	return c.pool.Close()
}
