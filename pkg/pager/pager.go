package pager

import (
	"errors"
	"os"
)

var (
	ErrBadPageSize = errors.New("bad page size--page size is not a multiple of 4096")
)

var pageMeta []byte

type Pager struct {
	dm *diskManager
	bp *bufferPool
}

func Open(pfile string, psize, pcount int) (*Pager, error) {
	// check the page size first
	if psize%4096 != 0 {
		return nil, ErrBadPageSize
	}
	var fd *os.File
	// check to see if there is a page file
	_, err := os.Stat(pfile)
	if os.IsNotExist(err) {
		// create new file instance
		fd, err = os.OpenFile(pfile, os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return nil, err
		}
		_, err = fd.Write(pageMeta)
		if err != nil {
			return nil, err
		}
		err = fd.Close()
		if err != nil {
			return nil, err
		}
	}
	// check page file for correct page file size
	fd, err = os.OpenFile(pfile, os.O_RDWR|os.O_SYNC, 0666)
	if err != nil {
		return nil, err
	}
	meta := make([]byte, 4096)
	_, err = fd.Read(meta)
	if err != nil {
		return nil, err
	}
	// TODO: do check and return with any errors
	//
	// open disk manager instance
	dm := newDiskManager(pfile)
	// return new pager instance
	return &Pager{
		dm: dm,
		bp: newBufferPool(psize, dm),
	}, nil
}
