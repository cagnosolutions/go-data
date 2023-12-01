package main

import (
	"flag"
	"fmt"
)

var help = `Ncat 7.94 ( https://nmap.org/ncat )
           Usage: ncat [options] [hostname] [port]

           Options taking a time assume seconds. Append 'ms' for milliseconds,
           's' for seconds, 'm' for minutes, or 'h' for hours (e.g. 500ms).
             -4                         Use IPv4 only
             -6                         Use IPv6 only
             -U, --unixsock             Use Unix domain sockets only
                 --vsock                Use vsock sockets only
             -C, --crlf                 Use CRLF for EOL sequence
             -c, --sh-exec <command>    Executes the given command via /bin/sh
             -e, --exec <command>       Executes the given command
                 --lua-exec <filename>  Executes the given Lua script
             -g hop1[,hop2,...]         Loose source routing hop points (8 max)
             -G <n>                     Loose source routing hop pointer (4, 8, 12, ...)
             -m, --max-conns <n>        Maximum <n> simultaneous connections
             -h, --help                 Display this help screen
             -d, --delay <time>         Wait between read/writes
             -o, --output <filename>    Dump session data to a file
             -x, --hex-dump <filename>  Dump session data as hex to a file
             -i, --idle-timeout <time>  Idle read/write timeout
             -p, --source-port port     Specify source port to use
             -s, --source addr          Specify source address to use (doesn't affect -l)
             -l, --listen               Bind and listen for incoming connections
             -k, --keep-open            Accept multiple connections in listen mode
             -n, --nodns                Do not resolve hostnames via DNS
             -t, --telnet               Answer Telnet negotiations
             -u, --udp                  Use UDP instead of default TCP
                 --sctp                 Use SCTP instead of default TCP
             -v, --verbose              Set verbosity level (can be used several times)
             -w, --wait <time>          Connect timeout
             -z                         Zero-I/O mode, report connection status only
                 --append-output        Append rather than clobber specified output files
                 --send-only            Only send data, ignoring received; quit on EOF
                 --recv-only            Only receive data, never send anything
                 --no-shutdown          Continue half-duplex when receiving EOF on stdin
                 --allow                Allow only given hosts to connect to Ncat
                 --allowfile            A file of hosts allowed to connect to Ncat
                 --deny                 Deny given hosts from connecting to Ncat
                 --denyfile             A file of hosts denied from connecting to Ncat
                 --broker               Enable Ncat's connection brokering mode
                 --chat                 Start a simple Ncat chat server
                 --proxy <addr[:port]>  Specify address of host to proxy through
                 --proxy-type <type>    Specify proxy type ("http", "socks4", "socks5")
                 --proxy-auth <auth>    Authenticate with HTTP or SOCKS proxy server
                 --proxy-dns <type>     Specify where to resolve proxy destination
                 --ssl                  Connect or listen with SSL
                 --ssl-cert             Specify SSL certificate file (PEM) for listening
                 --ssl-key              Specify SSL private key (PEM) for listening
                 --ssl-verify           Verify trust and domain name of certificates
                 --ssl-trustfile        PEM file containing trusted SSL certificates
                 --ssl-ciphers          Cipherlist containing SSL ciphers to use
                 --ssl-servername       Request distinct server name (SNI)
                 --ssl-alpn             ALPN protocol list to use
                 --version              Display Ncat's version information and exit`

const (
	CONNECT_MODE  = 0
	LISTEN_MODE   = 1
	TCP_MODE_BOTH = 0
	TCP_MODE_4    = 1
	TCP_MODE_6    = 2
)

var (
	tcp4 = flag.Bool("4", false, "Use IPv4 only (default)")
	tcp6 = flag.Bool("6", false, "Use IPv6 only")
	udp  = flag.Bool("u", false, "Use UDP for the connection (default is TCP)")
	crlf = flag.Bool("C", false, "Use CRLF for EOL sequence")
	c    = flag.String("c", "", "Executes the given command via /bin/sh")
	e    = flag.String("e", "", "Executes the given command")
	l    = flag.Bool("l", false, "Bind and listen for incoming connections")
	p    = flag.Int("p", 31337, "Set the port number for Ncat to bind to")
	m    = flag.Int("m", 100, "Maximum simultaneous connections")
	k    = flag.Bool("k", false, "Accept multiple connections after one is closed")
	h    = flag.Bool("h", false, "Display this help screen")
	o    = flag.Bool("o", false, "-o <filename> Dump session data to a file")
	v    = flag.Bool("v", false, "Set verbosity level (can be used several times)")

	connMode int // 0 is connect mode (client), 1 is listen mode (server)
	tcpMode  int // 0 is both, 1 is tcp4, 2 is tcp6
)

// var (
// 	listen = flag.Bool("l", false, "Listen")
// 	host   = flag.String("h", "localhost", "Host")
// 	port   = flag.Int("p", 0, "Port")
// )

func address(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

var (
	name        = "goncat"
	version     = "1.0.0"
	description = "Concatenate and redirect sockets"
)

type ncat struct {
	network string
	useCRLF bool
	listen  bool
	addr    string
	port    string
}

func newNCat(network string, useCRLF bool, listen bool, addr, port string) *ncat {
	return &ncat{
		network: network,
		useCRLF: useCRLF,
		listen:  listen,
		addr:    addr,
		port:    port,
	}
}

func main() {

	// usage function
	Usage := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s %s %s\n\n", name, version, description)
		fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n%s [OPTIONS...] [hostname] [port]:\n", name)
		flag.PrintDefaults()
	}

	// parse command line flags
	flag.Parse()

	// print help and usage
	if *h {
		Usage()
		return
	}
	// listen mode
	if *l {
		connMode = LISTEN_MODE
	}
	// tcp4 mode
	if *tcp4 && !*tcp6 {
		tcpMode = TCP_MODE_4
	}
	// tcp6 mode
	if !*tcp4 && *tcp6 {
		tcpMode = TCP_MODE_6
	}

	if connMode == {

	}

	// flag.Parse()
	// if *listen {
	// 	startServer()
	// 	return
	// }
	// if len(flag.Args()) < 2 {
	// 	fmt.Println("Hostname and port required")
	// 	return
	// }
	// serverHost := flag.Arg(0)
	// serverPort := flag.Arg(1)
	// startClient(fmt.Sprintf("%s:%s", serverHost, serverPort))
}

// func startServer() {
// 	// compose server address
// 	addr, err := net.ResolveTCPAddr("tcp", address(*host, *port))
// 	if err != nil {
// 		panic(err)
// 	}
// 	// start tcp listener
// 	ln, err := net.ListenTCP("tcp", addr)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	for {
// 		conn, err := ln.Accept()
// 		if err != nil {
// 			log.Printf("error accepting connection from client: %s", err)
// 			continue
// 		}
// 		go processConn(conn)
// 	}
// }
//
// func startClient(addr string) {
// 	conn, err := net.Dial("tcp", addr)
// 	if err != nil {
// 		fmt.Printf("Can't connect to server: %s\n", err)
// 		return
// 	}
// 	_, err = io.Copy(conn, os.Stdin)
// 	if err != nil {
// 		fmt.Printf("Connection error: %s\n", err)
// 	}
// }
//
// func processConn(conn net.Conn) {
// 	_, err := io.Copy(os.Stdout, conn)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	conn.Close()
// }
