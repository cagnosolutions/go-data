package parsing_utils

import (
	"strconv"
)

// UnescapedStackBuffSize is how much stack space to allocate in bytes. If a string
// longer than this needs to be escaped, it will result in a heap allocation.
const unescapedStackBuffSize = 64

// tokenEnd returns the index of any of the bytes representing the end of a type
// which also includes whitespace.
func tokenEnd(data []byte) int {
	for i, c := range data {
		switch c {
		case ' ', '\n', '\r', '\t', ',', '}', ']':
			return i
		}
	}

	return len(data)
}

// findTokenStart returns the index
func findTokenStart(data []byte, token byte) int {
	for i := len(data) - 1; i >= 0; i-- {
		switch data[i] {
		case token:
			return i
		case '[', '{':
			return 0
		}
	}

	return 0
}

func findKeyStart(data []byte, key string) (int, error) {
	i := nextToken(data)
	if i == -1 {
		return i, KeyPathNotFoundError
	}
	ln := len(data)
	if ln > 0 && (data[i] == '{' || data[i] == '[') {
		i += 1
	}
	var stackbuf [unescapedStackBuffSize]byte // stack-allocated array for allocation-free unescaping of small strings

	if ku, err := Unescape(StringToBytes(key), stackbuf[:]); err == nil {
		key = bytesToString(&ku)
	}

	for i < ln {
		switch data[i] {
		case '"':
			i++
			keyBegin := i

			strEnd, keyEscaped := stringEnd(data[i:])
			if strEnd == -1 {
				break
			}
			i += strEnd
			keyEnd := i - 1

			valueOffset := nextToken(data[i:])
			if valueOffset == -1 {
				break
			}

			i += valueOffset

			// if string is a key, and key level match
			k := data[keyBegin:keyEnd]
			// for unescape: if there are no escape sequences, this is cheap; if there are, it is a
			// bit more expensive, but causes no allocations unless len(key) > unescapedStackBuffSize
			if keyEscaped {
				if ku, err := Unescape(k, stackbuf[:]); err != nil {
					break
				} else {
					k = ku
				}
			}

			if data[i] == ':' && len(key) == len(k) && bytesToString(&k) == key {
				return keyBegin - 1, nil
			}

		case '[':
			end := blockEnd(data[i:], data[i], ']')
			if end != -1 {
				i = i + end
			}
		case '{':
			end := blockEnd(data[i:], data[i], '}')
			if end != -1 {
				i = i + end
			}
		}
		i++
	}

	return -1, KeyPathNotFoundError
}

func tokenStart(data []byte) int {
	for i := len(data) - 1; i >= 0; i-- {
		switch data[i] {
		case '\n', '\r', '\t', ',', '{', '[':
			return i
		}
	}

	return 0
}

// Find position of next character which is not whitespace
func nextToken(data []byte) int {
	for i, c := range data {
		switch c {
		case ' ', '\n', '\r', '\t':
			continue
		default:
			return i
		}
	}

	return -1
}

// Find position of last character which is not whitespace
func lastToken(data []byte) int {
	for i := len(data) - 1; i >= 0; i-- {
		switch data[i] {
		case ' ', '\n', '\r', '\t':
			continue
		default:
			return i
		}
	}

	return -1
}

// Tries to find the end of string
// Support if string contains escaped quote symbols.
func stringEnd(data []byte) (int, bool) {
	escaped := false
	for i, c := range data {
		if c == '"' {
			if !escaped {
				return i + 1, false
			} else {
				j := i - 1
				for {
					if j < 0 || data[j] != '\\' {
						return i + 1, true // even number of backslashes
					}
					j--
					if j < 0 || data[j] != '\\' {
						break // odd number of backslashes
					}
					j--

				}
			}
		} else if c == '\\' {
			escaped = true
		}
	}

	return -1, escaped
}

// Find end of the data structure, array or object.
// For array openSym and closeSym will be '[' and ']', for object '{' and '}'
func blockEnd(data []byte, openSym byte, closeSym byte) int {
	level := 0
	i := 0
	ln := len(data)

	for i < ln {
		switch data[i] {
		case '"': // If inside string, skip it
			se, _ := stringEnd(data[i+1:])
			if se == -1 {
				return -1
			}
			i += se
		case openSym: // If open symbol, increase level
			level++
		case closeSym: // If close symbol, increase level
			level--

			// If we have returned to the original level, we're done
			if level == 0 {
				return i + 1
			}
		}
		i++
	}

	return -1
}

