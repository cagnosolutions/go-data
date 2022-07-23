package io

import (
	"fmt"
	"strconv"
	"strings"
)

type fileNamer struct {
	prefix string
	suffix string
}

func newFileNamer(prefix, suffix string) *fileNamer {
	return &fileNamer{
		prefix: prefix,
		suffix: suffix,
	}
}

func (n *fileNamer) newName(seqID uint16) string {
	return fmt.Sprintf("seg-%.5d-%.4s%.3s", seqID, n.prefix, n.suffix)
}

func (n *fileNamer) getSeqID(name string) uint16 {
	i := strings.Index(name, "-")
	id, err := strconv.ParseUint(name[i+1:i+6], 10, 16)
	if err != nil {
		panic("error: could not get sequence id from name; " + err.Error())
	}
	return uint16(id)
}
