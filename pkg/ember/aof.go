package ember

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

const dataFilePerm = 1466

type span struct {
	id       int // Index number of this span
	beg, end int // Beginning and ending offsets for this span
}

func (s span) bounds() (int, int) {
	return s.beg, s.end
}

type aof struct {
	fp     *os.File
	index  []span
	prunes int
	seq    int
}

func open(path string) (*aof, error) {
	// Clean path
	path, err := filepath.Abs(filepath.ToSlash(path))
	if err != nil {
		return nil, err
	}
	var fp *os.File
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// Create a new instance
		err = os.MkdirAll(filepath.Dir(path), os.ModeDir|dataFilePerm)
		if err != nil {
			return nil, err
		}
		fp, err = os.OpenFile(path, os.O_CREATE|os.O_TRUNC, dataFilePerm)
		if err != nil {
			return nil, err
		}
		err = fp.Close()
		if err != nil {
			return nil, err
		}
	}
	// Open file at the fully cleaned path
	fp, err = os.OpenFile(path, os.O_RDWR|os.O_SYNC, dataFilePerm)
	if err != nil {
		return nil, err
	}
	// Index our file
	idx, err := indexSpans(fp, '\n', 64)
	if err != nil {
		return nil, err
	}
	// Create and return a new aof instance
	return &aof{
		fp:    fp,
		index: idx,
	}, nil
}

func (a *aof) write(b []byte) (int, error) {
	if b[len(b)-1] != '\n' {
		b = append(b, '\n')
	}
	// get our beginning mark
	beg, err := a.fp.Seek(0, io.SeekCurrent)
	if err != nil {
		return -1, err
	}
	// write to file
	n, err := a.fp.Write(b)
	if err != nil {
		return -1, err
	}
	// check our size
	if len(b) != n {
		return -1, io.ErrShortWrite
	}
	// add our span to our set and adjust the beginning, ending and id variables
	a.index = append(a.index, span{a.seq, int(beg), int(beg) + len(b)})
	a.seq++
	return a.seq - 1, nil
}

func (a *aof) read(id int) ([]byte, error) {
	// see if this id is in our spans
	at, err := a.findSpan(id)
	if err != nil {
		return nil, err
	}
	// get the bounds and make a buffer to read into
	beg, end := a.index[at].bounds()
	buf := make([]byte, end-beg)
	// read at the location
	_, err = a.fp.ReadAt(buf, int64(beg))
	if err != nil {
		return nil, err
	}
	// return data
	return buf, nil
}

func (a *aof) delete(id int) error {
	// see if this id is in our spans
	at, err := a.findSpan(id)
	if err != nil {
		return err
	}
	// we will not be removing any data from
	// the file, but instead from the index
	a.index[at].id = -1
	a.prunes++

	return nil
}

func (a *aof) findSpan(id int) (int, error) {
	// see if this id is in our spans
	for at, sp := range a.index {
		if sp.id == id {
			return at, nil
		}
	}
	return -1, errors.New("could not locate matching span")
}

func (a *aof) pruneFile() error {
	tmp, err := os.CreateTemp("", "tmp-*.aof")
	if err != nil {
		return err
	}
	var buf []byte
	for _, sp := range a.index {
		if sp.id == -1 {
			log.Printf("skipping writing span{id=%d, beg=%d, end=%d}\n", sp.id, sp.beg, sp.end)
			continue
		}
		// read from our main file, into our buffer
		buf = make([]byte, sp.end-sp.beg)
		_, err = a.fp.ReadAt(buf, int64(sp.beg))
		if err != nil {
			return err
		}
		if buf[len(buf)-1] != '\n' {
			buf = append(buf, '\n')
		}
		// write from our buffer into the new temp file
		_, err = tmp.Write(buf)
		if err != nil {
			return err
		}
		// _, err = tmp.Write([]byte{'\n'})
		// if err != nil {
		// 	return err
		// }
	}
	// get our tmp file path name
	tmpPath := tmp.Name()
	// sync our temp file, and close
	err = tmp.Sync()
	if err != nil {
		return err
	}
	err = tmp.Close()
	if err != nil {
		return err
	}
	// get our current file name
	path := a.fp.Name()
	// close our file
	err = a.fp.Close()
	if err != nil {
		return err
	}
	// rename our tmp file to our original file
	err = os.Rename(tmpPath, path)
	if err != nil {
		return err
	}
	// re-open our (used to be) tmp file
	fp, err := os.OpenFile(path, os.O_RDWR|os.O_SYNC, dataFilePerm)
	if err != nil {
		return err
	}
	// re-index the file
	index, err := indexSpans(fp, '\n', 64)
	if err != nil {
		return err
	}
	a.fp = nil
	a.index = nil
	a.fp = fp
	a.index = index
	// we are done!
	return nil
}

