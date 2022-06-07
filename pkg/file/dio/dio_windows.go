//go:build !darwin && !openbsd && !plan9 && !linux && !unix

package dio

import (
	"os"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

const (
	AlignSize = 4096 // Size to align the buffer to
	BlockSize = 4096 // Minimum block size

	winFlagNoBuffering  = 0x20000000 // Windows flag to disable buffering
	winFlagWriteThrough = 0x80000000 // Windows flag to enable write through

	// This error is an unexported error. The source for this can be found at the
	// location provided in the link below.
	//
	// [https://cs.opensource.google/go/go/+/master:src/syscall/syscall_windows.go;l=145]
	//
	_ERROR_BAD_NETPATH = syscall.Errno(53)
)

// makeInheritedSa is an unexported function in the syscall package. The source for
// the original function can be found at the location provided in the link below.
//
// [https://cs.opensource.google/go/go/+/master:src/syscall/syscall_windows.go;l=300]
//
func makeInheritSa() *syscall.SecurityAttributes {
	var sa syscall.SecurityAttributes
	sa.Length = uint32(unsafe.Sizeof(sa))
	sa.InheritHandle = 1
	return &sa
}

// syscallOpen is a modified version of the syscall.Open function. References for the
// syscall.Open (and the syscall.CreateFile) functions are listed below for reference
// and modification sake.
//
// TODO: This function may need to be updated from time to time if the language source
//  for either of these changes in a significant way.
//
// Source for syscall.Open. This last time this source changed was in version 1.14
// [https://cs.opensource.google/go/go/+/master:src/syscall/syscall_windows.go;l=308]
//
// Source for syscall.CreateFile. Last time this source changed was in version 1.15
// [https://cs.opensource.google/go/go/+/master:src/syscall/zsyscall_windows.go;l=506]
//
func syscallOpen(path string, mode int, perm uint32) (syscall.Handle, error) {
	if len(path) == 0 {
		return syscall.InvalidHandle, syscall.ERROR_FILE_NOT_FOUND
	}
	pathp, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return syscall.InvalidHandle, err
	}
	var access uint32
	switch mode & (os.O_RDONLY | os.O_WRONLY | os.O_RDWR) {
	case os.O_RDONLY:
		access = syscall.GENERIC_READ
	case os.O_WRONLY:
		access = syscall.GENERIC_WRITE
	case os.O_RDWR:
		access = syscall.GENERIC_READ | syscall.GENERIC_WRITE
	}
	if mode&syscall.O_CREAT != 0 {
		access |= syscall.GENERIC_WRITE
	}
	if mode&os.O_APPEND != 0 {
		access &^= syscall.GENERIC_WRITE
		access |= syscall.FILE_APPEND_DATA
	}
	sharemode := uint32(syscall.FILE_SHARE_READ | syscall.FILE_SHARE_WRITE)
	var sa *syscall.SecurityAttributes
	if mode&syscall.O_CLOEXEC == 0 {
		sa = makeInheritSa()
	}
	var createmode uint32
	switch {
	case mode&(syscall.O_CREAT|syscall.O_EXCL) == (syscall.O_CREAT | syscall.O_EXCL):
		createmode = syscall.CREATE_NEW
	case mode&(syscall.O_CREAT|syscall.O_TRUNC) == (syscall.O_CREAT | syscall.O_TRUNC):
		createmode = syscall.CREATE_ALWAYS
	case mode&syscall.O_CREAT == syscall.O_CREAT:
		createmode = syscall.OPEN_ALWAYS
	case mode&syscall.O_TRUNC == syscall.O_TRUNC:
		createmode = syscall.TRUNCATE_EXISTING
	default:
		createmode = syscall.OPEN_EXISTING
	}
	var attrs uint32 = syscall.FILE_ATTRIBUTE_NORMAL
	if perm&syscall.S_IWRITE == 0 {
		attrs = syscall.FILE_ATTRIBUTE_READONLY
		if createmode == syscall.CREATE_ALWAYS {
			// We have been asked to create a read-only file.
			// If the file already exists, the semantics of
			// the Unix open system call is to preserve the
			// existing permissions. If we pass CREATE_ALWAYS
			// and FILE_ATTRIBUTE_READONLY to CreateFile,
			// and the file already exists, CreateFile will
			// change the file permissions.
			// Avoid that to preserve the Unix semantics.
			h, e := syscall.CreateFile(
				pathp, access, sharemode, sa,
				syscall.TRUNCATE_EXISTING,
				syscall.FILE_ATTRIBUTE_NORMAL|winFlagNoBuffering|winFlagWriteThrough, 0,
			)
			switch e {
			case syscall.ERROR_FILE_NOT_FOUND, _ERROR_BAD_NETPATH, syscall.ERROR_PATH_NOT_FOUND:
				// File does not exist. These are the same
				// errors as Errno.Is checks for ErrNotExist.
				// Carry on to create the file.
			default:
				// Success or some different error.
				return h, e
			}
		}
	}
	h, e := syscall.CreateFile(
		pathp, access, sharemode, sa,
		createmode,
		attrs|winFlagNoBuffering|winFlagWriteThrough, 0,
	)
	return h, e
}