func searchKeys(data []byte, keys ...string) int {
	keyLevel := 0
	level := 0
	i := 0
	ln := len(data)
	lk := len(keys)
	lastMatched := true

	if lk == 0 {
		return 0
	}

	var stackbuf [unescapedStackBuffSize]byte // stack-allocated array for allocation-free unescaping of small strings

	for i < ln {
		switch data[i] {
		case '"':
			i++
			keyBegin := i

			strEnd, keyEscaped := stringEnd(data[i:])
			if strEnd == -1 {
				return -1
			}
			i += strEnd
			keyEnd := i - 1

			valueOffset := nextToken(data[i:])
			if valueOffset == -1 {
				return -1
			}

			i += valueOffset

			// if string is a key
			if data[i] == ':' {
				if level < 1 {
					return -1
				}

				key := data[keyBegin:keyEnd]

				// for unescape: if there are no escape sequences, this is cheap; if there are, it is a
				// bit more expensive, but causes no allocations unless len(key) > unescapedStackBuffSize
				var keyUnesc []byte
				if !keyEscaped {
					keyUnesc = key
				} else if ku, err := Unescape(key, stackbuf[:]); err != nil {
					return -1
				} else {
					keyUnesc = ku
				}

				if level <= len(keys) {
					if equalStr(&keyUnesc, keys[level-1]) {
						lastMatched = true

						// if key level match
						if keyLevel == level-1 {
							keyLevel++
							// If we found all keys in path
							if keyLevel == lk {
								return i + 1
							}
						}
					} else {
						lastMatched = false
					}
				} else {
					return -1
				}
			} else {
				i--
			}
		case '{':

			// in case parent key is matched then only we will increase the level otherwise can directly
			// can move to the end of this block
			if !lastMatched {
				end := blockEnd(data[i:], '{', '}')
				if end == -1 {
					return -1
				}
				i += end - 1
			} else {
				level++
			}
		case '}':
			level--
			if level == keyLevel {
				keyLevel--
			}
		case '[':
			// If we want to get array element by index
			if keyLevel == level && keys[level][0] == '[' {
				keyLen := len(keys[level])
				if keyLen < 3 || keys[level][0] != '[' || keys[level][keyLen-1] != ']' {
					return -1
				}
				aIdx, err := strconv.Atoi(keys[level][1 : keyLen-1])
				if err != nil {
					return -1
				}
				var curIdx int
				var valueFound []byte
				var valueOffset int
				curI := i
				ArrayEach(
					data[i:], func(value []byte, dataType ValueType, offset int, err error) {
						if curIdx == aIdx {
							valueFound = value
							valueOffset = offset
							if dataType == String {
								valueOffset = valueOffset - 2
								valueFound = data[curI+valueOffset : curI+valueOffset+len(value)+2]
							}
						}
						curIdx += 1
					},
				)

				if valueFound == nil {
					return -1
				} else {
					subIndex := searchKeys(valueFound, keys[level+1:]...)
					if subIndex < 0 {
						return -1
					}
					return i + valueOffset + subIndex
				}
			} else {
				// Do not search for keys inside arrays
				if arraySkip := blockEnd(data[i:], '[', ']'); arraySkip == -1 {
					return -1
				} else {
					i += arraySkip - 1
				}
			}
		case ':': // If encountered, JSON data is malformed
			return -1
		}

		i++
	}

	return -1
}

func sameTree(p1, p2 []string) bool {
	minLen := len(p1)
	if len(p2) < minLen {
		minLen = len(p2)
	}

	for pi_1, p_1 := range p1[:minLen] {
		if p2[pi_1] != p_1 {
			return false
		}
	}

	return true
}

const stackArraySize = 128

