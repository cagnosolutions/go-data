package scanner

// AsArray accepts an []byte encoded json array as an input and returns the array's elements
func AsArray(in []byte, pos int) ([][]byte, error) {
	pos, err := SkipSpace(in, pos)
	if err != nil {
		return nil, err
	}

	if v := in[pos]; v != '[' {
		return nil, NewError(pos, v)
	}
	pos++

	// clean initial spaces
	pos, err = SkipSpace(in, pos)
	if err != nil {
		return nil, err
	}

	if in[pos] == ']' {
		return [][]byte{}, nil
	}

	// 1. Count the number of elements in the array

	start := pos

	elements := make([][]byte, 0, 256)
	for {
		pos, err = SkipSpace(in, pos)
		if err != nil {
			return nil, err
		}

		start = pos

		// data
		pos, err = Any(in, pos)
		if err != nil {
			return nil, err
		}
		elements = append(elements, in[start:pos])

		pos, err = SkipSpace(in, pos)
		if err != nil {
			return nil, err
		}

		switch in[pos] {
		case ',':
			pos++
		case ']':
			return elements, nil
		}
	}
}
