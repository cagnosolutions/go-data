package lsm

import (
	"os"
	"path/filepath"
)

func clean(path string) string {
	full, err := filepath.Abs(filepath.ToSlash(path))
	if err != nil {
		panic("clean: " + err.Error())
	}
	return full
}

func initBasePath(base string) (string, error) {
	path, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	// sanitize any path separators
	path = filepath.ToSlash(path)
	// create any directories if they are not there
	err = os.MkdirAll(path, os.ModeDir)
	if err != nil {
		return "", err
	}
	// return "sanitized" path
	return path, nil
}

func createDirsAndFiles(path string) {
	// clean path
	root := clean(filepath.Dir(path))
	// make dirs
	err := os.MkdirAll(root, 0644)
	if err != nil {
		panic(err)
	}
	var fd *os.File

	// create file
	fd, err = os.Create(filepath.Join(root, filepath.Base(path)))
	if err != nil {
		panic(err)
	}
	err = fd.Close()
	if err != nil {
		panic(err)
	}
}
