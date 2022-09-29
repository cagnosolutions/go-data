package util

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

var DefaultTracer = NewTracer(3)

type Tracer struct {
	Depth int
}

type Model interface {
	Name() string
}

type Repository[M Model] struct {
	model M
	name  string
}

func NewRepository[M Model](m M) *Repository[M] {
	return &Repository[M]{
		model: m,
		name:  m.Name(),
	}
}

func NewTracer(depth int) *Tracer {
	return &Tracer{
		Depth: depth,
	}
}

func (t *Tracer) String() string {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(t.Depth, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	sfile := strings.Split(file, "/")
	sname := strings.Split(f.Name(), "/")
	return fmt.Sprintf("[%s:%d %s]", sfile[len(sfile)-1], line, sname[len(sname)-1])
}

func (t *Tracer) Panic(err error) {
	log.Panicf("%s (errType=%T, err=%q)", t, err, err)
}

func (t *Tracer) Log(err error) {
	log.Printf("%s (errType=%T, err=%q)", t, err, err)
}

func (t *Tracer) Logln(err error) {
	log.Printf("%s (errType=%T, err=%q)\n", t, err, err)
}

func (t *Tracer) Logf(format string, v ...interface{}) {
	log.Printf(t.String()+" "+format, v...)
}
