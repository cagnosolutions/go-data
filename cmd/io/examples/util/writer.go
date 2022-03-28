package util

import (
	"bytes"
)

type OutWriter struct {
	w *bytes.Buffer
}

func NewOutWriter() *OutWriter {
	return &OutWriter{
		w: new(bytes.Buffer),
	}
}

func (o *OutWriter) Write(p []byte) (int, error) {
	return o.w.Write(p)
}

func (o *OutWriter) Bytes() []byte {
	return o.w.Bytes()
}
