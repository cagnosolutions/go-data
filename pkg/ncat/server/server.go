package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var (
	listen = flag.Bool("l", false, "Listen")
	host   = flag.String("h", "localhost", "Host")
	port   = flag.Int("p", 0, "Port")
)

func address(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

func main() {
	flag.Parse()
	if *listen {
		startServer()
		return
	}
	if len(flag.Args()) < 2 {
		fmt.Println("Hostname and port required")
		return
	}
	serverHost := flag.Arg(0)
	serverPort := flag.Arg(1)
	startClient(fmt.Sprintf("%s:%s", serverHost, serverPort))
}

func startServer() {
	// compose server address
	addr, err := net.ResolveTCPAddr("tcp", address(*host, *port))
	if err != nil {
		panic(err)
	}
	// start tcp listener
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("error accepting connection from client: %s", err)
			continue
		}
		go processConn(conn)
	}
}

func startClient(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("Can't connect to server: %s\n", err)
		return
	}
	_, err = io.Copy(conn, os.Stdin)
	if err != nil {
		fmt.Printf("Connection error: %s\n", err)
	}
}

func processConn(conn net.Conn) {
	_, err := io.Copy(os.Stdout, conn)
	if err != nil {
		fmt.Println(err)
	}
	conn.Close()
}
