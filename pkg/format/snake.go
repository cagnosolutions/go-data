package format

import (
	"log"
	"strings"
	"unicode"
)

const Lower byte = 32

func out(mark int, c byte) {
	log.Printf("marker=%d, ch=%c\n", mark, c)
}

func ToSnakeCase(s string) string {
	// "fast path" checks
	if s == "" {
		return s
	}
	if len(s) == 1 {
		return strings.ToLower(s)
	}
	// okay, so we are in it now
	var skipNext bool
	var dist strings.Builder
	// avoid reallocation memory, 33% ~ 50% is recommended
	dist.Grow(len(s) + len(s)/3)
	// start looping
	for i := 0; i < len(s); i++ {
		// if current character is - or _
		if s[i] == '-' || s[i] == '_' {
			dist.WriteByte('_')
			skipNext = true
			continue
		}
		// if current character is lowercase or a number
		if IsAlphaLower(s[i]) || IsNumeric(s[i]) {
			dist.WriteByte(s[i])
			continue
		}
		// if current character is the first letter
		if i == 0 {
			dist.WriteByte(s[i] + Lower)
			continue
		}
		// if previous character is punctuation or lowercase
		if IsPunctuation(s[i-1]) || IsAlphaLower(s[i-1]) {
			if skipNext {
				skipNext = false
				dist.WriteByte(s[i] + Lower)
				continue
			}
			dist.WriteByte('_')
			dist.WriteByte(s[i] + Lower)
			continue
		}
		// if not at the end and next character is lowercase
		if i < len(s)-1 && IsAlphaLower(s[i+1]) {
			if skipNext {
				skipNext = false
				dist.WriteByte(s[i] + Lower)
				continue
			}
			dist.WriteByte('_')
			dist.WriteByte(s[i] + Lower)
			continue
		}
		// all other cases, just write lowercase
		dist.WriteByte(s[i] + Lower)
	}
	return dist.String()
}

func _ToSnakeCaseCopy(s string) string {
	if s == "" {
		return s
	}
	if len(s) == 1 {
		return strings.ToLower(s)
	}
	source := []rune(s)
	dist := strings.Builder{}
	dist.Grow(len(s) + len(s)/3) // avoid reallocation memory, 33% ~ 50% is recommended
	skipNext := false
	for i := 0; i < len(source); i++ {
		cur := source[i]
		switch cur {
		case '-', '_':
			dist.WriteRune('_')
			skipNext = true
			continue
		}
		if unicode.IsLower(cur) || unicode.IsDigit(cur) {
			dist.WriteRune(cur)
			continue
		}

		if i == 0 {
			dist.WriteRune(unicode.ToLower(cur))
			continue
		}

		last := source[i-1]
		if (!unicode.IsLetter(last)) || unicode.IsLower(last) {
			if skipNext {
				skipNext = false
			} else {
				dist.WriteRune('_')
			}
			dist.WriteRune(unicode.ToLower(cur))
			continue
		}
		// last is upper case
		if i < len(source)-1 {
			next := source[i+1]
			if unicode.IsLower(next) {
				if skipNext {
					skipNext = false
				} else {
					dist.WriteRune('_')
				}
				dist.WriteRune(unicode.ToLower(cur))
				continue
			}
		}
		dist.WriteRune(unicode.ToLower(cur))
	}

	return dist.String()
}
