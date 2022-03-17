package scanner

// FindIndex accepts a JSON array and return the value of the element at the specified index
func FindIndex(in []byte, pos, index int) ([]byte, error) {
	pos, err := SkipSpace(in, pos)
	if err != nil {
		return nil, err
	}

	if v := in[pos]; v != '[' {
		return nil, NewError(pos, v)
	}
	pos++

	idx := 0
	for {
		pos, err = SkipSpace(in, pos)
		if err != nil {
			return nil, err
		}

		itemStart := pos
		// data
		pos, err = Any(in, pos)
		if err != nil {
			return nil, err
		}
		if index == idx {
			return in[itemStart:pos], nil
		}

		pos, err = SkipSpace(in, pos)
		if err != nil {
			return nil, err
		}

		switch in[pos] {
		case ',':
			pos++
		case ']':
			return nil, ErrIndexOutOfBounds
		}

		idx++
	}
}
