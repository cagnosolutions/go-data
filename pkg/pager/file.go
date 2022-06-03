package pager

import (
	"os"
	"path/filepath"
)

const (
	flagTouch = os.O_CREATE | os.O_TRUNC
	flagOpen  = os.O_RDWR | os.O_SYNC
	// the leading 1 is the sticky bit
	permMode = 1466 // 0=none, 1=exec, 2=write, 3=exec+write, 4=read, 5=exec+read, 6=write+read, 7=write+exec+read
)

func pathSplit(path string) (dir, file string) {
	return filepath.Split(filepath.ToSlash(path))
}

func pathClean(path string) string {
	path, err := filepath.Abs(path)
	if err != nil {
		panic("pathClean: " + err.Error())
	}
	return filepath.ToSlash(path)
}

func fileTouch(path string) error {
	fp, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC, permMode)
	if err != nil {
		return err
	}
	err = fp.Close()
	if err != nil {
		return err
	}
	return nil
}

func fileMake(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, permMode)
}

func fileOpen(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_SYNC, permMode)
}

func fileOpenOrMake(path string) (*os.File, error) {
	dir, name := pathSplit(pathClean(path))
	full := filepath.Join(dir, name)
	if _, err := os.Stat(full); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModeDir|permMode)
		if err != nil {
			return nil, err
		}
		err = fileTouch(full)
		if err != nil {
			return nil, err
		}
	}
	return fileOpen(full)
}
