package main

import (
	"fmt"
	"io"
	"log"

	"github.com/cagnosolutions/go-data/pkg/tcp/experimental"
	v1 "github.com/cagnosolutions/go-data/pkg/tcp/v1"
)

func main() {

	srv := v1.NewServer("tcp", "0.0.0.0:7000")
	log.Panicln(srv.ListenAndServe(nil))

}

func example1() {
	// create handlers and serve without setting
	// up a server explicitly
	experimental.Handle("foo", exampleHandler{})
	experimental.HandleFunc("bar", exampleHandlerFunc)
	log.Panic(experimental.ListenAndServe(":4321", nil))
}

func example2() {
	// create handlers and serve, setting
	// up and using a server explicitly
	srv := &experimental.Server{
		Addr: ":4321",
		// Other options can be set here
	}
	srv.Handle("foo", exampleHandler{})
	srv.HandleFunc("bar", exampleHandlerFunc)
	log.Panic(srv.ListenAndServe())
}

type exampleHandler struct{}

func (eh exampleHandler) Serve(w experimental.ResponseWriter, r *experimental.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "got an error: %s\n", err)
		return
	}
	fmt.Fprintf(w, "got some data: %s\n", data)
	return
}

func exampleHandlerFunc(w experimental.ResponseWriter, r *experimental.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "got an error: %s\n", err)
		return
	}
	fmt.Fprintf(w, "got some data: %s\n", data)
	return
}
