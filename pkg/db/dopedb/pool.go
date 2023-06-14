package dopedb

import (
	"strings"
	"sync"
)

var bytesPool = &_bytpool{}

type _bytpool struct {
	pool sync.Pool
}

func (p *_bytpool) Get() *[]byte {
	if p.pool.New == nil {
		p.pool = sync.Pool{
			New: func() any {
				b := make([]byte, 4096)
				return &b
			},
		}
	}
	return p.pool.Get().(*[]byte)
}

func resetBytes(v *[]byte) {
	for i := range *v {
		(*v)[i] = 0x00
	}
}

func (p *_bytpool) Put(v *[]byte) {
	resetBytes(v)
	p.pool.Put(v)
}

var bytesBool = sync.Pool{
	New: func() any {
		b := make([]byte, 8192)
		return &b
	},
}

var stringPool = &_strpool{}

type _strpool struct {
	pool sync.Pool
}

func (p *_strpool) Get() *strings.Builder {
	if p.pool.New == nil {
		p.pool = sync.Pool{
			New: func() any {
				return new(strings.Builder)
			},
		}
	}
	return p.pool.Get().(*strings.Builder)
}

func (p *_strpool) Put(v *strings.Builder) {
	v.Reset()
	p.pool.Put(v)
}

func readString(p []byte) string {
	sb := stringPool.Get()
	sb.Grow(len(p))
	sb.Write(p)
	s := sb.String()
	stringPool.Put(sb)
	return s
}

func readBytes(p []byte) []byte {
	b := make([]byte, len(p))
	copy(b, p)
	return b
}
