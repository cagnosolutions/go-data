package pager

import (
	"fmt"
	"io"
	"os"
	"sort"
)

const (
	minPgInSg = 1
	maxPgInSg = int(^uint8(0)) // 255
)

type pageIDs []pageID

func (x pageIDs) Len() int           { return len(x) }
func (x pageIDs) Less(i, j int) bool { return x[i] < x[j] }
func (x pageIDs) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

// diskManager is a storage manager for working with files
// on a long term storage medium, aka the hard drive.
type diskManager struct {
	path    string
	segSize int
	file    *os.File
	free    pageIDs
	fsize   int64
}

// newDiskManager instantiates and returns a new diskManager.
func newDiskManager(path string, pagesInSeg int) *diskManager {
	// initialize path and file (create or open)
	file, err := initPathAndFile(path)
	if err != nil {
		panic(err)
	}
	// get the file size for later
	// fi, err := file.Stat()
	// if err != nil {
	// 	panic(err)
	// }
	// var allocSeg bool
	// size := fi.Size()
	// if size == 0 {
	// 	allocSeg = true
	// }
	// // Set up the nextID according to the file size.
	// npgs := uint32(size / szPg)
	// var nextID pageID
	// if npgs > 0 {
	// 	nextID = npgs + 1
	// }
	// create a new diskManager instance to return.
	if pagesInSeg > maxPgInSg {
		pagesInSeg = maxPgInSg
	}
	dm := &diskManager{
		path:    path,
		segSize: pagesInSeg * szPg,
		file:    file,
		free:    make(pageIDs, 0),
		fsize:   getFileSize(path),
	}
	// if the file was fresh, allocate a segment
	if dm.fsize == 0 {
		err = dm.allocateSegment()
		if err != nil {
			panic(err)
		}
	} else {
		err = dm.loadFreePageIDs()
		if err != nil {
			panic(err)
		}
	}
	return dm
}

// getFileSize returns the current file size
func (dm *diskManager) getFileSize() int64 {
	fi, err := dm.file.Stat()
	if err != nil {
		panic(err)
	}
	return fi.Size()
}

// allocateSegment allocates and initializes a new segment
func (dm *diskManager) allocateSegment() error {
	// go to the end of the file
	off, err := dm.file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	// get the current file size
	size := dm.getFileSize()
	// truncate the file; grow by segment size
	err = dm.file.Truncate(size + int64(dm.segSize))
	if err != nil {
		return err
	}
	// find the last logical page ID after the
	// resize using the offset
	lastPid := (off + int64(dm.segSize)) / szPg
	// write logical page data
	var pg page
	var pid pageID
	slSize := len(dm.free)
	for i := int64(0); i < lastPid; i++ {
		// create an empty page for this offset
		pid = pageID(i)
		pg = newEmptyPage(pid)
		// write empty page
		_, err = dm.file.WriteAt(pg, i*szPg)
		if err != nil {
			return err
		}
		// append empty page ID to free set
		dm.free = append(dm.free[slSize:], pid)
	}
	// sort free page IDs
	sort.Sort(dm.free)
	fmt.Printf("allocating segment: before=%d, now=%d, slice=%v\n", slSize, len(dm.free), dm.free)
	// return
	return nil
}

// loadFreePageIDs reads through the file and populates the free
// list of pageIDs
func (dm *diskManager) loadFreePageIDs() error {
	var pid int64
	buf := make([]byte, 2)
	for {
		_, err := dm.file.ReadAt(buf, (pid*szPg)+4)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return err
		}
		magic := bin.Uint16(buf)
		if magic&stFree > 0 {
			dm.free = append(dm.free, pageID(pid))
		}
		fmt.Printf("page=%d, offset=%d, magic=%.4x\n", pid, pid*szPg, magic)
		pid++
	}
	sort.Sort(dm.free)
	return nil
}

// getOffset is a helper method that checks to ensure the page is not nil
// and also calculates, checks and returns the logical offset of the provided
// page using the provided page ID.
func (dm *diskManager) getOffset(pid pageID, p page) (int64, error) {
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
func (dm *diskManager) getNextID() (pageID, error) {
	if len(dm.free) <= 1 {
		err := dm.allocateSegment()
		if err != nil {
			return 0, err
		}
	}
	next, newFree := dm.free[0], dm.free[1:]
	fmt.Println("NextID:", next)
	(*dm).free = newFree
	return next, nil
	// id := dm.nextID
	// dm.nextID++
	// return id
}

func (dm *diskManager) getFreePageIDs() pageIDs {
	return dm.free
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
	// check the free page count
	pid, err := dm.getNextID()
	if err != nil {
		panic(err)
	}
	return pid
}

// deallocate attempts to deallocate a page at the offset that is
// calculated using the provided pageID. It will mark the page status
// as unused allowing it to be used later on. The page's prevPid and
// nextPid will also get removed but the pid will stay there for re-use.
func (dm *diskManager) deallocate(pid pageID) {
	// TODO implement me
	panic("implement me")
}

// readPage tries to read the page data from the disk. It uses the provided page ID
// to calculate the logical offset of the disk resident page. It attempts to read
// the data into the provided page. If no errors are encountered, the function will
// return a nil error.
func (dm *diskManager) readPage(pid pageID, p page) error {
	// First we do a little error checking, and calculate what the page
	// offset is supposed to be.
	offset, err := dm.getOffset(pid, p)
	if err != nil {
		return err
	}
	// Next, we can attempt to read the contents of the page data
	// directly from the calculated offset. Using ReadAt makes one
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
	offset, err := dm.getOffset(pid, p)
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
	fp, err := fileOpenOrCreate(path)
	if err != nil {
		return nil, err
	}
	return fp, nil

	/*
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
	*/
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
