package compress

const defaultSize = 16

// Prefix compresses while ensuring the prefix maintains
// the ability for accurate comparing and sorting.
func Prefix(b []byte) []byte {
	return PrefixSize(b, defaultSize)
}

//go:noinline
func PrefixSize(b []byte, size int) []byte {
	return nil
}

// Suffix compresses while ensuring the suffix maintains
// the ability for accurate comparing and sorting.
func Suffix(b []byte) []byte {
	return SuffixSize(b, defaultSize)
}

//go:noinline
func SuffixSize(b []byte, size int) []byte {
	return nil
}

// Affix compresses while ensuring the prefix and suffix
// maintain the ability for accurate comparing and sorting.
func Affix(b []byte) []byte {
	return AffixSize(b, defaultSize)
}

//go:noinline
func AffixSize(b []byte, size int) []byte {
	return nil
}
