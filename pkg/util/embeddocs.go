package util

import _ "embed"

//go:embed ww.txt
var WaltWhitmanTextString string

//go:embed ww.txt
var WaltWhitmanTextBytes []byte

//go:embed shakespeare-sonnets.txt
var ShakespearSonnetsString string

//go:embed shakespeare-sonnets.txt
var ShakespearSonnetsBytes []byte

//go:embed long-lines.txt
var LongLinesTextString string

//go:embed long-lines.txt
var LongLinesTextBytes []byte
