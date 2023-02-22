package spanindexer

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
)

type Segment struct {
	Beg    int
	End    int
	Status int
}

type seg struct {
	beg    int
	end    int
	status int // 1 = there, 0 = removed
}

type index map[int]*seg

type IndexFunc func(delim byte) bool

type Indexer struct {
	mu    sync.Mutex
	ix    index
	delim byte
}

func NewIndexer(delim byte) *Indexer {
	return &Indexer{
		ix:    make(index),
		delim: delim,
	}
}

func (i *Indexer) Index(b []byte) {
	i.mu.Lock()
	defer i.mu.Unlock()
	br := bufio.NewReaderSize(bytes.NewReader(b), 64)
	var part, off int
	for {
		p, err := br.ReadBytes(i.delim)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("reading line: %s", err)
		}
		i.ix[part] = &seg{beg: off, end: off + len(p), status: 1}
		part++
		off += len(p)
	}
}

func (i *Indexer) GetSegment(n int) *Segment {
	i.mu.Lock()
	defer i.mu.Unlock()
	v, found := i.ix[n]
	if !found {
		return nil
	}
	if v.status == 0 {
		return nil
	}
	return &Segment{v.beg, v.end, v.status}
}

func (i *Indexer) PruneSegment(n int) {
	i.mu.Lock()
	defer i.mu.Unlock()
	v, found := i.ix[n]
	if !found {
		return
	}
	v.status = 0
}

var SkipSpan = errors.New("skip span")

func (i *Indexer) Range(fn func(span, status, beg, end int) error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	for j := 0; j < len(i.ix); j++ {
		v := i.ix[j]
		err := fn(j, v.status, v.beg, v.end)
		if err != nil {
			if err == SkipSpan {
				continue
			}
			break
		}
	}
}

func (i *Indexer) Len() int {
	i.mu.Lock()
	defer i.mu.Unlock()
	return len(i.ix)
}

func (i *Indexer) String() string {
	i.mu.Lock()
	defer i.mu.Unlock()
	var ss string
	for j := 0; j < len(i.ix); j++ {
		v := i.ix[j]
		ss += fmt.Sprintf("span=%d: {beg:%d, end:%d, status:%d}\n", j, v.beg, v.end, v.status)
	}
	return ss
}
