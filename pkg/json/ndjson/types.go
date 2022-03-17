package ndjson

import (
	"errors"
)

// Any returns the position of the end of the current element that begins at pos; handles any valid json element
func Any(in []byte, pos int) (int, error) {
	pos, err := SkipSpace(in, pos)
	if err != nil {
		return 0, err
	}
	switch in[pos] {
	case '{':
		return Object(in, pos)
	case '[':
		return Array(in, pos)
	case '"':
		return String(in, pos)
	case '.', '-', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
		return Number(in, pos)
	case 't', 'f':
		return Boolean(in, pos)
	case 'n':
		return Null(in, pos)
	default:
		max := len(in) - pos
		if max > 20 {
			max = 20
		}
		return 0, OpErr{
			Pos:     pos,
			Msg:     "invalid object",
			Content: string(in[pos : pos+max]),
		}
	}
}

// Object returns the position of the end of the object that begins at the specified pos
func Object(in []byte, pos int) (int, error) {
	pos, err := SkipSpace(in, pos)
	if err != nil {
		return 0, err
	}

	if v := in[pos]; v != '{' {
		return 0, NewError(pos, v)
	}
	pos++

	// clean initial spaces
	pos, err = SkipSpace(in, pos)
	if err != nil {
		return 0, err
	}

	if in[pos] == '}' {
		return pos + 1, nil
	}

	for {
		// key
		pos, err = String(in, pos)
		if err != nil {
			return 0, err
		}

		// leading spaces
		pos, err = SkipSpace(in, pos)
		if err != nil {
			return 0, err
		}

		// colon
		pos, err = Expect(in, pos, ':')
		if err != nil {
			return 0, err
		}

		// data
		pos, err = Any(in, pos)
		if err != nil {
			return 0, err
		}

		pos, err = SkipSpace(in, pos)
		if err != nil {
			return 0, err
		}

		switch in[pos] {
		case ',':
			pos++
		case '}':
			return pos + 1, nil
		}
	}
}

// Array returns the position of the end of the array that begins at the position specified
func Array(in []byte, pos int) (int, error) {
	pos, err := SkipSpace(in, pos)
	if err != nil {
		return 0, err
	}

	if v := in[pos]; v != '[' {
		return 0, NewError(pos, v)
	}
	pos++

	// clean initial spaces
	pos, err = SkipSpace(in, pos)
	if err != nil {
		return 0, err
	}

	if in[pos] == ']' {
		return pos + 1, nil
	}

	for {
		// data
		pos, err = Any(in, pos)
		if err != nil {
			return 0, err
		}

		pos, err = SkipSpace(in, pos)
		if err != nil {
			return 0, err
		}

		switch in[pos] {
		case ',':
			pos++
		case ']':
			return pos + 1, nil
		}
	}
}

// AsArray accepts an []byte encoded json array as an input and returns the array's elements
func AsArray(in []byte, pos int) ([][]byte, error) {
	pos, err := SkipSpace(in, pos)
	if err != nil {
		return nil, err
	}

	if v := in[pos]; v != '[' {
		return nil, NewError(pos, v)
	}
	pos++

	// clean initial spaces
	pos, err = SkipSpace(in, pos)
	if err != nil {
		return nil, err
	}

	if in[pos] == ']' {
		return [][]byte{}, nil
	}

	// 1. Count the number of elements in the array

	start := pos

	elements := make([][]byte, 0, 256)
	for {
		pos, err = SkipSpace(in, pos)
		if err != nil {
			return nil, err
		}

		start = pos

		// data
		pos, err = Any(in, pos)
		if err != nil {
			return nil, err
		}
		elements = append(elements, in[start:pos])

		pos, err = SkipSpace(in, pos)
		if err != nil {
			return nil, err
		}

		switch in[pos] {
		case ',':
			pos++
		case ']':
			return elements, nil
		}
	}
}

// String returns the position of the string that begins at the specified pos
func String(in []byte, pos int) (int, error) {
	pos, err := SkipSpace(in, pos)
	if err != nil {
		return 0, err
	}

	max := len(in)

	if v := in[pos]; v != '"' {
		return 0, NewError(pos, v)
	}
	pos++

	for {
		switch in[pos] {
		case '\\':
			if in[pos+1] == '"' {
				pos++
			}
		case '"':
			return pos + 1, nil
		}
		pos++

		if pos >= max {
			break
		}
	}

	return 0, errors.New("unclosed string")
}

// Number returns the end position of the number that begins at the specified pos
func Number(in []byte, pos int) (int, error) {
	pos, err := SkipSpace(in, pos)
	if err != nil {
		return 0, err
	}

	max := len(in)
	for {
		v := in[pos]
		switch v {
		case '-', '+', '.', 'e', 'E', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			pos++
		default:
			return pos, nil
		}
		if pos >= max {
			return pos, nil
		}
	}
	return pos, nil
}

var (
	t = []byte("true")
	f = []byte("false")
)

// Boolean matches a boolean at the specified position
func Boolean(in []byte, pos int) (int, error) {
	switch in[pos] {
	case 't':
		return Expect(in, pos, t...)
	case 'f':
		return Expect(in, pos, f...)
	default:
		return 0, ErrUnexpectedValue
	}
}

var (
	n = []byte("null")
)

// Null verifies the contents of bytes provided is a null starting as pos
func Null(in []byte, pos int) (int, error) {
	switch in[pos] {
	case 'n':
		return Expect(in, pos, n...)
		return pos + 4, nil
	default:
		return 0, ErrUnexpectedValue
	}
}

type OpErr struct {
	Pos     int
	Msg     string
	Content string
}

func (o OpErr) Error() string {
	return o.Msg + "; ..." + o.Content
}
