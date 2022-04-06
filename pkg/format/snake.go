package format

// ToSnakeCase takes a string such as "SnakeCase" and returns the string
// formatted in the snake case format, ie: "snake_case"
func ToSnakeCase(s string) string {
	var b []byte
	for i := range s {
		if isAlphaUpper(s[i]) {
			if i > 0 && isAlphaLower(s[i-1]) {
				b = append(b, '_')
			}
			b = append(b, upperToLower(s[i]))
			continue
		}
		b = append(b, s[i])
	}
	return string(b)
}