// syscallMode is a local make of the unexported version is the os package. The source for
// the original function can be found at the location provided in the link below.
//
//  [https://cs.opensource.google/go/go/+/master:src/os/file_posix.go;l=62]
//
func syscallMode(i os.FileMode) (o uint32) {
	o |= uint32(i.Perm())
	if i&os.ModeSetuid != 0 {
		o |= syscall.S_ISUID
	}
	if i&os.ModeSetgid != 0 {
		o |= syscall.S_ISGID
	}
	if i&os.ModeSticky != 0 {
		o |= syscall.S_ISVTX
	}
	// No mapping for Go's ModeTemporary (plan9 only).
	return
}

func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	fh, err := syscallOpen(name, flag|syscall.O_CLOEXEC, syscallMode(perm))
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fh), name), nil
}

// utf16FromString returns the UTF-16 encoding of the UTF-8 string
// s, with a terminating NUL added. If s contains a NUL byte at any
// location, it returns (nil, EINVAL).
//
// FIXME copied from go source
func _utf16FromString(s string) ([]uint16, error) {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return nil, syscall.EINVAL
		}
	}
	return utf16.Encode([]rune(s + "\x00")), nil
}

// OpenFile is a modified version of os.OpenFile which sets the
// passes the following flags to windows CreateFile.
//
// The FileFlagNoBuffering takes this concept one step further and
// eliminates all read-ahead file buffering and disk caching as well,
// so that all reads are guaranteed to come from the file and not from
// any system buffer or disk cache. When using FileFlagNoBuffering,
// disk reads and writes must be done on sector boundaries, and buffer
// addresses must be aligned on disk sector boundaries in memory.
//
// FIXME copied from go source then modified
func _OpenFile(path string, mode int, perm os.FileMode) (file *os.File, err error) {
	if len(path) == 0 {
		return nil, &os.PathError{"open", path, syscall.ERROR_FILE_NOT_FOUND}
	}
	pathp, err := _utf16FromString(path)
	if err != nil {
		return nil, &os.PathError{"open", path, err}
	}
	var access uint32
	switch mode & (os.O_RDONLY | os.O_WRONLY | os.O_RDWR) {
	case os.O_RDONLY:
		access = syscall.GENERIC_READ
	case os.O_WRONLY:
		access = syscall.GENERIC_WRITE
	case os.O_RDWR:
		access = syscall.GENERIC_READ | syscall.GENERIC_WRITE
	}
	if mode&syscall.O_CREAT != 0 {
		access |= syscall.GENERIC_WRITE
	}
	if mode&os.O_APPEND != 0 {
		access &^= syscall.GENERIC_WRITE
		access |= syscall.FILE_APPEND_DATA
	}
	sharemode := uint32(syscall.FILE_SHARE_READ | syscall.FILE_SHARE_WRITE)
	var sa *syscall.SecurityAttributes
	var createmode uint32
	switch {
	case mode&(syscall.O_CREAT|os.O_EXCL) == (syscall.O_CREAT | os.O_EXCL):
		createmode = syscall.CREATE_NEW
	case mode&(syscall.O_CREAT|os.O_TRUNC) == (syscall.O_CREAT | os.O_TRUNC):
		createmode = syscall.CREATE_ALWAYS
	case mode&syscall.O_CREAT == syscall.O_CREAT:
		createmode = syscall.OPEN_ALWAYS
	case mode&os.O_TRUNC == os.O_TRUNC:
		createmode = syscall.TRUNCATE_EXISTING
	default:
		createmode = syscall.OPEN_EXISTING
	}
	h, e := syscall.CreateFile(
		&pathp[0], access, sharemode, sa, createmode,
		syscall.FILE_ATTRIBUTE_NORMAL|winFlagNoBuffering|winFlagWriteThrough, 0,
	)
	if e != nil {
		return nil, &os.PathError{"open", path, e}
	}
	return os.NewFile(uintptr(h), path), nil
}
