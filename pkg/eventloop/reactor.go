package eventloop

import (
	"context"
	"errors"
	"log"
	"time"
)

var ErrNext = errors.New("next handler")

type Reactor struct {
	d Dispatcher
}

func NewReactor(dispatcher Dispatcher) *Reactor {
	if dispatcher == nil {
		log.Panicln("NewReactor requires a dispatcher to be provided!")
	}
	return &Reactor{
		d: dispatcher,
	}
}

func (r *Reactor) Run(ticks, timeout time.Duration) {
	// do something with ticks and timeout eventually
	for {
		h, err := r.d.Accept()
		if err != nil {
			if err == ErrNext {
				h = h.Next()
				goto serve
			}
			log.Printf("serve err: %s\n", err)
		}
	serve:
		c := h.GetContext()
		err = h.Serve(c)
		if err != nil {
			log.Printf("serve err: %s\n", err)
		}
	}
}

type Dispatcher interface {
	Accept() (Handler, error)
}

type Handler interface {
	GetContext() context.Context
	Serve(context.Context) error
	Next() Handler
}
