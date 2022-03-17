package scanner

// Array returns the position of the end of the array that begins at the position specified
func Array(in []byte, pos int) (int, error) {
	pos, err := SkipSpace(in, pos)
	if err != nil {
		return 0, err
	}

	if v := in[pos]; v != '[' {
		return 0, NewError(pos, v)
	}
	pos++

	// clean initial spaces
	pos, err = SkipSpace(in, pos)
	if err != nil {
		return 0, err
	}

	if in[pos] == ']' {
		return pos + 1, nil
	}

	for {
		// data
		pos, err = Any(in, pos)
		if err != nil {
			return 0, err
		}

		pos, err = SkipSpace(in, pos)
		if err != nil {
			return 0, err
		}

		switch in[pos] {
		case ',':
			pos++
		case ']':
			return pos + 1, nil
		}
	}
}
