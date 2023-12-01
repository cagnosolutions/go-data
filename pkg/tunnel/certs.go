package tunnel

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"os"
)

var (
	//go:embed certs/server.crt
	LocalServerCRTFile []byte

	//go:embed certs/server.key
	LocalServerKeyFile []byte

	//go:embed certs/ca.crt
	LocalCAFile []byte
)

func LoadServerTLSConfig() *tls.Config {
	cert, err := tls.X509KeyPair(LocalServerCRTFile, LocalServerKeyFile)
	if err != nil {
		panic(err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	return tlsConfig
}

func LoadClientTLSConfig() *tls.Config {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(LocalCAFile)
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	return tlsConfig
}

func LoadServerTLSConfigFromFile(crtFile, keyFile string) *tls.Config {
	cert, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		panic(err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	return tlsConfig
}

func LoadClientTLSConfigFromFile(caFile string) *tls.Config {
	cert, err := os.ReadFile(caFile)
	if err != nil {
		panic(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cert)
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	return tlsConfig
}
