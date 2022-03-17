package scanner

// Number returns the end position of the number that begins at the specified pos
func Number(in []byte, pos int) (int, error) {
	pos, err := SkipSpace(in, pos)
	if err != nil {
		return 0, err
	}

	max := len(in)
	for {
		v := in[pos]
		switch v {
		case '-', '+', '.', 'e', 'E', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			pos++
		default:
			return pos, nil
		}
		if pos >= max {
			return pos, nil
		}
	}
	return pos, nil
}
