package scanner

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	reArray = regexp.MustCompile(`^\s*\[\s*(\d+)(\s*:\s*(\d+))?\s*]\s*$`)
)

// Must is a convenience method similar to template.Must
func Must(op Op, err error) Op {
	if err != nil {
		panic(fmt.Errorf("unable to parse selector; %v", err.Error()))
	}

	return op
}

// Parse takes a string representation of a selector and returns the corresponding Op definition
func Parse(selector string) (Op, error) {
	segments := strings.Split(selector, ".")

	ops := make([]Op, 0, len(segments))
	for _, segment := range segments {
		key := strings.TrimSpace(segment)
		if key == "" {
			continue
		}

		if op, ok := ParseArray(key); ok {
			ops = append(ops, op)
			continue
		}

		ops = append(ops, Dot(key))
	}

	return Chain(ops...), nil
}

func ParseArray(key string) (Op, bool) {
	match := reArray.FindAllStringSubmatch(key, -1)
	if len(match) != 1 {
		return nil, false
	}

	fromStr := match[0][1]
	from, err := strconv.Atoi(fromStr)
	if err != nil {
		return nil, false
	}

	toStr := match[0][3]
	if toStr == "" {
		return Index(from), true
	}

	to, err := strconv.Atoi(toStr)
	if err != nil {
		return nil, false
	}

	return Range(from, to), true
}
