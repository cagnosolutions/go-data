package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/cagnosolutions/go-data/pkg/tunnel"
)

func main() {

	go func() {
		serverConf := tunnel.LoadServerTLSConfig()

		ln1, err := net.Listen("tcp", ":7001")
		if err != nil {
			panic(err)
		}
		ln2 := tls.NewListener(ln1, serverConf)

		pipe := func(w io.Writer, r io.Reader) {
			_, err := io.Copy(w, r)
			if err != nil {
				panic(err)
			}
		}

		handler := func(conn net.Conn) {
			go pipe(os.Stdout, conn)
			go pipe(conn, os.Stdin)
		}

		for {
			conn, err := ln2.Accept()
			if err != nil {
				continue
			}
			fmt.Printf("got connection from %s\n", conn.RemoteAddr())

			go handler(conn)
		}
	}()

	clientConf := tunnel.LoadClientTLSConfig()

	conn, err := tls.Dial("tcp", "localhost:7001", clientConf)
	if err != nil {
		panic(err)
	}

	conn.Write([]byte("foo bar\n"))
	conn.Write([]byte("baz\n"))

	conn.Close()

	// e1, err := tunnel.NewEndpoint("e1", "0.0.0.0:7000", "204.11.243.143:7001")
	// if err != nil {
	// 	panic(err)
	// }
	// err = e1.Connect()
	// if err != nil {
	// 	panic(err)
	// }

}
