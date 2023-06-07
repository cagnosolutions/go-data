package ember

import (
	"encoding/binary"
	"io"
)

// bin is the binary byte order for this package.
// It just makes it easier to set in one place.
var bin = binary.BigEndian
var binw = func(w io.Writer, data any) error {
	return binary.Write(w, binary.BigEndian, data)
}
