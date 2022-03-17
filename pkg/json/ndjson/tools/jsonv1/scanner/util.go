package scanner

import (
	"errors"
	"fmt"
	"unicode"
	"unicode/utf8"
)

var (
	ErrUnexpectedEOF    = errors.New("unexpected EOF")
	ErrKeyNotFound      = errors.New("key not found")
	ErrIndexOutOfBounds = errors.New("index out of bounds")
	ErrToLessThanFrom   = errors.New("to index less than from index")
	ErrUnexpectedValue  = errors.New("unexpected value")
)

func SkipSpace(in []byte, pos int) (int, error) {
	for {
		r, size := utf8.DecodeRune(in[pos:])
		if size == 0 {
			return 0, ErrUnexpectedEOF
		}
		if !unicode.IsSpace(r) {
			break
		}
		pos += size
	}

	return pos, nil
}

func Expect(in []byte, pos int, content ...byte) (int, error) {
	if pos+len(content) > len(in) {
		return 0, ErrUnexpectedEOF
	}

	for _, b := range content {
		if v := in[pos]; v != b {
			return 0, ErrUnexpectedValue
		}
		pos++
	}

	return pos, nil
}

func NewError(pos int, b byte) error {
	return fmt.Errorf("invalid character at position, %v; %v", pos, string([]byte{b}))
}
