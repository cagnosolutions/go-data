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

func IsAlpha(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
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

func IsNumeric(ch uint8) bool {
	return '0' <= ch && ch <= '9'
}

// IsPunctuation reports whether the byte is an ASCII english punctuation `.,"'-?:!;/\`
func IsPunctuation(ch uint8) bool {
	// ASCII space is dec 32, and slash is 47. The characters in
	// between are as follows: {space}!"#$%&'()*+,-./
	return ch == 32 || ch == 33 || ch == 34 || ch == 39 || ch == 44 ||
		ch == 45 || ch == 46 || ch == 47 || ch == 58 || ch == 59 ||
		ch == 63 || ch == 92 || ch == 96
}

/*
` ` -> 32 (20)
`!` -> 33 (21)
`"` -> 34 (22)

`'` -> 39 (27)

`,` -> 44 (2c)
`-` -> 45 (2d)
`.` -> 46 (2e)
`/` -> 47 (2f)

`:` -> 58 (3a)
`;` -> 59 (3b)
`?` -> 63 (3f)

`\` -> 92 (5c)
``` -> 96 (60)

*/
