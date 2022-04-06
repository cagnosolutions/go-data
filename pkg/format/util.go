package format

// IsAlphaUpper reports if the supplied character
// is an upper case letter of the english alphabet.
func IsAlphaUpper(ch byte) bool {
	return 'A' <= ch && ch <= 'Z'
}

// IsAlphaLower reports if the supplied character
// is a lower case letter of the english alphabet.
func IsAlphaLower(ch byte) bool {
	return 'a' <= ch && ch <= 'z'
}

// UpperToLower converts an upper case ASCII character
// to a lower case ASCII character, and returns it.
func UpperToLower(ch byte) byte {
	return ch + 32
}

// LowerToUpper converts a lower case ASCII character
// to an upper case ASCII character, and returns it.
func LowerToUpper(ch byte) byte {
	return ch - 32
}