func EachKey(data []byte, cb func(int, []byte, ValueType, error), paths ...[]string) int {
	var x struct{}
	var level, pathsMatched, i int
	ln := len(data)

	pathFlags := make([]bool, stackArraySize)[:]
	if len(paths) > cap(pathFlags) {
		pathFlags = make([]bool, len(paths))[:]
	}
	pathFlags = pathFlags[0:len(paths)]

	var maxPath int
	for _, p := range paths {
		if len(p) > maxPath {
			maxPath = len(p)
		}
	}

	pathsBuf := make([]string, stackArraySize)[:]
	if maxPath > cap(pathsBuf) {
		pathsBuf = make([]string, maxPath)[:]
	}
	pathsBuf = pathsBuf[0:maxPath]

	for i < ln {
		switch data[i] {
		case '"':
			i++
			keyBegin := i

			strEnd, keyEscaped := stringEnd(data[i:])
			if strEnd == -1 {
				return -1
			}
			i += strEnd

			keyEnd := i - 1

			valueOffset := nextToken(data[i:])
			if valueOffset == -1 {
				return -1
			}

			i += valueOffset

			// if string is a key, and key level match
			if data[i] == ':' {
				match := -1
				key := data[keyBegin:keyEnd]

				// for unescape: if there are no escape sequences, this is cheap; if there are, it is a
				// bit more expensive, but causes no allocations unless len(key) > unescapedStackBuffSize
				var keyUnesc []byte
				if !keyEscaped {
					keyUnesc = key
				} else {
					var stackbuf [unescapedStackBuffSize]byte
					if ku, err := Unescape(key, stackbuf[:]); err != nil {
						return -1
					} else {
						keyUnesc = ku
					}
				}

				if maxPath >= level {
					if level < 1 {
						cb(-1, nil, Unknown, MalformedJsonError)
						return -1
					}

					pathsBuf[level-1] = bytesToString(&keyUnesc)
					for pi, p := range paths {
						if len(p) != level || pathFlags[pi] || !equalStr(&keyUnesc, p[level-1]) || !sameTree(
							p, pathsBuf[:level],
						) {
							continue
						}

						match = pi

						pathsMatched++
						pathFlags[pi] = true

						v, dt, _, e := Get(data[i+1:])
						cb(pi, v, dt, e)

						if pathsMatched == len(paths) {
							break
						}
					}
					if pathsMatched == len(paths) {
						return i
					}
				}

				if match == -1 {
					tokenOffset := nextToken(data[i+1:])
					i += tokenOffset

					if data[i] == '{' {
						blockSkip := blockEnd(data[i:], '{', '}')
						i += blockSkip + 1
					}
				}

				if i < ln {
					switch data[i] {
					case '{', '}', '[', '"':
						i--
					}
				}
			} else {
				i--
			}
		case '{':
			level++
		case '}':
			level--
		case '[':
			var ok bool
			arrIdxFlags := make(map[int]struct{})
			pIdxFlags := make([]bool, len(paths))

			if level < 0 {
				cb(-1, nil, Unknown, MalformedJsonError)
				return -1
			}

			for pi, p := range paths {
				if len(p) < level+1 || pathFlags[pi] || p[level][0] != '[' || !sameTree(p, pathsBuf[:level]) {
					continue
				}
				if len(p[level]) >= 2 {
					aIdx, _ := strconv.Atoi(p[level][1 : len(p[level])-1])
					arrIdxFlags[aIdx] = x
					pIdxFlags[pi] = true
				}
			}

			if len(arrIdxFlags) > 0 {
				level++

				var curIdx int
				arrOff, _ := ArrayEach(
					data[i:], func(value []byte, dataType ValueType, offset int, err error) {
						if _, ok = arrIdxFlags[curIdx]; ok {
							for pi, p := range paths {
								if pIdxFlags[pi] {
									aIdx, _ := strconv.Atoi(p[level-1][1 : len(p[level-1])-1])

									if curIdx == aIdx {
										of := searchKeys(value, p[level:]...)

										pathsMatched++
										pathFlags[pi] = true

										if of != -1 {
											v, dt, _, e := Get(value[of:])
											cb(pi, v, dt, e)
										}
									}
								}
							}
						}

						curIdx += 1
					},
				)

				if pathsMatched == len(paths) {
					return i
				}

				i += arrOff - 1
			} else {
				// Do not search for keys inside arrays
				if arraySkip := blockEnd(data[i:], '[', ']'); arraySkip == -1 {
					return -1
				} else {
					i += arraySkip - 1
				}
			}
		case ']':
			level--
		}

		i++
	}

	return -1
}
