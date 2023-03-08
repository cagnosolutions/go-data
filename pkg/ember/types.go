package ember

const (
	_ byte = iota
	typString
	typBytes
	typInteger
	typFloat
	typBitset
	typList
	typSortedList
	typMap

	encInt8
	encInt16
	encInt32
	encInt64
	encUint8
	encUint16
	encUint32
	encUint64
	encFloat32
	encFloat64
	encString
	encBytes
)

type value struct {
	typ uint8
}

type zl struct {
	zlbytes uint32 // total bytes the ziplist contains
	zltail  uint32 // offset to the last entry
	zllen   uint16 // number of entries
	entries []zlentry
	zlend   uint16 // single byte (0xffff) indicating the end of the ziplist
}

type zlentry struct {
	prevlen  uint16 // previous entry length (for backwards traversal)
	encoding uint16
	data     []byte
}

type ziplist []byte

func (zl *ziplist) setZipListHeader(total uint32, lastoff uint32, nument uint16) {
	// set total bytes
	bin.PutUint32((*zl)[0:4], total)
	// set offset to last entry
	bin.PutUint32((*zl)[4:8], lastoff)
	// set number of entries
	bin.PutUint16((*zl)[8:10], nument)
}

func newZiplist() ziplist {
	b := make(ziplist, 12)
	b.setZipListHeader(12, 10, 0)
	return b
}

func makeZLEntry(prevlen uint16, encoding uint16, data any) []byte {
	b := make([]byte, 4+len(data))
	// encode the length to the previous entry
	bin.PutUint16(b[0:2], prevlen)
	// encode the encoding
	bin.PutUint16(b[2:4], encoding)
	// encode data
	copy(b[4:], data)
	return b
}

func getEncoding(v any) (byte, int) {
	switch t := v.(type) {
	case int8:
		return encUint8, 1
	case int16:
		return encInt16, 2
	case int32:
		return encInt32, 4
	case int64:
		return encInt64, 8
	case uint8:
		return encUint8, 1
	case uint16:
		return encUint16, 2
	case uint32:
		return encUint32, 4
	case uint64:
		return encUint64, 8
	case float32:
		return encFloat32, 4
	case float64:
		return encFloat64, 8
	case string:
		return encString, len(t)
	case []byte:
		return encBytes, len(t)
	}
	return 0x00, 0
}

func (zl *ziplist) rpush(ss ...any) {
	// iterate our values and add them, updating the header
	var lastoff uint32
	var enc byte
	var size int
	for _, s := range ss {
		// get our encoding type and size
		enc, size = getEncoding(s)
		// grow our ziplist
		*zl = append(*zl, make([]byte, 4+size)...)
		// get the offset of the last entry
		lastoff = bin.Uint32((*zl)[4:8])
		// encode new entry
		entry := makeZLEntry(lastoff, enc, s)
	}

}
