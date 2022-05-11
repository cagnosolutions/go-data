package spage

import (
	"errors"
)

var (
	ErrBadPageSize             = errors.New("raw-page: got incorrect or misaligned page size")
	ErrBadAlignmentSize        = errors.New("pageManagerFile: bad alignment size")
	ErrNoMoreRoomInPage        = errors.New("Page: there is not enough room left in the Page")
	ErrInvalidRecordID         = errors.New("Page: invalid record id")
	ErrRecordHasBeenMarkedFree = errors.New("Page: record has been marked free (aka, removed)")
	ErrRecordNotFound          = errors.New("Page: record could not be found")
	ErrPageNotFound            = errors.New("pageManagerFile: Page could not be found")
	ErrWritingPage             = errors.New("pageManagerFile: error writing Page")
	ErrDeletingPage            = errors.New("pageManagerFile: error deleting Page")
	ErrMinRecordSize           = errors.New("Page: record is smaller than the min record size allowed")
	ErrMaxRecordSize           = errors.New("Page: record is larger than the max record size allowed")
	ErrRecordMaxKeySize        = errors.New("record: record key is longer than max size allowed (255)")
	ErrPageIsNotOverflow       = errors.New("pagemanager: error Page is not an overflow Page")
)
