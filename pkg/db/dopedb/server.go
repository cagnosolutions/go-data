package dopedb

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type NetHandler func(*net.TCPConn)

type IOHandler func(in *bytes.Buffer, out io.Writer) error

func EchoNetHandler() NetHandler {
	return func(c *net.TCPConn) {
		defer func(c *net.TCPConn) {
			err := c.Close()
			if err != nil {
				panic(err)
			}
		}(c)
		_, err := io.Copy(c, c)
		if err != nil {
			return
		}
	}
}

func EchoIOHandler() IOHandler {
	return func(in *bytes.Buffer, out io.Writer) error {
		_, err := io.Copy(out, in)
		if err != nil {
			return err
		}
		return nil
	}
}

var defaultIOHandler = EchoIOHandler()

// ListenAndServe creates a TCP listener using the provided
// host string and handles each connection concurrently using
// the provided NetHandler
func ListenAndServe(host string, handle IOHandler) error {
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
	// Check the io handler
	if handle == nil {
		handle = defaultIOHandler
	}
	// Wait for connections
	for {
		// Block for a connection
		conn, err := ln.AcceptTCP()
		if err != nil {
			// Get back to the accepting state pronto
			log.Printf("accept conn: %v\n", err)
			continue
		}
		// We have a connection. We will hand it off to
		// the handler (in its own goroutine) and promptly
		// get back to the accepting state
		log.Printf("accepted conn: %q\n", conn.RemoteAddr())
		go handleTCPConn(conn, handle)
	}
}

func handleTCPConn(c *net.TCPConn, handler IOHandler) {

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

func handleConn(c *net.TCPConn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	for {
		argc, argv := handleConnRead(c)
		// data, err := bufio.NewReader(c).ReadString('\n')
		// if err != nil {
		// 	log.Printf("reading from client: %v\n", err)
		// 	return
		// }
		// fields := strings.Fields(data)

		// check for client disconnect
		if argc == 1 && argv[0] == "STOP" {
			break
		}

		// log line
		log.SetPrefix("[SERVER]")
		log.Printf("Received %d arguments: %q\n", argc, argv)

		// write to client
		handleConnWrite(c, "echo: %v\n", argv)
	}
	// disconnect
	handleConnWrite(c, "Goodbye!\n")
	err := c.Close()
	if err != nil {
		log.Panicf("closing connection: %v\n", err)
	}
}
