package main

import (
	"fmt"
	"io"
	"log"

	"github.com/cagnosolutions/go-data/pkg/tcp"
)

func main() {
}

func example1() {
	// create handlers and serve without setting
	// up a server explicitly
	tcp.Handle("foo", exampleHandler{})
	tcp.HandleFunc("bar", exampleHandlerFunc)
	log.Panic(tcp.ListenAndServe(":4321", nil))
}

func example2() {
	// create handlers and serve, setting
	// up and using a server explicitly
	srv := &tcp.Server{
		Addr: ":4321",
		// Other options can be set here
	}
	srv.Handle("foo", exampleHandler{})
	srv.HandleFunc("bar", exampleHandlerFunc)
	log.Panic(srv.ListenAndServe())
}

type exampleHandler struct{}

func (eh exampleHandler) Serve(w tcp.ResponseWriter, r *tcp.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "got an error: %s\n", err)
		return
	}
	fmt.Fprintf(w, "got some data: %s\n", data)
	return
}

func exampleHandlerFunc(w tcp.ResponseWriter, r *tcp.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "got an error: %s\n", err)
		return
	}
	fmt.Fprintf(w, "got some data: %s\n", data)
	return
}
