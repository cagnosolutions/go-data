package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const pageSize = 8 << 10 // 8 KB

type SimplePager struct {
	index map[uint32]struct {
		records   int
		freeSpace int
	}
	pager map[uint32][pageSize]byte // a map of pages
}

func NewSimplePager(pages int) *SimplePager {
	return &SimplePager{
		index: make(map[uint32]struct {
			records   int
			freeSpace int
		}),
		pager: make(map[uint32][pageSize]byte, pages),
	}
}

func (sp *SimplePager) WriteRecord(pageID uint32, rec []byte) error {
	// check free space for this page
	if sp.index[pageID].freeSpace < len(rec)+1 {
		// this record cannot fit in this page
		return errors.New("no more room in this page")
	}
	// otherwise, it can...

	page, found := sp.pager[pageID]
	if !found {
		return errors.New("no such page")
	}
	page = append(page, rec...)
}

// OpenFile takes a directory and file or just a file name. It creates any
// directories that need to be created and creates or opens a new file.
func OpenFile(filePath string) *os.File {

	// separate the directory path so we can create it if we need to
	dirName := filepath.Dir(filePath)

	// check to see if the directory exists
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		// just a double check to make sure we actually had a directry structure
		if dirName != "" {
			// the directory does not exist, so we will create it
			err := os.MkdirAll(dirName, os.ModePerm)
			if err != nil {
				log.Fatalf("error creating directory: %q\n", err)
			}
		}
	}
	// check to see if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// the file does not exist, so we will create it
		fp, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, dataFilePerm)
		if err != nil {
			log.Fatalf("error creating file: %q\n", err)
		}
		// return newly created file pointer
		return fp
	}
	// the file does exist, so simply open it and return the file pointer
	fp, err := os.OpenFile(filePath, os.O_RDWR|os.O_SYNC, dataFilePerm)
	if err != nil {
		log.Fatalf("error opening file (%q): %q\n", filePath, err)
	}
	return fp
}

// Record represents a single data record and implements the interfaces
// BinaryMarshaler and BinaryUnmarshaler from the encoding package.
type Record struct {
	Header uint32
	Length uint32
	Data   []byte
}

func (rec *Record) UnmarshalBinary(data []byte) error {
	// if the Record is empty reassign to a new pointer
	// note: I don't really even know if this matters.
	if rec == nil {
		*rec = *new(Record)
	}
	// decode the record header
	rec.Header = binary.NativeEndian.Uint32(data[0:4])
	// decode the record length
	rec.Length = binary.NativeEndian.Uint32(data[4:8])
	// initialize and set the data
	rec.Data = make([]byte, rec.Length)
	copy(rec.Data, data[8:])

	return nil
}

func (rec *Record) MarshalBinary() (data []byte, err error) {
	// if the Record is empty, return
	if rec == nil {
		return nil, ErrRecordEmpty
	}
	// otherwise, we must create a new buffer to write into
	buf := make([]byte, recordPadding+len(rec.Data))

	// encode record header and length into buffer
	binary.NativeEndian.PutUint32(buf[0:4], rec.Header)
	binary.NativeEndian.PutUint32(buf[4:8], rec.Length)

	// write record data into buffer
	copy(buf[8:], rec.Data)

	// return buffer and nil error
	return buf, nil
}

func (rec *Record) isValid() error {
	if rec == nil {
		return ErrRecordEmpty
	}
	if rec.Header == 0 {
		return ErrRecordHeaderInvalid
	}
	if rec.Length == 0 {
		return ErrRecordLengthInvalid
	}
	if rec.Data == nil || len(rec.Data) < 1 {
		return ErrRecordDataInvalid
	}
	if rec.Data[len(rec.Data)-1] != recordSepDelim {
		return ErrRecordHasNoSeparator
	}
	return nil
}

func main() {

	fileName := "data/1/baz.txt"

	df, err := OpenDataFile(fileName)
	if err != nil {
		log.Fatalf("error opening data file: %q\n", err)
	}

	df.WriteRecord([]byte("Hello, World!"))

	defer df.Close()
}

