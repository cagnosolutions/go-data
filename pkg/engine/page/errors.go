package page

import (
	"fmt"
)

var ErrPageIDHasNotBeenAllocated = func(pid PageID) error {
	return fmt.Errorf("Page ID has not been allocated yet (PID=%d)", pid)
}

// Page errors
var (
	ErrRecordTooSmall = fmt.Errorf("[Page] record is too small")
	ErrNoRoom         = fmt.Errorf("[Page] the Page is full")
	ErrEmptyPage      = fmt.Errorf("[Page] the Page is empty")
	ErrInvalidPID     = fmt.Errorf("[Page] Page ID is not valid")
	ErrInvalidSID     = fmt.Errorf("[Page] cellPtr ID is not valid")
	ErrRecordNotFound = fmt.Errorf("[Page] record not found")
)
