package tcp

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type contextKey struct {
	name string
}

var LocalAddrCtxKey = &contextKey{"local-addr"}
var ServerCtxKey = &contextKey{"tcp-server"}

type Server struct {
	Addr         string
	Handler      Handler
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	ErrorLog     *log.Logger

	// BaseContext optionally specifies a function that returns
	// the base context for incoming requests on this server.
	// The provided Listener is the specific Listener that's
	// about to start accepting requests.
	// If BaseContext is nil, the default is context.Background().
	// If non-nil, it must return a non-nil context.
	BaseContext func(net.Listener) context.Context

	// ConnContext optionally specifies a function that modifies
	// the context used for a new connection c. The provided ctx
	// is derived from the base context and has a ServerContextKey
	// value.
	ConnContext func(ctx context.Context, c net.Conn) context.Context

	listeners  map[*net.Listener]struct{}
	activeConn map[*conn]struct{}

	ln *net.TCPListener
}

type entry struct {
	h       Handler
	pattern string
}

// NewServer allocates and returns a new Server.
// func NewServer() *Server {
// 	return new(Server)
// }

// DefaultServer is the default Server used by Serve.
var DefaultServer = &defaultServer

var defaultServer Server

func (s *Server) Handle(pattern string, handler Handler) {

}

func (s *Server) HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {

}

var defaultHandler = HandlerFunc(
	func(w ResponseWriter, r *Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			Error(w, err)
			return
		}
		ct := time.Now()
		_, err = fmt.Fprintf(w, "[%s] received: %v\n", ct.Format(time.RFC3339Nano), data)
		if err != nil {
			Error(w, err)
			return
		}
		return
	},
)

// ListenAndServe listens on the TCP network address srv.Addr and then
// calls Serve to handle requests on incoming connections.
// Accepted connections are configured to enable TCP keep-alives.
//
// If srv.Addr is blank, ":http" is used.
//
// ListenAndServe always returns a non-nil error. After Shutdown or Close,
// the returned error is ErrServerClosed.
func (s *Server) ListenAndServe() error {
	host := s.Addr
	if host == "" {
		host = ":tcp"
	}
	// Resolve TCP address
	addr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		return err
	}
	// Initialize a listening socket
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	return s.Serve(ln)
}

func (s *Server) Serve(ln *net.TCPListener) error {
	// Check the io handler
	handler := s.Handler
	if handler == nil {
		handler = defaultHandler
	}

	defer ln.Close()

	// Establish contexts
	baseCtx := context.Background()
	if s.BaseContext != nil {
		baseCtx = s.BaseContext(s.ln)
		if baseCtx == nil {
			panic("BaseContext returned a nil context")
		}
	}

	ctx := context.WithValue(baseCtx, ServerCtxKey, s)

	// Wait for connections
	for {
		// Block for a connection
		rw, err := ln.AcceptTCP()
		if err != nil {
			// Get back to the accepting state pronto
			s.logf("accept conn: %v\n", err)
			continue
		}
		// We have a connection. We will hand it off to
		// the handler (in its own goroutine) and promptly
		// get back to the accepting state
		s.logf("accepted conn: %q\n", rw.RemoteAddr())

		connCtx := ctx
		if cc := s.ConnContext; cc != nil {
			connCtx = cc(connCtx, rw)
			if connCtx == nil {
				panic("ConnContext returned nil")
			}
		}

		c := s.newConn(rw)

		go c.serve(connCtx)
	}
}

func (s *Server) newConn(c *net.TCPConn) *conn {
	return &conn{
		server: s,
		rwc:    c,
	}
}

/*
func serve(c *net.TCPConn) {

	// Create buffered io handlers
	r := bufio.NewReader(c)
	w := bufio.NewWriter(c)
	// rw := bufio.NewReadWriter(r, w)

	// Don't forget to close!
	defer func(c *net.TCPConn) {
		err := c.Close()
		if err != nil {
			log.Printf("closing conn: %v\n", err)
		}
	}(c)

	for {
		// Read data from connection until we reach newline delim
		data, err := r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				// If we reach EOF before we find a newline delim
				// we might as well just stop
				break
			}
			// If it's not an EOF error, just log it and return
			log.Printf("error reading from connection: %s\n", err)
			return
		}
		// Otherwise, we are good. Let's handle the request and
		// response using the supplied TCPHandler function.
		err = handler(bytes.NewBuffer(data), c)
		if err != nil {
			log.Printf("error from io handler: %s\n", err)
			break
		}
	}
}
*/

func (s *Server) Close() error {
	err := s.ln.Close()
	if err != nil {
		return err
	}
	return nil
}

// Handle registers the handler for the given pattern in the DefaultServer.
func Handle(pattern string, handler Handler) {
	DefaultServer.Handle(pattern, handler)
}

// HandleFunc registers the handler function for the given pattern
// in the DefaultServer.
func HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
	DefaultServer.HandleFunc(pattern, handler)
}

func ListenAndServe(addr string, handler Handler) error {
	server := &Server{Addr: addr, Handler: handler}
	return server.ListenAndServe()
}

func (s *Server) logf(format string, args ...any) {
	if s.ErrorLog != nil {
		s.ErrorLog.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}
