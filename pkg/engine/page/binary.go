package page

import (
	"encoding/binary"
)

// Binary encoding and decoding helpers
var (
	encU16 = binary.LittleEndian.PutUint16
	encU32 = binary.LittleEndian.PutUint32
	decU16 = binary.LittleEndian.Uint16
	decU32 = binary.LittleEndian.Uint32
)
