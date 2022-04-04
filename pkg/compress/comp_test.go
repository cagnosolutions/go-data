package compress

import (
	"fmt"
	"testing"
)

func TestPrefix(t *testing.T) {
	for category, items := range samples {
		for _, item := range items {
			prefix := Prefix([]byte(item))
			fmt.Printf("category=%q, prefix=%q, item=%q\n", category, prefix, item)
		}
	}
}

func TestPrefixSize(t *testing.T) {
	for category, items := range samples {
		for _, item := range items {
			prefix := PrefixSize([]byte(item), 4)
			fmt.Printf("category=%q, prefix(4)=%q, item=%q\n", category, prefix, item)
		}
	}
}

func TestSuffix(t *testing.T) {
	for category, items := range samples {
		for _, item := range items {
			suffix := Suffix([]byte(item))
			fmt.Printf("category=%q, suffix=%q, item=%q\n", category, suffix, item)
		}
	}
}

func TestSuffixSize(t *testing.T) {
	for category, items := range samples {
		for _, item := range items {
			suffix := SuffixSize([]byte(item), 4)
			fmt.Printf("category=%q, suffix(4)=%q, item=%q\n", category, suffix, item)
		}
	}
}

func TestAffix(t *testing.T) {
	for category, items := range samples {
		for _, item := range items {
			affix := Affix([]byte(item))
			fmt.Printf("category=%q, affix=%q, item=%q\n", category, affix, item)
		}
	}
}

func TestAffixSize(t *testing.T) {
	for category, items := range samples {
		for _, item := range items {
			affix := AffixSize2([]byte(item), 4)
			fmt.Printf("category=%q, affix(4)=%q, item=%q\n", category, affix, item)
		}
	}
}
