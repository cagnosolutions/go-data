package main

import (
	"errors"
	"os"
	"path"
	"path/filepath"
)

func TouchAny(filePath string) (err error) {
	var dironly bool
	filePath = filepath.ToSlash(filePath)
	// if the filepath ends in a slash
	if filePath[len(filePath)-1] == '/' {
		// treat it like a `mkdir -p` call
		dironly = true
	}
	// otherwise, create the dir first...
	err = os.MkdirAll(path.Dir(filePath), 0655)
	if err != nil && os.IsNotExist(err) {
		return errors.Join(errors.New("error creating dir"), err)
	}
	if dironly {
		return
	}
	// then create the file...
	_, err = os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		fp, err := os.Create(filePath)
		if err != nil {
			return errors.Join(errors.New("error creating file"), err)
		}
		err = fp.Close()
		if err != nil {
			return errors.Join(errors.New("error closing file"), err)
		}
	}
	return
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err != nil && os.IsExist(err)
}
