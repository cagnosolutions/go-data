package pager

import (
	"io"
	"os"
)

type diskMan struct {
	f    *os.File
	size int64
}

func (dm *diskMan) writePage(pid pageID, pg page) error {
	// get offset
	offset := int64(pid * szPg)
	// set write cursor to offset
	_, err := dm.f.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}
	// write the data from page
	_, err = dm.f.Write(pg)
	if err != nil {
		return err
	}
	// flush
	err = dm.f.Sync()
	if err != nil {
		return err
	}
	return nil
}

func (dm *diskMan) readPage(pid pageID, pg page) error {
	// get offset
	offset := int64(pid * szPg)
	// check if we are beyond file length
	if offset > dm.size {
		return io.ErrUnexpectedEOF
	}
	// set read cursor to offset
	_, err := dm.f.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}
	// read the data into page
	_, err = dm.f.Read(pg)
	if err != nil {
		return err
	}
	// return nil error
	return nil
}

func (dm *diskMan) Close() error {
	return nil
}

// func (l *Log) loadSegmentEntries(s *segment) error {
// 	data, err := ioutil.ReadFile(s.path)
// 	if err != nil {
// 		return err
// 	}
// 	ebuf := data
// 	var epos []bpos
// 	var pos int
// 	for exidx := s.index; len(data) > 0; exidx++ {
// 		var n int
// 		n, err = loadNextBinaryEntry(data)
// 		if err != nil {
// 			return err
// 		}
// 		data = data[n:]
// 		epos = append(epos, bpos{pos, pos + n})
// 		pos += n
// 	}
// 	s.ebuf = ebuf
// 	s.epos = epos
// 	return nil
// }
