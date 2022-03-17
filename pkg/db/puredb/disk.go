package puredb

import (
	"io"
	"os"
	"path/filepath"
)

func sanitize2(base, file string) string {
	return filepath.ToSlash(filepath.Clean(filepath.Join(base, file)))
}

func getOffset(fp *os.File) int64 {
	off, err := fp.Seek(0, io.SeekCurrent)
	if err != nil {
		panic(err)
	}
	return off
}

func openFile(base, file string) (*os.File, error) {
	// get a correct base before we do anything else
	path := sanitize2(base, file)
	// initialize file pointer
	var fp *os.File
	// check to see of we need to create a new directory structure
	_, err := os.Stat(path)
	// file does not exist
	if os.IsNotExist(err) {
		// create any directories
		err = os.MkdirAll(filepath.Dir(path), 0660)
		if err != nil {
			return nil, err
		}
		// simply continue
	}
	// file may or may not exist, but we use O_CREATE flag
	fp, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return nil, err
	}
	// return file pointer
	return fp, err
}

func removeFile(base, file string) error {
	// get a correct base before we do anything else
	path := sanitize2(base, file)
	// remove file at the base
	return os.Remove(path)
}
