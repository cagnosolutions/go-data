package v1

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

type ConnectionHandler = func(conn *net.TCPConn)

type Server struct {
	Network string
	Address *net.TCPAddr
	*log.Logger
}

func NewServer(network, address string) *Server {
	// setup binding address
	addr, err := net.ResolveTCPAddr(network, address)
	if err != nil {
		panic(err)
		return nil
	}
	// return new server instance
	return &Server{
		Network: network,
		Address: addr,
		Logger:  log.New(os.Stdout, "[server] ", log.LstdFlags),
	}
}

func (s *Server) ListenAndServe(handler ConnectionHandler) error {
	// compute handler
	if handler == nil {
		handler = s.defaultConnectionHandler
	}
	// use binding address to establish a listening socket
	ln, err := net.ListenTCP(s.Network, s.Address)
	if err != nil {
		return err
	}
	// defer the listening socket's close
	defer func(ln *net.TCPListener) {
		err := ln.Close()
		if err != nil {
			panic(err)
		}
	}(ln)
	s.tagf("listening", "on %s accepting connections\n", s.Address.String())
	// start accept loop
	for {
		// accept incoming connections
		conn, err := ln.AcceptTCP()
		if err != nil {
			s.logf("error accepting connection: %s\n", err)
			continue
		}
		// handle client connection
		go handler(conn)
	}
}

func (s *Server) tagf(tag, format string, args ...any) {
	s.Logger.Printf("[%s] %s", tag, fmt.Sprintf(format, args))
}

func (s *Server) logf(format string, args ...any) {
	s.Logger.Printf(format, args...)
}

func (s *Server) defaultConnectionHandler(conn *net.TCPConn) {
	// accepted new connection
	s.tagf("accepted", "connection from %s", conn.RemoteAddr())
	// setup read and write buffer
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	for {
		// time out one minute from now if no data is received.
		_ = conn.SetReadDeadline(time.Now().Add(time.Minute))
		// Read and process data from client
		data, err := rw.ReadBytes('\n')
		if err != nil {
			// if err == io.EOF {
			// 	break
			// }
			// panic(err)
			break
		}
		// check for break
		if len(data) == 1 && data[0] == '\n' {
			rw.Discard(len(data))
			rw = nil
			break
		}
		// Write data back to the client
		_, err = rw.Write(data)
		if err != nil {
			panic(err)
		}
		// Flush and reset
		err = rw.Flush()
		if err != nil {
			panic(err)
		}
	}
	// close the connection
	err := conn.Close()
	if err != nil {
		panic(err)
	}
	// closed client connection
	s.tagf("closed", "connection from %s", conn.RemoteAddr())
}
