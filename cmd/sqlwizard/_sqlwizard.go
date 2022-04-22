package main

import (
	"errors"
	"go/parser"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

const (
	srcDir = 1 << iota
	srcFile
	srcExpr

	srcModeFlags = parser.AllErrors | parser.ParseComments
)

var (
	ErrPkgNotSpecified = errors.New("package name not specified")
	ErrPkgNotFound     = errors.New("package not found")
	ErrFileNotFound    = errors.New("file not found")
	ErrLoadingTemplate = errors.New("error loading template")
)

type FilterFunc func(fs.FileInfo) bool

type SQLWizard struct {
	lock   sync.RWMutex
	filter FilterFunc
	dir    string
}

// ParseDir parses a set of files in the directory specified by dir. Only
// entries passing through the filter (and ending in ".go") are considered.
// If the filter is nil, it is skipped, and it attempts to parse all files
// ending in ".go" in the directory specified by dir. All files that are
// successfully parsed are added to the files map.
func (s *SQLWizard) ParseDir(dir string, filter FilterFunc) error {
	// Locker
	s.lock.Lock()
	defer s.lock.Unlock()
	// Attempt to parse the source files provided by dir and filtered
	// through the filter, if one exists.
	pkgs, err := parser.ParseDir(s.fs, dir, filter, srcModeFlags)
	if err != nil {
		return err
	}
	// Attempt to locate package using dir specified.
	astp, found := pkgs[dir]
	if !found {
		return ErrPkgNotFound
	}
	// Parsing was successful, and we now have an *ast.Package type. We must
	// now range the files within this package and add them to our files map.
	for filename, astf := range astp.Files {
		// Read in the source code of the file. The source gets added to the
		// *File in case we need to parse anything using the positional markers
		// found in the syntax tree.
		src, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		// Create a new file instance.
		file := &File{
			Name:   filepath.ToSlash(filename),
			File:   astf,
			Source: src,
		}
		// Sanitize filename
		filename = filepath.Base(filename)
		// Run our struct collector and add collected
		// structs to our structs map.
		s.Structs[filename] = s.collectStructs(file)
		// Add the parsed file to the files map.
		s.Files[filename] = file
	}
	// And finally, we return
	return nil
}
