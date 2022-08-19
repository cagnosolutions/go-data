package dbms

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
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

func getDataFileID(name string) uint32 {
	id, err := strconv.ParseUint(name, 16, 32)
	if err != nil {
		panic(err)
	}
	return uint32(id)
}

// Namespace is a namespace struct
type Namespace struct {
	meta *os.File
	data *os.File
	base string
	name string
}

// OpenNamespace opens and returns a new *Namespace
func OpenNamespace(base, name string) (*Namespace, error) {
	// Add namespace string to base to make complete dir
	dir := filepath.Join(base, name)
	// Strip of any suffixes (if there are any) from the path
	dir = strings.Replace(dir, filepath.Ext(dir), "", -1)
	// Clean the path
	dir, err := filepath.Abs(filepath.ToSlash(dir))
	if err != nil {
		log.Panicf("cleaning path: %s\n", err)
	}
	// Create a new namespace instance we can return later
	ns := &Namespace{
		meta: nil,
		data: nil,
		base: dir,
		name: name,
	}
	// Set up our filenames and file pointers for later
	metaName := filepath.Join(dir, metaFile)
	// Check to see if we need to create the files
	_, err = os.Stat(metaName)
	if os.IsNotExist(err) {
		// Touch any directories and/or file
		err = os.MkdirAll(dir, os.ModeDir|dataFilePerm)
		if err != nil {
			return nil, err
		}
		// Create our meta file
		ns.meta, err = os.OpenFile(metaName, os.O_CREATE|os.O_RDWR|os.O_SYNC, dataFilePerm)
		if err != nil {
			return nil, err
		}
		// Now, we can return without walking the directory (because this is the first time)
		return ns, nil
	}
	// Otherwise, we have to open our existing meta file
	ns.meta, err = os.OpenFile(metaName, os.O_RDWR|os.O_SYNC, dataFilePerm)
	if err != nil {
		return nil, err
	}
	// And then, we can return
	return ns, nil
}

// Walk walks a directory, running the supplied function for each glob match encountered
func (ns *Namespace) Walk(matchGlob string, fn func(de fs.DirEntry) error) error {
	// Walk our directory path
	err := filepath.WalkDir(
		filepath.Join(ns.base, ns.name), func(lpath string, info fs.DirEntry, err error) error {
			// Handle any local path errors
			if err != nil {
				log.Printf("prevent panic by handling failure accessing: %q: %v\n", lpath, err)
				return err
			}
			// Check for local file match
			matched, err := filepath.Match(matchGlob, lpath)
			if !info.IsDir() && matched {
				// We have a match, run function...
				err = fn(info)
				if err != nil {
					return err
				}
				// Otherwise, we will just keep going...
			}
			// Otherwise, return a nil error
			return nil
		},
	)
	// Check to see if there were any errors to return
	if err != nil {
		return err
	}
	return nil
}

// Close closes the namespace and files
func (ns *Namespace) Close() error {
	err := ns.meta.Close()
	if err != nil {
		return err
	}
	// err = ns.data.Close()
	// if err != nil {
	// 	return err
	// }
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
