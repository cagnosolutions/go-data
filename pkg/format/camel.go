package format

import (
	"strings"
)

func ToCamelCase(s string) string {
	// "fast path" checks
	if s == "" {
		return s
	}
	if len(s) == 1 {
		if IsAlphaLower(s[0]) {
			return string(s[0] - Lower)
		}
		return s
	}
	// okay, so we are in it now
	var upperNext bool
	var dist strings.Builder
	// avoid reallocation memory, 33% ~ 50% is recommended
	dist.Grow(len(s))
	// start looping
	for i := 0; i < len(s); i++ {
		// if current character is _ and we are not at the first character
		if s[i] == '_' && i > 0 {
			upperNext = true
			continue
		}
		// if current character is lowercase and upperNext = true
		if IsAlphaLower(s[i]) && upperNext {
			upperNext = false
			dist.WriteByte(s[i] - Lower)
			continue
		}
		// if current character is the first letter
		if i == 0 {
			// and if it is _, write it and set upperNext=true
			if s[i] == '_' {
				dist.WriteByte(s[i])
				upperNext = true
				continue
			}
			// otherwise, just write upper
			dist.WriteByte(s[i] - Lower)
			continue
		}
		// all other cases, just write lowercase
		dist.WriteByte(s[i])
	}
	return dist.String()
}
