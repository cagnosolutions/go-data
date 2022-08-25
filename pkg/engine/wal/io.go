package wal

import (
	"encoding"
)

type BinaryEntry interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

type BinaryReader interface {
	// ReadFrom returns a BinaryReader that is set up to read
	// from the provided path
	ReadFrom(path string) (BinaryReader, error)
	// ReadEntry reads and returns the next sequential entry
	ReadEntry() (BinaryEntry, error)
	// ReadEntryAt reads and returns the entry at the provided offset
	ReadEntryAt(off int64) (BinaryEntry, error)
	// Offset returns the current offset of the cursor
	Offset() (int64, error)
	// Seek fulfills the io.Seeker interface
	Seek(off int64, whence int) (int64, error)
	// Close fulfills the io.Closer interface
	Close() error
}

type BinaryWriter interface {
	// ReadFrom returns a BinaryReader that is set up to read
	// from the provided path
	ReadFrom(path string) (BinaryReader, error)
	// WriteEntry writes the provided entry
	WriteEntry(e BinaryEntry) (int64, error)
	// Offset returns the current offset of the cursor
	Offset() (int64, error)
	// Sync flushes any content to the underlying writer
	Sync(off int64, whence int) (int64, error)
	// Close fulfills the io.Closer interface
	Close() error
}
