package tcp

// A Handler responds to a TCP request.
//
// Serve should write data to the ResponseWriter and then
// return. Returning signals that the request is finished; it is
// not valid to use the ResponseWriter or read from the Request.Body
// after or concurrently with the completion of the
// Serve call.
//
// Depending on the TCP client software, and any intermediaries
// between the client and the Go server, it may not be possible to
// read from the Request.Body after writing to the ResponseWriter.
// Cautious handlers should read the Request.Body first, and then reply.
//
// Except for reading the body, handlers should not modify the
// provided Request.
//
// If Serve panics, the server (the caller of Serve) assumes that
// the effect of the panic was isolated to the active request.
// It recovers the panic, logs a stack trace to the server error log,
// and closes the network connection. To abort a handler so the client
// sees an interrupted response but the server doesn't log an error,
// panic with the value ErrAbortHandler.
type Handler interface {
	Serve(w ResponseWriter, r *Request)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as TCP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(w ResponseWriter, r *Request)

// Serve calls f(w, r).
func (f HandlerFunc) Serve(w ResponseWriter, r *Request) {
	f(w, r)
}

type serverHandler struct {
	server *Server
}

func (sh serverHandler) Serve(w ResponseWriter, r *Request) {
	sh.server.Handler.Serve(w, r)
}
