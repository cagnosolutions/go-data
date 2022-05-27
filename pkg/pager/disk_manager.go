package pager

import (
	"os"
	"path/filepath"
)

// diskManager is a storage manager for working with files
// on a long term storage medium, aka the hard drive.
type diskManager struct {
	path   string
	file   *os.File
	nextID pageID
	index  *bitset
	fsize  int64
}

// newDiskManager instantiates and returns a new diskManager.
func newDiskManager(path string) *diskManager {
	// Create or open file
	file, err := initPathAndFile(path)
	if err != nil {
		panic(err)
	}
	// Get file size for calculating page offsets and pageIDs.
	fi, err := file.Stat()
	if err != nil {
		panic(err)
	}
	size := fi.Size()
	// Set up the nextID according to the file size.
	npgs := uint32(size / szPg)
	var nextID pageID
	if npgs > 0 {
		nextID = pageID(npgs + 1)
	}
	// Create a new diskManager instance to return.
	dm := &diskManager{
		path:   path,
		file:   file,
		nextID: nextID,
		fsize:  size,
		index:  newBitset(32),
	}
	return dm
}

// getPageOffset is a helper method that checks to ensure the page is not nil
// and also calculates, checks and returns the logical offset of the provided
// page using the provided pageID.
func (dm *diskManager) getPageOffset(pid pageID, p *page) (int64, error) {
	// First we do a little error checking to ensure the provided page
	// is not nil, and then we will check that the provided offset is
	// not outsize of the bounds of the file.
	if p == nil {
		return -1, ErrNilPage
	}
	// We should now calculate the logical page offset address using the
	// provided pageID.
	offset := int64(pid * szPg)
	// Check to ensure that the calculated offset is not outsize the
	// bounds of the file.
	if offset > dm.fsize {
		return -1, ErrOffsetOutOfBounds
	}
	// Otherwise, our page offset and our page are good, so return.
	return offset, nil
}

// getNextID increments and returns the next pageID
func (dm *diskManager) getNextID() pageID {
	id := dm.nextID
	dm.nextID++
	return id
}

// allocate checks the underlying size of the file and grows it in
// segment sized (2MB) chunks at a time once it reaches an 80% full
// rate. After it checks the size, it will potentially grow the file,
// and then it will return a valid pageID. It may not need to grow
// very often, in which case it will simply increment and return the
// next pageID. If there are a lot of "free" (otherwise deallocated)
// pages in the file, it will attempt to reuse and return those pageID's
// before continuing to increment and return new ones.
func (dm *diskManager) allocate() pageID {
	// First check the size of the underlying file in order to determine
	// if we need to grow it or not.
	sz := getFileSize(dm.path)
	// Next, we will get the number of used un-used pages within the file.
	fp := getFreePageCount(dm.path)
	// Next, determine if the size of the file is close to the 80% full mark.
	if int64(fp*szPg) < sz-(512*1<<10) {
		// I DON'T THINK THE ABOVE CALCULATION IS CORRECT...
		// Grow the underlying file
	}
	// File is not full enough, check for and possibly return free page.
	if fp > 0 {
		// Get and return free page pageID
	}
	// Otherwise, we just increment and return our next pageID
	return dm.getNextID()
}

// deallocate attempts to deallocate a page at the offset that is
// calculated using the provided pageID. It will mark the page status
// as unused allowing it to be used later on. The page's prevPid and
// nextPid will also get removed but the pid will stay there for re-use.
func (dm *diskManager) deallocate(pid pageID) {
	// TODO implement me
	panic("implement me")
}

// read attempts to read a page from the underlying storage. It uses
// the pageID provided to calculate the logical page offset and will
// attempt to read the contents of the page located at the offset.
// Any errors encountered while calculating the logical page offset,
// or while trying to read will be returned.
func (dm *diskManager) read(pid pageID, p page) error {
	// First we do a little error checking, and calculate what the page
	// offset is supposed to be.
	offset, err := dm.getPageOffset(pid, &p)
	if err != nil {
		return err
	}
	// Next, we can attempt to read the contents of the page data
	// directly from the calculated offset. **Using ReadAt makes one
	// syscall, vs using Seek+Read which calls syscall twice.
	n, err := dm.file.ReadAt(p, offset)
	if err != nil {
		return err
	}
	// Check to ensure that we actually read the contents of a full page.
	if n < szPg {
		// **It should be noted that we could also alternatively choose
		// to pad out the remaining byte of the page right in here if
		// we want to. For now though, we will error out. I feel that
		// we will always be wanting to read and write full pages so
		// in that case an error is a valid response.
		return ErrPartialPageRead
	}
	// Finally, we can return a nil error
	return nil
}

// write attempts to write a page to the underlying storage. It uses
// the pageID provided to calculate the logical page offset and will
// attempt to write the contents of the provided page to that offset.
// Any errors encountered while calculating the logical page offset or
// while writing will be returned.
func (dm *diskManager) write(pid pageID, p page) error {
	// First we do a little error checking, and calculate what the page
	// offset is supposed to be.
	offset, err := dm.getPageOffset(pid, &p)
	if err != nil {
		return err
	}
	// Next, we can attempt to write the contents of the page data
	// directly to the calculated offset. **Using WriteAt makes one
	// syscall, vs using Seek+Write which calls syscall twice. We
	// should reserve the Seek+Write pattern only if we are dealing
	// with append only writing type of instance.
	n, err := dm.file.WriteAt(p, offset)
	if err != nil {
		return err
	}
	// Check to ensure that we actually wrote the contents of a full page.
	if n != szPg {
		return ErrPartialPageWrite
	}
	// Update the diskManager file size if necessary.
	if offset >= dm.fsize {
		dm.fsize = offset + int64(n)
	}
	// Before we are finished, we should call sync.
	err = dm.file.Sync()
	if err != nil {
		return err
	}
	// Finally, we can return a nil error because everything is good.
	return nil
}

// size returns the size of the underlying file used by the diskManager.
func (dm *diskManager) size() int64 {
	return getFileSize(dm.path)
}

// close will call Close on the underlying file. Any errors are returned.
func (dm *diskManager) close() error {
	err := dm.file.Close()
	if err != nil {
		return err
	}
	return nil
}

// initPathAndFile takes a path to a file and creates or returns
// it. If the path contains directories that do not exist, the
// directories are created. If the file does not exist, the file
// will be created. If the path and file exist, the file is simply
// opened and returned.
func initPathAndFile(path string) (*os.File, error) {
	// sanitize path
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	// split path
	dir, name := filepath.Split(filepath.ToSlash(path))
	// init files and dirs
	var fp *os.File
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// create dir
		err = os.MkdirAll(dir, os.ModeDir)
		if err != nil {
			return nil, err
		}
		// create file
		fp, err = os.Create(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		// initial file creation, so we will size it
		// to a full segment size.
		err = fp.Truncate(szSg)
		if err != nil {
			return nil, err
		}
		// close file
		err = fp.Close()
		if err != nil {
			return nil, err
		}
	}
	// open existing file
	fp, err = os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	// return file and nil error
	return fp, nil
}

// getFileSize is a helper function that returns the file size for the
// file matching the provided path. If the file cannot be found or if
// any error occurs the resulting file fsize will be -1.
func getFileSize(path string) int64 {
	// Attempt to get the file fsize at the provided path.
	fi, err := os.Stat(path)
	if err != nil {
		return -1
	}
	return fi.Size()
}

// getFreePageCount returns the number of free or unused pages for the
// file matching the provided path.
func getFreePageCount(path string) int {
	return -1
}
