package page

import (
	"encoding/binary"
)

// Binary encoding and decoding helpers
var (
	encU16 = binary.LittleEndian.PutUint16
	encU32 = binary.LittleEndian.PutUint32
	encU64 = binary.LittleEndian.PutUint64
	decU16 = binary.LittleEndian.Uint16
	decU32 = binary.LittleEndian.Uint32
	decU64 = binary.LittleEndian.Uint64
)
