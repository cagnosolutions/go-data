package tcp

import (
	"bufio"
	"context"
	"fmt"
	"net"
)

type conn struct {
	server *Server
	rwc    *net.TCPConn
	bufr   *bufio.Reader
	bufw   *bufio.Writer
}

func (c *conn) serve(ctx context.Context) {

	ctx = context.WithValue(ctx, LocalAddrCtxKey, c.rwc.LocalAddr())

	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	c.bufr = newBufioReader(c.rwc)
	c.bufw = newBufioWriterSize(c.rwc, 4<<10)

	for {
		w, err := c.readRequest(ctx)
		if err != nil {
			fmt.Fprintf(c.rwc, "bad request: %s\n", err)
			return
		}
		serverHandler{c.server}.Serve(w, w.req)
	}

}

func (c *conn) readRequest(ctx context.Context) (*Response, error) {
	data, err := c.bufr.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	r := new(Request)
	err = r.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}
	return &Response{req: r}, nil
}

/*
func handleTCPConn(c *net.TCPConn, handler Handler) {

	// Create buffered io handlers
	r := bufio.NewReader(c)
	// w := bufio.NewWriter(c)
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
