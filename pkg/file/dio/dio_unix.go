//go:build !windows && !darwin && !openbsd && !plan9

package dio

import (
	"os"
	"syscall"
)

const (
	AlignSize = 4096 // Size to align the buffer to
	BlockSize = 4096 // Minimum block size
)

// OpenFile is a modified version of os.OpenFile which sets O_DIRECT
func OpenFile(name string, flag int, perm os.FileMode) (file *os.File, err error) {
	return os.OpenFile(name, syscall.O_DIRECT|flag, perm)
}
