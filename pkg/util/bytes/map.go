package bytes

// ref: [https://cs.opensource.google/go/go/+/refs/tags/go1.19.1:src/bytes/bytes.go;l=559]

func Map(mapping func(b byte) (byte, bool), s []byte) []byte {
	max, off := len(s), 0
	var keep bool
	b := make([]byte, max)
	for i := 0; i < len(s); i++ {
		r := s[i]
		r, keep = mapping(r)
		if keep {
			if off > max {
				// Grow the buffer.
				max = max * 2
				nb := make([]byte, max)
				copy(nb, b[0:off])
				b = nb
			}
			b[off] = r
			off++
		}
	}
	return b[0:off]
}
