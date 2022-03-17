package scanner

type jsonType byte
type JSONType jsonType

const (
	TypeError jsonType = iota
	TypeObject
	TypeArray
	TypeString
	TypeNumber
	TypeBoolean
	TypeNull
	TypeUnknown
)

// TypeOf returns the type of the object at the current position or TypeUnknown if it cannot be determined.
func TypeOf(in []byte, pos int) (jsonType, error) {
	// skip any whitespace first
	pos, err := SkipSpace(in, pos)
	if err != nil {
		return TypeError, err
	}
	switch in[pos] {
	case '{':
		return TypeObject, nil
	case '[':
		return TypeArray, nil
	case '"':
		return TypeString, nil
	case '.', '-', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
		return TypeNumber, nil
	case 't', 'f':
		return TypeBoolean, nil
	case 'n':
		return TypeNull, nil
	default:
		max := len(in) - pos
		if max > 20 {
			max = 20
		}
		return TypeError, OpErr{
			Pos:     pos,
			Msg:     "invalid object",
			Content: string(in[pos : pos+max]),
		}
	}
}
