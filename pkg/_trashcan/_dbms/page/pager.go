package page

// Pager represents the main API of a page buffer or cache.
// Individual implementations may differ quite a bit.
type Pager interface {

	// WritePage takes the data provided and writes it to the
	// page using the PageID provided. If successful a nil
	// error is returned.
	WritePage(d []byte, pid PageID) error

	// ReadPage attempts to read the contents of the page using
	// the PageID provided. If successful a nil error is returned.
	ReadPage(pid PageID) ([]byte, error)

	// SyncPage performs a file system fsync which forces the OS
	// and disk controller to flush the buffered contents of the
	// page found using the PageID provided onto the physical media.
	// If successful, a nil error is returned.
	SyncPage(pid PageID) error

	// FreePage attempts to free the page found using the PageID
	// provided. If the page is currently dirty, it first attempts
	// to call fsync. A boolean indicating the success of the free
	// call will be returned along with a nil error, if successful.
	FreePage(pid PageID) (bool, error)

	// WriteRecord takes the record data provided and writes it to
	// the page using the PageID provided. If successful a RecordID
	// will be returned along with a nil error. The RecordID can be
	// used at a later time to delete or update the specific record.
	WriteRecord(d []byte, pid PageID) (RecordID, error)

	// ReadRecord attempts to read and return a copy of the contents
	// of the selected record using the PageID and RecordID provided.
	// If successful, a nil error will be returned.
	ReadRecord(pid PageID, rid RecordID) ([]byte, error)

	// DeleteRecord attempts to delete the contents of the selected
	// record using the PageID and RecordID provided. A boolean will
	// be returned indicating the success of the call to delete and
	// if successful, a nil error will be returned.
	DeleteRecord(pid PageID, rid RecordID) error
}

// type PageID uint32
type RecordID uint32
