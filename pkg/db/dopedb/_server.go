package dopedb

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

type DBServer struct {
	db      *DB
	addr    string
	timeout time.Duration
}

func NewDBServer(db *DB, timeout time.Duration) *DBServer {
	return &DBServer{
		db:      db,
		timeout: timeout,
	}
}

func (s *DBServer) ListenAndServe(host string) error {
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
	// Don't forget to close!
	defer func(ln *net.TCPListener) {
		err := ln.Close()
		if err != nil {
			log.Panicf("closing: %v", err)
		}
	}(ln)
	// Wait for connections
	for {
		// Block for a connection
		conn, err := ln.AcceptTCP()
		if err != nil {
			// Get back to the accepting state pronto
			log.Printf("accept conn: %v\n", err)
			continue
		}
		// We have a connection. We will set the deadline
		// right away and update the deadline with every
		// successful transaction in the handler, thus creating
		// an idle timeout.
		err = conn.SetDeadline(time.Now().Add(s.timeout))
		if err != nil {
			log.Printf("set deadline: %v\n", err)
		}
		// And now, we will hand it end to the handler (in its own
		// goroutine) and promptly get back to the accepting state
		log.Printf("accepted conn: %q\n", conn.RemoteAddr())
		go s.handleTCPConn(conn)
	}
}

func (s *DBServer) handleTCPConn(c *net.TCPConn) {

	// Create buffered io handlers
	r := bufio.NewReader(c)
	w := bufio.NewWriter(c)

	// Don't forget to close!
	defer func(c *net.TCPConn) {
		err := c.Close()
		if err != nil {
			log.Printf("closing conn: %v\n", err)
		}
	}(c)
	var badReq int
	for {
		// Read data from connection until we reach newline delim
		req, err := r.ReadBytes('\n')
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
		// Otherwise, we are good. Let's handle the request...
		resp, err := s.db.Exec(req)
		if err != nil {
			_, err = w.Write([]byte(fmt.Sprintf("error: %s\n", err)))
			if err != nil {
				log.Printf("writing to conn: %v\n", err)
			}
			if badReq > 5 {
				break
			}
			badReq++
		}
		// As long as the request was good, we can respond
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("writing to conn: %v\n", err)
		}
		// Update the deadline
		err = c.SetDeadline(time.Now().Add(s.timeout))
		if err != nil {
			log.Printf("set deadline: %v\n", err)
		}
	}
	_, err := w.Write([]byte("connection closed.\n"))
	if err != nil {
		log.Printf("writing to conn: %v\n", err)
	}
	log.Printf("closing conn: %v\n", err)
}

func handleConnClose(c *net.TCPConn) {
	_, err := c.Write([]byte("connection closed.\n"))
	if err != nil {
		log.Printf("writing to conn: %v\n", err)
	}
	err = c.Close()
	if err != nil {
		log.Printf("closing conn: %v\n", err)
	}
	c = nil
	return
}

func handleConnRead(c *net.TCPConn) (int, []string) {
	data, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		log.Printf("conn read: %v\n", err)
		return -1, nil
	}
	fields := strings.Fields(data)
	return len(fields), fields
}

func handleConnWrite(c *net.TCPConn, format string, args ...any) {
	if format[len(format)-1] != '\n' {
		format += "\n"
	}
	_, err := fmt.Fprintf(c, format, args...)
	if err != nil {
		log.Printf("conn write: %v\n", err)
	}
}
