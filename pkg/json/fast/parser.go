package fast

import (
	"bytes"
	"errors"
	"strconv"
)

var (
	ErrKeyNotFound        = errors.New("key path not found")
	ErrUnknownType        = errors.New("unknown type")
	ErrMalformedJSON      = errors.New("malformed json data")
	ErrMalformedString    = errors.New("value is string but can't find closing '\"' symbol")
	ErrMalformedArray     = errors.New("value is array but can't find closing ']' symbol")
	ErrMalformedObject    = errors.New("value is object but can't find closing '}' symbol")
	ErrMalformedValue     = errors.New("value is number/bool/null but can't find the end (missing ',', ']', or '}')")
	ErrIntegerOverflow    = errors.New("value is number but overflowed while parsing")
	ErrMalformedStringEsc = errors.New("got an invalid escape sequence in a string")
	ErrNullValue          = errors.New("value is null")
)

const stackBufSize = 64

func tokenBeg(p []byte) int {
	for i := len(p) - 1; i >= 0; i-- {
		switch p[i] {
		case '\n', '\r', '\t', ',', '{', '[':
			return i
		}
	}
	return 0
}

func tokenEnd(p []byte) int {
	for i, c := range p {
		switch c {
		case ' ', '\n', '\r', '\t', ',', '}', ']':
			return i
		}
	}
	return len(p)
}

func findTokenStart(p []byte, t byte) int {
	for i := len(p) - 1; i >= 0; i-- {
		switch p[i] {
		case t:
			return i
		case '[', '{':
			return 0
		}
	}
	return 0
}

// nextToken locates the position of the next character
// that is not whitespace
func nextToken(p []byte) int {
	for i, c := range p {
		switch c {
		case ' ', '\n', '\r', '\t':
			continue
		default:
			return i
		}
	}
	return -1
}

func lastToken(p []byte) int {
	for i := len(p) - 1; i >= 0; i-- {
		switch p[i] {
		case ' ', '\n', '\r', '\t':
			continue
		default:
			return i
		}
	}
	return -1
}

func findKeyStart(p []byte, k string) (int, error) {
	i := nextToken(p)
	if i == -1 {
		return i, ErrKeyNotFound
	}
	sz := len(p)
	if sz > 0 && (p[i] == '{' || p[i] == '[') {
		i += 1
	}

	// stack-allocated array for allocation-free un-escaping of small strings
	var stackbuf [stackBufSize]byte

	ku, err := Unescape(stringToBytes(k), stackbuf[:])
	if err == nil {
		k = bytesToString(&ku)
	}

	for i < sz {
		switch p[i] {
		case '"':
			i++
			keyBeg := i
			strEnd, keyEsc := stringEnd(p[i:])
			if strEnd == -1 {
				break
			}
			i += strEnd
			keyEnd := i - 1
			valOff := nextToken(p[i:])
			if valOff == -1 {
				break
			}
			i += valOff

			// if string is a key, and key level match
			key := p[keyBeg:keyEnd]
			// for unescape: if there are no escape sequences, this is cheap. otherwise it
			// is a little bit more expensive but causes no allocations unless len(key) > stackBufSize
			if keyEsc {
				ku, err = Unescape(key, stackbuf[:])
				if err != nil {
					break
				}
				key = ku
			}

			// check for end of key string
			if p[i] == ':' && len(k) == len(key) && bytesToString(&key) == k {
				return keyBeg - 1, nil
			}
		case '[':
			end := blockEnd(p[i:], p[i], ']')
			if end != -1 {
				i += end
			}
		case '{':
			end := blockEnd(p[i:], p[i], '}')
			if end != -1 {
				i += end
			}
		}
		i++
	}

	return 0, nil
}

// stringEnd attempts to find the end of a string (supports strings containing escaped quotes)
func stringEnd(p []byte) (int, bool) {
	var esc bool
	for i, c := range p {
		if c == '"' {
			if !esc {
				return i + 1, false
			}
			j := i - 1
			for {
				if j < 0 || p[j] != '\\' {
					// we have an even number of backslashes
					return i + 1, true
				}
				j--
				if j < 0 || p[j] != '\\' {
					// we have an odd number of backslashes
					break
				}
				j--
			}
		}
		if c == '\\' {
			// we have an escape sequence
			esc = true
		}
	}
	return -1, esc
}

