package pager

var ()

type Pager struct {
	dm *diskManager
	bp *bufferPool
}

// Open attempts to open and return, or create and return an instance of a Pager.
// It takes a file name or path ending with a file name, strips the suffix and
// writes to that file. It also creates a meta file with information in it that the
// pager can use.
func Open(path string, pageSize, pageCount uint32) (*Pager, error) {
	// First, check the page size to ensure it is acceptable.
	if pageSize&(szPg-1) != 0 {
		return nil, ErrBadPageSize
	}
	// Next, we want to sanitize the provided path, and trim the provided file suffix.
	full, err := pathCleanAndTrimSUffix(path)
	if err != nil {
		return nil, err
	}
	// Then, we can pass the path to the disk manager instantiation instance.
	dm, err := newDiskManager(full, pageSize, pageCount)
	if err != nil {
		return nil, err
	}
	// Finally, we can instantiate and return a new *Pager instance.
	return &Pager{
		dm: dm,
		bp: newBufferPool(uint16(pageSize), int(pageCount), dm),
	}, nil
}
