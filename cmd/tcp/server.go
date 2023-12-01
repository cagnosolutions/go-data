package main

import (
	"log"
	"net"
)

type Server struct {
	listenAddr string
	ln         net.Listener
	quit       chan struct{}
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		ln:         nil,
		quit:       make(chan struct{}),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.ln = ln

	go s.acceptLoop()

	<-s.quit
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			log.Printf("accept error: %s", err)
			continue
		}

		go s.readLoop(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("read error: %s", err)
			continue
		}
		_, err = conn.Write(buf[:n])
		if err != nil {
			log.Printf("write error: %s", err)
			continue
		}
	}
}