const (
	// file permissions
	dataFilePerm = 1466

	// record separator delimiter
	recordSepDelim = 0x1E // ASCII RECORD SEPARATOR

	recordStatusActive  = 0xCAFEBEEF
	recordStatusDeleted = 0xDEADBEEF

	recordPadding = 9 // 8 bytes for the header, and 1 for the separator
)

//type Record struct {
//	Status uint32
//	Length uint32
//	Data   []byte
//}
//
//func (rec *Record) WriteTo(w io.Writer) (int64, error) {
//
//	// create buffer to write into (the plus 9 is 8 for the
//	// header, and 1 for the record separator)
//	buf := make([]byte, len(rec.Data)+9)
//
//	// write header
//	binary.BigEndian.PutUint32(buf[0:4], rec.Status)
//	binary.BigEndian.PutUint32(buf[4:8], rec.Length)
//
//	// write data
//	copy(buf[8:], rec.Data)
//	buf[8+len(rec.Data)] = recordSepDelim
//
//	// write buffer to underlying writer
//	n, err := w.Write(buf)
//	if err != nil {
//		return int64(n), err
//	}
//	return int64(n), nil
//}
//
//func (rec *Record) ReadFrom(r io.Reader) (int64, error) {
//	return -1, nil
//}
//
//func DeleteRecord(r *Record) {
//	r.Status = recordStatusDeleted
//	for i := range r.Data {
//		r.Data[i] = 0
//	}
//}

var (
	ErrRecordHeaderInvalid  = errors.New("invalid record header")
	ErrRecordEmpty          = errors.New("record is empty")
	ErrRecordLengthInvalid  = errors.New("invalid record length")
	ErrRecordDataInvalid    = errors.New("invalid record data")
	ErrRecordHasNoSeparator = errors.New("record has no separator")
)

type DataFile struct {
	sync.RWMutex
	fp *os.File
}

// OpenDataFile opens and return a new *DataFile. If there is already
// a file name matching the provided name it is used, otherwise a new
// file with that name is created.
func OpenDataFile(filePath string) (*DataFile, error) {
	return &DataFile{
		fp: OpenFile(filePath),
	}, nil
}

func (df *DataFile) WriteRecord(rec *Record) error {
	// lock (and unlock) for safety
	df.RLock()
	defer df.RUnlock()

	// check to ensure the record valid
	err := rec.isValid()
	if err != nil {
		return err
	}

	// encode record
	data, err := rec.MarshalBinary()
	if err != nil {
		return err
	}

	// write record
	_, err = df.fp.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (df *DataFile) ReadRecord() (*Record, error) {
	// lock (and unlock) for safety
	df.RLock()
	defer df.RUnlock()

	// create a new record to return
	rec := new(Record)

	// check to ensure the record valid
	err := rec.isValid()
	if err != nil {
		return nil, err
	}

}

//func (df *DataFile) WriteRecord(p []byte) error {
//	// lock for safety
//	df.Lock()
//	defer df.Unlock()
//	// create new record
//	rec := &Record{
//		Status: recordStatusActive,
//		Length: uint32(len(p) + 1),
//		Data:   p,
//	}
//	// write record to underlying file
//	_, err := rec.WriteTo(df.fp)
//	if err != nil {
//		return err
//	}
//	// return
//	return nil
//}
//
//func (df *DataFile) ReadRecord() (*Record, error) {
//	// lock for safety
//	df.RLock()
//	defer df.RUnlock()
//	// create new record to read data into
//	rec := new(Record)
//	_, err := rec.ReadFrom(df.fp)
//	if err != nil {
//		return nil, err
//	}
//	return rec, nil
//}

func (df *DataFile) Close() error {
	// sync any remaining file buffer contents to disk
	err := df.fp.Sync()
	if err != nil {
		return err
	}
	// close file pointer
	err = df.fp.Close()
	if err != nil {
		return err
	}
	return nil
}

func ListDir(dir string) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Printf("dir: %v: name: %s\n", info.IsDir(), path)
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
}
