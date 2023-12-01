package tunnel

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

type Endpoint struct {
	laddr  *net.TCPAddr
	raddr  *net.TCPAddr
	remote *net.TCPConn
	logf   *log.Logger
}

func NewEndpoint(name string, local, remote string) (*Endpoint, error) {
	laddr, err := net.ResolveTCPAddr("tcp", local)
	if err != nil {
		return nil, err
	}
	raddr, err := net.ResolveTCPAddr("tcp", remote)
	if err != nil {
		return nil, err
	}
	return &Endpoint{
		laddr: laddr,
		raddr: raddr,
		logf:  log.New(os.Stdout, fmt.Sprintf("[%s] ", name), log.LstdFlags),
	}, nil
}

func (e *Endpoint) out(format string, args ...any) {
	e.logf.Printf(format, args...)
}

func (e *Endpoint) Connect() error {

	// Open a listening socket on the local address
	ln, err := net.ListenTCP("tcp", e.laddr)
	if err != nil {
		return err
	}
	e.out("listening on %s (1 of 2)\n", e.laddr.String())

	// Connect to the remote address
	e.remote, err = net.DialTCP("tcp", e.laddr, e.raddr)
	if err != nil {
		return err
	}
	e.out("connected to %s (2 of 2)\n", e.raddr.String())

	for {
		con, err := ln.AcceptTCP()
		if err != nil {
			e.out("accept error: %s", err)
			continue
		}

		go e.handle(con)
	}

	return nil
}

func (e *Endpoint) handle(conn *net.TCPConn) {
	chan_to_local := stream_copy(e.remote, conn)
	chan_to_remote := stream_copy(conn, e.remote)
	select {
	case <-chan_to_local:
		log.Println("Remote connection closed")
	case <-chan_to_remote:
		log.Println("Local connection closed")
	}
}

func stream_copy(dst io.Writer, src io.Reader) <-chan int {
	buf := make([]byte, 1024)
	sync_channel := make(chan int)
	go func() {
		defer func() {
			if con, ok := dst.(net.Conn); ok {
				con.Close()
				log.Printf("Connection from %v is closed\n", con.RemoteAddr())
			}
			sync_channel <- 0 // Notify that processing is finished
		}()
		for {
			var nBytes int
			var err error
			nBytes, err = src.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("Read error: %s\n", err)
				}
				break
			}
			_, err = dst.Write(buf[0:nBytes])
			if err != nil {
				log.Fatalf("Write error: %s\n", err)
			}
		}
	}()
	return sync_channel
}

func pipe(w, r *net.TCPConn) {
	_, err := io.Copy(w, r)
	if err != nil {
		panic(err)
	}
}
