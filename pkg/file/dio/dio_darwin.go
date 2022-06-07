//go:build !windows && !openbsd && !plan9 && !linux && !unix

package dio

const (
	AlignSize = 0    // OSX doesn't need any alignment
	BlockSize = 4096 // Minimum block size
)

// OpenFile is the OSX function used to open and return a file for direct IO.
func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	file, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return
	}
	// F_NOCACHE 	Avoids caching.
	// F_NOCACHE    Turns data caching off/on.
	// 				A non-zero value in arg turns data caching off.
	// 				A value of zero in arg turns data caching on.
	_, _, e1 := syscall.Syscall(syscall.SYS_FCNTL, uintptr(file.Fd()), syscall.F_NOCACHE, 1)
	if e1 != 0 {
		err = fmt.Errorf("Failed to set F_NOCACHE: %s", e1)
		file.Close()
		file = nil
	}
	return
}