// blockEnd attempts to locate the end of the data structure, array or object. For an array
// the openSym and closeSym will be '[' and ']' and for an object it will be '{' and '}'
func blockEnd(p []byte, openSym, closeSym byte) int {
	var i, level int

	for i < len(p) {
		switch p[i] {
		case '"':
			// if we are inside a string, skip
			se, _ := stringEnd(p[i+1:])
			if se == -1 {
				return -1
			}
			i += se
		case openSym:
			// if we find an openSym, increase level
			level++
		case closeSym:
			// if we find an closeSym, decrease level
			level--
			// and if we are at the original level, we are done
			if level == 0 {
				return i + 1
			}
		}
		i++
	}
	return -1
}

func searchKeys(p []byte, keys ...string) int {
	var keyLevel, level, i int
	lastMatched := true

	sz := len(p)
	szKeys := len(keys)

	if szKeys == 0 {
		return 0
	}

	// stack-allocated array for allocation-free un-escaping of small strings
	var stackbuf [stackBufSize]byte

	for i < sz {
		switch p[i] {
		case '"':
			i++
			keyBeg := i

			strEnd, keyEsc := stringEnd(p[i:])
			if strEnd == -1 {
				return -1
			}
			i += strEnd
			keyEnd := i - 1

			valOff := nextToken(p[i:])
			if valOff == -1 {
				return -1
			}
			i += valOff

			// and now, we can do this if the string is a key
			if p[i] == ':' {
				if level < 1 {
					return -1
				}
				key := p[keyBeg:keyEnd]

				// for unescape: if there are no escape sequences, this is cheap. otherwise it
				// is a little bit more expensive but causes no allocations unless len(key) > stackBufSize
				var keyUnesc []byte
				if !keyEsc {
					keyUnesc = key
					continue
				}
				ku, err := Unescape(key, stackbuf[:])
				if err != nil {
					return -1
				}
				keyUnesc = ku

				if level <= szKeys {
					if equalStr(&keyUnesc, keys[level-1]) {
						lastMatched = true

						if keyLevel == level-1 {
							keyLevel++
							if keyLevel == szKeys {
								return i + 1
							}
						}
						continue
					}
					lastMatched = false
					continue
				}
				return -1
			}
			i--
		case '{':
			if !lastMatched {
				end := blockEnd(p[i:], '{', '}')
				if end == -1 {
					return -1
				}
				i += end - 1
				continue
			}
			level++
		case '}':
			level--
			if level == keyLevel {
				keyLevel--
			}
		case '[':
			if keyLevel == level && keys[level][0] == '[' {
				keyLen := len(keys[level])
				if keyLen < 3 || keys[level][0] != '[' || keys[level][keyLen-1] != ']' {
					return -1
				}
				aIdx, err := strconv.Atoi(keys[level][1 : keyLen-1])
				if err != nil {
					return -1
				}
				var curIdx, valOff int
				var valFound []byte
				curI := i
				arrayEach(
					p[i:], func(val []byte, datType ValueType, off int, err error) {
						if curIdx == aIdx {
							valFound = val
							valOff = off
							if datType == String {
								valOff -= 2
								valFound = p[curI+valOff : curI+valOff+len(val)+2]
							}
						}
						curIdx += 1
					},
				)

				if valFound == nil {
					return -1
				}
				subIdx := searchKeys(valFound, keys[level+1:]...)
				if subIdx < 0 {
					return -1
				}
				return i + valOff + subIdx
			}
			// otherwise, we do not have anything to search for inside array
			arraySkip := blockEnd(p[i:], '[', ']')
			if arraySkip == -1 {
				return -1
			}
			i += arraySkip - 1
		case ':':
			// if we find this, then the JSON data is malformed
			return -1
		}
		i++
	}
	return -1
}

