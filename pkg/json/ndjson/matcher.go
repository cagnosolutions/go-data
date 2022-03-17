package ndjson

import (
	"bytes"
)

const (
	_ = iota
	matchOR
	matchAND
)

func Match(data, pattern []byte, expected int) bool {
	return match(data, pattern, expected)
}

func match(data, pattern []byte, expected int) bool {
	// sanitize and process pattern
	pp := transform(pattern)
	// setup match counter
	var matches int
	// otherwise, lets try our luck out
	for _, field := range bytes.FieldsFunc(data, jsonFF) {
		// range pattern parts
		for _, patt := range pp {
			// check for match
			if matchField(field, patt) {
				// we got one, increment match counter
				matches++
			}
		}
	}
	// return boolean result of matches found compared to
	// our minimum expectation of how many we expected
	return matches >= expected
}

func matchField(field []byte, patternPart []byte) bool {
	// fmt.Printf("in match field: field=%q, patternPart=%q\n", field, patternPart)
	i := bytes.IndexByte(patternPart, ':')
	return bytes.Contains(field, patternPart[:i]) && bytes.Contains(field, patternPart[i+1:])
}

var jsonFF = func(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0:
		return true
	case '{', '}', ',':
		return true
	}
	return false
}

var jsonMF = func(r rune) rune {
	switch r {
	case '+':
		return ':'
	case '"', '*':
		return -1
	}
	return r
}

func transform(p []byte) [][]byte {
	// sanitize and process pattern
	return bytes.Split(bytes.Map(jsonMF, p), []byte{';'})
}
