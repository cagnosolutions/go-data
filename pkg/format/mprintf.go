package format

import (
	"fmt"
	"os"
)

func Mprintf(s string, m map[string]interface{}) string {
	return os.Expand(
		s, func(k string) string {
			v, ok := m[k]
			if !ok {
				return ""
			}
			switch v.(type) {
			case string, []byte:
				return fmt.Sprintf("'%s'", v)
			default:
				return fmt.Sprintf("%v", v)
			}
		},
	)
}

func _Format(s string, m map[string]string) string {
	return os.Expand(s, func(k string) string { return m[k] })
}

func _FormatArgs(s string, kvs ...string) string {
	return os.Expand(
		s, func(k string) string {
			for i := 1; i < len(kvs); i++ {
				if kvs[i-1] == k {
					return kvs[i]
				}
			}
			return ""
		},
	)
}
