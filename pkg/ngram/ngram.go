package ngram

import (
	"errors"
)

func invertCase(ch byte) byte {
	// ch ^= 0x20 // the number 32
	return ch ^ 0x20
}

func isLower(ch byte) bool {
	return (ch | 0x20) == 0
}

func isUpper(ch byte) bool {
	return (ch & 0x20) == 0
}

func toLower(ch *byte) {
	*ch |= 0x20
}

func toUpper(ch *byte) {
	*ch &= 0x20
}

type NgramFn func(g []byte, beg, end int) bool

// NgramScanner will iterate a string and emit n-grams into the provided function
// NGramFn. NGram takes n, which is the gram size you wish to isolate. If len(p)
// (which is the data provided) is less than the n size provided, an error will
// be returned.
func NgramScanner(n int, p []byte, fn NgramFn) error {
	if len(p) < n {
		return errors.New("the provided data is not large enough")
	}
	if fn == nil {
		return errors.New("the provided NgramFn must not be nil")
	}
	for i, j := 0, n; j < len(p); i, j = i+1, j+1 {
		for k := i; k < j; k++ {
			if isUpper(p[k]) {
				toLower(&p[k])
			}
		}
		if !fn(p[i:j], i, j) {
			break
		}
	}
	return nil
}

func NNgramScanner(lo, hi int, p []byte, fn NgramFn) error {
	if len(p) < hi {
		return errors.New("the provided data is not large enough")
	}
	if fn == nil {
		return errors.New("the provided NgramFn must not be nil")
	}
	for i, j := 0, hi; j < len(p); i, j = i+1, j+1 {
		for k := i; k < j; k++ {
			if isUpper(p[k]) {
				toLower(&p[k])
			}
		}
		// TODO: make this work
		if !fn(p[i:j], i, j) {
			break
		}
	}
	return nil
}

var gram = map[string]byte{
	"!":  0x00,
	"@":  0x00,
	"#":  0x00,
	"$":  0x00,
	"%":  0x00,
	"^":  0x00,
	"&":  0x00,
	"*":  0x00,
	"(":  0x00,
	")":  0x00,
	" ":  0x00,
	"_":  0x00,
	"+":  0x00,
	"-":  0x00,
	"=":  0x00,
	"\\": 0x00,
	"|":  0x00,
	"`":  0x00,
	"~":  0x00,
	"/":  0x00,
	"?":  0x00,
	".":  0x00,
	",":  0x00,
	"\"": 0x00,
	"'":  0x00,
	";":  0x00,
	":":  0x00,
	"<":  0x00,
	">":  0x00,
	"{":  0x00,
	"}":  0x00,
	"[":  0x00,
	"]":  0x00,
}

var bigram = map[string]byte{
	"th": 0x00,
	"he": 0x01,
	"in": 0x02,
	"er": 0x03,
	"an": 0x04,
	"re": 0x05,
	"nd": 0x06,
	"on": 0x07,
	"en": 0x08,
	"at": 0x09,
	"ou": 0x0a,
	"ed": 0x0b,
	"ha": 0x0c,
	"to": 0x0d,
	"or": 0x0e,
	"it": 0x0f,
	"is": 0x10,
	"hi": 0x11,
	"es": 0x12,
	"ng": 0x13,
	"nt": 0x14,
	"ti": 0x15,
	"se": 0x16,
	"ar": 0x17,
	"al": 0x18,
	"te": 0x19,
	"co": 0x1a,
	"de": 0x1b,
	"ra": 0x1c,
	"et": 0x1d,
	"sa": 0x1e,
	"em": 0x1f,
	"ro": 0x20,
}

var trigram = map[string]byte{
	"the": 0x21,
	"and": 0x22,
	"ing": 0x23,
	"her": 0x24,
	"hat": 0x25,
	"his": 0x26,
	"tha": 0x27,
	"ere": 0x28,
	"for": 0x29,
	"ent": 0x2a,
	"ion": 0x2b,
	"ter": 0x2c,
	"was": 0x2d,
	"you": 0x2e,
	"ith": 0x2f,
	"ver": 0x30,
	"all": 0x31,
	"wit": 0x32,
	"thi": 0x33,
	"tio": 0x34,
	"nde": 0x35,
	"has": 0x36,
	"nce": 0x37,
	"edt": 0x38,
	"tis": 0x39,
	"oft": 0x3a,
	"sth": 0x3b,
	"men": 0x3c,
}

var quadgram = map[string]byte{
	"that": 0x00,
	"ther": 0x00,
	"with": 0x00,
	"tion": 0x00,
	"here": 0x00,
	"ould": 0x00,
	"ight": 0x00,
	"have": 0x00,
	"hich": 0x00,
	"whic": 0x00,
	"this": 0x00,
	"thin": 0x00,
	"they": 0x00,
	"atio": 0x00,
	"ever": 0x00,
	"from": 0x00,
	"ough": 0x00,
	"were": 0x00,
	"hing": 0x00,
	"ment": 0x00,
}