func (a *aof) checkPrune() error {
	// check the percent of how much has been pruned
	// and if it's 30% or higher, write a new file.
	p := percent(a.prunes, len(a.index))
	if p >= 0.30 {
		err := a.pruneFile()
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *aof) entries() int {
	return len(a.index) - a.prunes
}

func (a *aof) close() error {
	// check the percent of how much has been pruned
	// and if it's 30% or higher, write a new file.
	err := a.checkPrune()
	if err != nil {
		return err
	}
	err = a.fp.Sync()
	if err != nil {
		return err
	}
	err = a.fp.Close()
	if err != nil {
		return err
	}
	return nil
}

func percent(a, b int) float32 {
	return float32(a) / float32(b)
}

func newLineBytes() int {
	switch runtime.GOOS {
	case "linux", "unix", "darwin":
		return 1 // \n
	case "windows":
		return 2 // \r\n
	}
	return 0
}

// indexSpans can be used to index spans of text based around a delimiter of your
// choice. The size argument allows you to tune it a bit and have some control over
// the overhead used by the function. The delimiter is not included in the returned
// span bound set and empty lines are not ignored.
func indexSpans(r io.Reader, delim byte, size int) ([]span, error) {
	// Setup initial variables for the function
	var id, beg, end int
	// Drop is a helper func used to drop the correct number of bytes. Right now it
	// is mostly used to handle the special case of \n and \r\n
	drop := func(p []byte, c byte) int {
		if c == '\n' {
			if len(p) > 1 && p[len(p)-2] == '\r' {
				return 2
			}
		}
		return 1
	}
	// Initialize our spans
	spans := make([]span, 0, 8)
	// get a new buffered reader set to our determined buffer size.
	br := bufio.NewReaderSize(r, size)
	for {
		// Read up to buffer size length of data and look for the delimiter. If we fill
		// up the buffer and do not find the delimiter we are looking for, we will just
		// keep reading, one buffer length at a time, until we find it. Note: ReadSlice
		// attempts to re-use the same buffer internally, so that helps a lot.
		data, err := br.ReadSlice(delim)
		if err != nil {
			if err == io.EOF {
				// We have reached the end--we are going to check for any remaining data.
				if len(data) > 0 {
					// We have some leftover data, which means the stream was not delimiter
					// terminated. Add the remaining data to one last span before breaking.
					spans = append(spans, span{id + 1, beg, end + len(data)})
				}
				// Otherwise, the stream is indeed delimiter terminated, so we can just break.
				break
			}
			if err == bufio.ErrBufferFull {
				// Our buffer seems to be full, so at this point we will simply update the
				// ending offset and continue reading (skipping all the stuff below, and
				// restarting the loop from the next iteration.)
				end += len(data)
				continue
			}
			// Uh oh, we have some other issue going on.
			return nil, err
		}
		// We were able to locate a delimiter without filling the buffer, so we should update
		// our ending offset; then add our span data to our set.
		end += len(data)
		// Calculate number of bytes to drop
		n := drop(data, delim)
		// Add our span to our set and adjust the beginning, ending and id variables
		spans = append(spans, span{id, beg, end - n})
		// We will grow the beginning offset up to where the end is, and increment the id.
		beg = end
		id++
	}
	return spans, nil
}
