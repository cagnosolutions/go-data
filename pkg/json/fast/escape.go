package fast

import (
	"bytes"
	"unicode/utf8"
)

func Unescape(in, out []byte) ([]byte, error) {

	firstBackslash := bytes.IndexByte(in, '\\')
	if firstBackslash == -1 {
		return in, nil
	}

	if cap(out) < len(in) {
		out = make([]byte, len(in))
	} else {
		out = out[0:len(in)]
	}

	copy(out, in[:firstBackslash])
	in = in[firstBackslash:]
	buf := out[firstBackslash:]

	for len(in) > 0 {
		// Unescape the next escaped character
		inLen, bufLen := unescapeToUTF8(in, buf)
		if inLen == -1 {
			return nil, ErrMalformedStringEsc
		}

		in = in[inLen:]
		buf = buf[bufLen:]

		// Copy everything up until the next backslash
		nextBackslash := bytes.IndexByte(in, '\\')
		if nextBackslash == -1 {
			copy(buf, in)
			buf = buf[len(in):]
			break
		} else {
			copy(buf, in[:nextBackslash])
			buf = buf[nextBackslash:]
			in = in[nextBackslash:]
		}
	}

	// Trim the out buffer to the amount that was actually emitted
	return out[:len(out)-len(buf)], nil
}

func unescapeToUTF8(in, out []byte) (inLen int, outLen int) {
	if len(in) < 2 || in[0] != '\\' {
		// Invalid escape due to insufficient characters for any escape or no initial backslash
		return -1, -1
	}

	// https://tools.ietf.org/html/rfc7159#section-7
	switch e := in[1]; e {
	case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
		// Valid basic 2-character escapes (use lookup table)
		out[0] = backslashCharEscapeTable[e]
		return 2, 1
	case 'u':
		// Unicode escape
		if r, inLen := decodeUnicodeEscape(in); inLen == -1 {
			// Invalid Unicode escape
			return -1, -1
		} else {
			// Valid Unicode escape; re-encode as UTF8
			outLen := utf8.EncodeRune(out, r)
			return inLen, outLen
		}
	}

	return -1, -1
}

var backslashCharEscapeTable = [...]byte{
	'"':  '"',
	'\\': '\\',
	'/':  '/',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
}

func decodeUnicodeEscape(in []byte) (rune, int) {
	if r, ok := decodeSingleUnicodeEscape(in); !ok {
		// Invalid Unicode escape
		return utf8.RuneError, -1
	} else if r <= basicMultilingualPlaneOffset && !isUTF16EncodedRune(r) {
		// Valid Unicode escape in Basic Multilingual Plane
		return r, 6
	} else if r2, ok := decodeSingleUnicodeEscape(in[6:]); !ok { // Note: previous decodeSingleUnicodeEscape success guarantees at least 6 bytes remain
		// UTF16 "high surrogate" without manditory valid following Unicode escape for the "low surrogate"
		return utf8.RuneError, -1
	} else if r2 < lowSurrogateOffset {
		// Invalid UTF16 "low surrogate"
		return utf8.RuneError, -1
	} else {
		// Valid UTF16 surrogate pair
		return combineUTF16Surrogates(r, r2), 12
	}
}

// isUTF16EncodedRune checks if a rune is in the range for non-BMP characters,
// which is used to describe UTF16 chars.
// Source: https://en.wikipedia.org/wiki/Plane_(Unicode)#Basic_Multilingual_Plane
func isUTF16EncodedRune(r rune) bool {
	return highSurrogateOffset <= r && r <= basicMultilingualPlaneReservedOffset
}

func decodeSingleUnicodeEscape(in []byte) (rune, bool) {
	// We need at least 6 characters total
	if len(in) < 6 {
		return utf8.RuneError, false
	}

	// Convert hex to decimal
	h1, h2, h3, h4 := h2I(in[2]), h2I(in[3]), h2I(in[4]), h2I(in[5])
	if h1 == badHex || h2 == badHex || h3 == badHex || h4 == badHex {
		return utf8.RuneError, false
	}

	// Compose the hex digits
	return rune(h1<<12 + h2<<8 + h3<<4 + h4), true
}

const badHex = -1

func h2I(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'A' && c <= 'F':
		return int(c - 'A' + 10)
	case c >= 'a' && c <= 'f':
		return int(c - 'a' + 10)
	}
	return badHex
}
