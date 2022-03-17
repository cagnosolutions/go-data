package ndjson

import (
	"bytes"
	"errors"
	"fmt"
	"unicode"
	"unicode/utf8"
)

// FindKey accepts a JSON object and returns the value associated with the key specified
func FindKey(in []byte, pos int, k []byte) ([]byte, error) {
	// The start variable will be available to hold our start position for each type
	// we scan through.
	var start int
	// Skip past any initial space that may be present. We are looking for the beginning
	// of a JSON object, denoted by a left brace. `{`
	pos, err := SkipSpace(in, pos)
	if err != nil {
		return nil, err
	}
	// We have skipped past the space, so we should be at the start of an object. Check
	// to make sure that is the case.
	if v := in[pos]; v != '{' {
		// If we have not found the start of an object, throw an error for now. But it
		// must be noted that *technically* we may be starting with an array of objects
		// and not know it. In which case we would need to look for a left bracket. `[`
		return nil, NewError(pos, v)
	}
	// If we reach this point, we have found our opening left brace '{'. We should now
	// increment the positional counter and then go into our loop.
	pos++
	// We are now inside a JSON object.
	for {
		// Skip any leading whitespace, then look for a string.
		pos, err = SkipSpace(in, pos)
		if err != nil {
			return nil, err
		}

		// Mark our start position, in case we find our key without any errors.
		start = pos
		// Our key will be a string, so lets see if we have it by trying to return
		// the ending position of a string. If we do not encounter an error we
		// have found it.
		pos, err = String(in, pos)
		if err != nil {
			return nil, err
		}
		// We have successfully identified our first key.
		key := in[start+1 : pos-1]
		// Check it against our supplied key to determine if we have a match and
		// store the result of our potential match for later.
		match := bytes.Equal(k, key)
		// It might be worth noting here that maybe we should check to see if we
		// have a match sooner than later, like now. And if we do not have a match,
		// then we can potentially make a choice when we get to our any value call.

		// Next, skip past any potential whitespace.
		pos, err = SkipSpace(in, pos)
		if err != nil {
			return nil, err
		}

		// Look for a colon. If we do not find one, return an error.
		pos, err = Expect(in, pos, ':')
		if err != nil {
			return nil, err
		}
		// Otherwise, consume it and continue.

		// Skip past any potential whitespace.
		pos, err = SkipSpace(in, pos)
		if err != nil {
			return nil, err
		}
		// We are now on the lookout for a value.

		// Mark our start position, in case we find our value without any errors.
		start = pos
		// Our value could be of any type, but we think we have it in our sights
		// and the best way to find out is to try and find the end of it without
		// encountering any errors.
		pos, err = Any(in, pos)
		if err != nil {
			return nil, err
		}
		// We must have found it, because we were not met with any error.

		// Now, lets check to see if we have a match.
		if match {
			// If we do, we will return the value that we have isolated.
			return in[start:pos], nil
		}

		// Otherwise, we did not have a matching key. So we must continue on to
		// inspect more keys. So, we must skip past any potential whitespace.
		pos, err = SkipSpace(in, pos)
		if err != nil {
			return nil, err
		}

		// After which, we will either be met with a comma, indicating that we
		// have more keys to inspect, or the end of the JSON object.
		switch in[pos] {
		case ',':
			// More keys to inspect, so lets increment our positional counter
			// and start the loop over.
			pos++
		case '}':
			// Oh no, we have found the end of the JSON object, and have not
			// located our matching key. Return an error.
			return nil, ErrKeyNotFound
		}
	}
}

// FindIndex accepts a JSON array and return the value of the element at the specified index
func FindIndex(in []byte, pos, index int) ([]byte, error) {
	pos, err := SkipSpace(in, pos)
	if err != nil {
		return nil, err
	}

	if v := in[pos]; v != '[' {
		return nil, NewError(pos, v)
	}
	pos++

	idx := 0
	for {
		pos, err = SkipSpace(in, pos)
		if err != nil {
			return nil, err
		}

		itemStart := pos
		// data
		pos, err = Any(in, pos)
		if err != nil {
			return nil, err
		}
		if index == idx {
			return in[itemStart:pos], nil
		}

		pos, err = SkipSpace(in, pos)
		if err != nil {
			return nil, err
		}

		switch in[pos] {
		case ',':
			pos++
		case ']':
			return nil, ErrIndexOutOfBounds
		}

		idx++
	}
}

// FindRange finds the elements of an array between the specified indexes; inclusive
func FindRange(in []byte, pos, from, to int) ([]byte, error) {
	if to < from {
		return nil, ErrToLessThanFrom
	}

	pos, err := SkipSpace(in, pos)
	if err != nil {
		return nil, err
	}

	if v := in[pos]; v != '[' {
		return nil, NewError(pos, v)
	}
	pos++

	idx := 0
	itemStart := pos

	for {
		pos, err = SkipSpace(in, pos)
		if err != nil {
			return nil, err
		}

		if idx == from {
			itemStart = pos
		}

		// data
		pos, err = Any(in, pos)
		if err != nil {
			return nil, err
		}

		if idx == to {
			data := in[itemStart:pos]
			result := make([]byte, 0, len(data)+2)
			result = append(result, '[')
			result = append(result, data...)
			result = append(result, ']')
			return result, nil
		}

		pos, err = SkipSpace(in, pos)
		if err != nil {
			return nil, err
		}

		switch in[pos] {
		case ',':
			pos++
		case ']':
			return nil, ErrIndexOutOfBounds
		}

		idx++
	}
}

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
