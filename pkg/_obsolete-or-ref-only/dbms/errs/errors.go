package errs

import (
	"errors"
	"fmt"
)

// page errors
var (
	ErrRecordTooSmall = fmt.Errorf("[page] record is too small")
	ErrNoRoom         = fmt.Errorf("[page] the page is full")
	ErrEmptyPage      = fmt.Errorf("[page] the page is empty")
	ErrInvalidPID     = fmt.Errorf("[page] page ID is not valid")
	ErrInvalidSID     = fmt.Errorf("[page] slot ID is not valid")
	ErrRecordNotFound = fmt.Errorf("[page] record not found")
)

// segment errors
var (
	ErrSegmentSizeTooSmall     = errors.New("segment size is too small (min 2 MiB)")
	ErrSegmentSizeTooLarge     = errors.New("segment size is too large (max 32 MiB)")
	ErrSegmentHeaderShortWrite = errors.New("buffer is to small to write segment header into")
	ErrSegmentHeaderShortRead  = errors.New("buffer is to small to read segment header from")
	ErrSegmentNotFound         = errors.New("segment has not been found")
)

var (
	ErrPageNotFound = errors.New("page could not be found")
	ErrPageInUse    = errors.New("page is currently in use or has not been unpinned")
	ErrNilPage      = errors.New("page is nil")

	ErrOffsetOutOfBounds = errors.New("calculated offset is outside file bounds")
	ErrPartialPageWrite  = errors.New("page write was not a full page")
	ErrPartialPageRead   = errors.New("page read was not a full page")
	ErrBadPageSize       = errors.New("bad page size--page size is not a multiple of 4096")
	ErrSlotIDOutOfBounds = errors.New("slot id is outside of the lower bounds")
	ErrMinRecSize        = errors.New("record is smaller than the minimum allowed record size")
	ErrMaxRecSize        = errors.New("record is larger than the maximum allowed record size")
	ErrPossiblePageFull  = errors.New("page might be full (but may have fragmented space available)")
	ErrPageFull          = errors.New("page is full and out of available space")
	ErrBadRID            = errors.New("bad record id; either the page id or the slot id did not match")
	ErrRecNotFound       = errors.New("record has not been found")

	ErrMetaInfoMismatch    = errors.New("meta file information does not match provided information")
	ErrUsableFrameNotFound = errors.New("usable frame ID could not be found; this is not good")
	ErrOpeningDiskManager  = errors.New("unable to open disk manager")
	ErrMetaFileExists      = errors.New("meta info file already exists")
	ErrMetaFileNotExists   = errors.New("meta info file does not exists")
	ErrDataFileExists      = errors.New("meta info file already exists")
	ErrDataFileNotExists   = errors.New("meta info file does not exists")

	ErrWriteFileHeader = errors.New("there was an issue writing the file header")
	ErrReadFileHeader  = errors.New("there was an issue reading the file header")
	ErrCRCFileHeader   = errors.New("the crc checksum does not match in the file header")
)
