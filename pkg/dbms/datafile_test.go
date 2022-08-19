package dbms

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDataFile_All(t *testing.T) {
	// open a data file
	df, err := OpenDataFile("testing", 0)
	if err != nil {
		t.Error(err)
	}
	defer func(df *DataFile) {
		err := df.Close()
		if err != nil {
			t.Error(err)
		}
	}(df)
}

const (
	matchGlob   = "**.db"
	filePrefix  = "dat-"
	fileSuffix  = ".db"
	metaFile    = filePrefix + "meta" + fileSuffix
	newDataFile = filePrefix + "current" + fileSuffix
)

func makeDataFileName(id int) string {
	return fmt.Sprintf("%s%.8x%s", filePrefix, id, fileSuffix)
}

// Namespace is a namespace struct
type Namespace struct {
	meta  *os.File
	data  *os.File
	dir   string
	files []string
}

func OpenNamespace(base, namespace string) (*Namespace, error) {
	// Add namespace string to base to make complete dir
	dir := filepath.Join(base, namespace)
	// Strip of any suffixes (if there are any) from the path
	dir = strings.Replace(dir, filepath.Ext(dir), "", -1)
	// Clean the path
	dir, err := filepath.Abs(filepath.ToSlash(dir))
	if err != nil {
		log.Panicf("cleaning path: %s\n", err)
	}
	// Create a new namespace instance we can return later
	ns := &Namespace{
		meta:  nil,
		data:  nil,
		dir:   dir,
		files: make([]string, 0),
	}
	// Set up our filenames and file pointers for later
	metaName := filepath.Join(dir, metaFile)
	dataName := filepath.Join(dir, newDataFile)
	// Check to see if we need to create the files
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		// Touch any directories and/or file
		err = os.MkdirAll(dir, os.ModeDir|dataFilePerm)
		if err != nil {
			return nil, err
		}
		// Open our meta file, and our new data file
		ns.meta, err = os.OpenFile(metaName, os.O_CREATE|os.O_RDWR|os.O_SYNC, dataFilePerm)
		if err != nil {
			return nil, err
		}
		ns.data, err = os.OpenFile(dataName, os.O_CREATE|os.O_RDWR|os.O_SYNC, dataFilePerm)
		if err != nil {
			return nil, err
		}
		// Add our files to our files set in the namespace
		ns.files = append(ns.files, filepath.Base(metaName))
		ns.files = append(ns.files, filepath.Base(dataName))
		// Now, we can return without walking the directory (because this is the first time)
		return ns, nil
	}
	// Otherwise, we have to walk the directory
	err = filepath.WalkDir(
		dir, func(lpath string, info fs.DirEntry, err error) error {
			// Handle any local path errors
			if err != nil {
				log.Printf("prevent panic by handling failure accessing: %q: %v\n", lpath, err)
				return err
			}
			// Check for local file match
			matched, err := filepath.Match(matchGlob, lpath)
			if !info.IsDir() && matched {
				// We have a match, append to our namespace file list
				ns.files = append(ns.files, filepath.Base(lpath))
				// Check to see if the file is the meta file
				if lpath == metaName {
					// It is the meta file, so we will open it up
					ns.meta, err = os.OpenFile(metaName, os.O_CREATE|os.O_RDWR|os.O_SYNC, dataFilePerm)
					if err != nil {
						return err
					}
				}
				// Check to see if the file is the current data file
				if lpath == dataName {
					// It is the current data file, so we will open it up
					ns.data, err = os.OpenFile(dataName, os.O_CREATE|os.O_RDWR|os.O_SYNC, dataFilePerm)
					if err != nil {
						return err
					}
				}
				// Otherwise, we will just keep going...
			}
			// Otherwise, return a nil error
			return nil
		},
	)
	return ns, nil
}

func (ns *Namespace) Close() error {
	err := ns.meta.Close()
	if err != nil {
		return err
	}
	err = ns.data.Close()
	if err != nil {
		return err
	}
	return nil
}

func TestDataFile_Namespace(t *testing.T) {
	// open namespace
	ns, err := OpenNamespace("data/db", "users")
	if err != nil {
		t.Error(err)
	}
	// don't forget to close
	defer func(ns *Namespace) {
		err := ns.Close()
		if err != nil {
			t.Error(err)
		}
	}(ns)
}
