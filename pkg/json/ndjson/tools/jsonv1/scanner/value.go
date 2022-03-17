package scanner

// Value returns the position of the end of the current element that begins at pos; handles any valid json element
func Value(in []byte, pos int) (jsonType, int, error) {
	pos, err := SkipSpace(in, pos)
	if err != nil {
		return TypeError, 0, err
	}
	switch in[pos] {
	case '{':
		pos, err := Object(in, pos)
		return TypeObject, pos, err
	case '[':
		pos, err := Array(in, pos)
		return TypeArray, pos, err
	case '"':
		pos, err := String(in, pos)
		return TypeString, pos, err
	case '.', '-', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
		pos, err := Number(in, pos)
		return TypeNumber, pos, err
	case 't', 'f':
		pos, err := Boolean(in, pos)
		return TypeBoolean, pos, err
	case 'n':
		pos, err := Null(in, pos)
		return TypeNull, pos, err
	default:
		max := len(in) - pos
		if max > 20 {
			max = 20
		}
		return TypeError, 0, OpErr{
			Pos:     pos,
			Msg:     "invalid object",
			Content: string(in[pos : pos+max]),
		}
	}
}
