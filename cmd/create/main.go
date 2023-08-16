package main

import (
	"errors"
	"log"
	"os"
	"path"
	"path/filepath"
)

func MkdirAll(path string) (err error) {
	err = os.MkdirAll(path, 0655)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}
	return err
}

func TouchFile(filePath string) (err error) {
	_, err = os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		// file does not exist, create it
		fp, err := os.Create(filePath)
		if err != nil {
			return errors.Join(errors.New("error creating file"), err)
		}
		err = fp.Close()
		if err != nil {
			return errors.Join(errors.New("error closing file"), err)
		}
		return
	}
	// // file does exist, update times
	// currentTime := time.Now().Local()
	// err = os.Chtimes(filePath, currentTime, currentTime)
	// if err != nil {
	// 	return errors.Join(errors.New("error changing file timestamps"), err)
	// }
	return
}

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

func main() {
	err := TouchAny("/cmd/create/files/file1.txt")
	if err != nil {
		log.Println(err)
	}

	err = TouchAny("/cmd/create/files2/")
	if err != nil {
		log.Println(err)
	}
}