func arrayEach(p []byte, cb func(val []byte, dataType ValueType, off int, err error), keys ...string) (int, error) {

	var offset int

	if len(p) == 0 {
		return -1, ErrMalformedObject
	}

	nT := nextToken(p)
	if nT == -1 {
		return -1, ErrMalformedJSON
	}

	offset = nT + 1

	if len(keys) > 0 {
		if offset = searchKeys(p, keys...); offset == -1 {
			return offset, ErrKeyNotFound
		}

		// Go to closest value
		nO := nextToken(p[offset:])
		if nO == -1 {
			return offset, ErrMalformedJSON
		}

		offset += nO

		if p[offset] != '[' {
			return offset, ErrMalformedArray
		}

		offset++
	}

	nO := nextToken(p[offset:])
	if nO == -1 {
		return offset, ErrMalformedJSON
	}

	offset += nO

	if p[offset] == ']' {
		return offset, nil
	}

	for true {
		val, typ, off, err := get(p[offset:])

		if err != nil {
			return offset, err
		}

		if off == 0 {
			break
		}

		if typ != NotExist {
			cb(val, typ, offset+off-len(val), err)
		}

		if err != nil {
			break
		}

		offset += off

		skipToToken := nextToken(p[offset:])
		if skipToToken == -1 {
			return offset, ErrMalformedArray
		}
		offset += skipToToken

		if p[offset] == ']' {
			break
		}

		if p[offset] != ',' {
			return offset, ErrMalformedArray
		}

		offset++
	}

	return offset, nil
}

// get is a convenience wrapper for the internalGet call
func get(p []byte, keys ...string) ([]byte, ValueType, int, error) {
	val, dataType, _, endOff, err := internalGet(p, keys...)
	return val, dataType, endOff, err
}

// internalGet returns the value, dataType, begOffset, endOffset and any potential errors
func internalGet(p []byte, keys ...string) ([]byte, ValueType, int, int, error) {
	var offset int

	if len(keys) > 0 {
		if offset = searchKeys(p, keys...); offset == -1 {
			return nil, NotExist, -1, -1, ErrKeyNotFound
		}
	}

	// Go to closest value
	nO := nextToken(p[offset:])
	if nO == -1 {
		return nil, NotExist, offset, -1, ErrMalformedJSON
	}

	offset += nO
	value, dataType, endOffset, err := getType(p, offset)
	if err != nil {
		return value, dataType, offset, endOffset, err
	}

	// Strip quotes from string values
	if dataType == String {
		value = value[1 : len(value)-1]
	}

	return value[:len(value):len(value)], dataType, offset, endOffset, nil
}

func getType(data []byte, offset int) ([]byte, ValueType, int, error) {
	var dataType ValueType
	endOffset := offset

	// if string value
	if data[offset] == '"' {
		dataType = String
		idx, _ := stringEnd(data[offset+1:])
		if idx == -1 {
			return nil, dataType, offset, ErrMalformedString
		}
		endOffset += idx + 1
		goto done
	}

	// if array value
	if data[offset] == '[' { // if array value
		dataType = Array
		// break label, for stopping nested loops
		endOffset = blockEnd(data[offset:], '[', ']')
		if endOffset == -1 {
			return nil, dataType, offset, ErrMalformedArray
		}
		endOffset += offset
		goto done
	}

	if data[offset] == '{' { // if object value
		dataType = Object
		// break label, for stopping nested loops
		endOffset = blockEnd(data[offset:], '{', '}')

		if endOffset == -1 {
			return nil, dataType, offset, ErrMalformedObject
		}

		endOffset += offset
		goto done

	}

	// otherwise, try to deal with a potential Number, Boolean or None
	if data[offset] != '"' && data[offset] != '[' && data[offset] != '{' {
		end := tokenEnd(data[endOffset:])

		if end == -1 {
			return nil, dataType, offset, ErrMalformedValue
		}

		value := data[offset : endOffset+end]

		switch data[offset] {
		case 't', 'f': // true or false
			if bytes.Equal(value, trueLiteral) || bytes.Equal(value, falseLiteral) {
				dataType = Boolean
			} else {
				return nil, Unknown, offset, ErrUnknownType
			}
		case 'u', 'n': // undefined or null
			if bytes.Equal(value, nullLiteral) {
				dataType = Null
			} else {
				return nil, Unknown, offset, ErrUnknownType
			}
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
			dataType = Number
		default:
			return nil, Unknown, offset, ErrUnknownType
		}
		endOffset += end
	}

done:
	return data[offset:endOffset], dataType, endOffset, nil
}
