package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"golang.org/x/net/http2"
)

var (
	cert, key, server, port, serverName string
	httpClient                          = http.Client{}
)

const (
	protocolHTTP  = "http"
	protocolHTTPS = "https"

	pingURL = "%s://%s/ping"
)

func main() {
	flag.StringVar(&cert, "cert", "ssl/server.crt", "Path to the SSL certificate file for the server")
	flag.StringVar(&key, "key", "ssl/server.key", "Path to the SSL private key file for the server")
	flag.StringVar(&server, "server", "localhost", "Server for the server to ping")
	flag.StringVar(&port, "port", "8080", "The port on which to ping the server")
	flag.StringVar(&serverName, "servername", "foobar.com", "The FQDN for which the SSL certificate is valid")

	flag.Parse()

	// Read the certificate file.
	crt, err := ioutil.ReadFile(cert)
	if err != nil {
		log.Fatalln("read cert:", err)
	}

	// Get the certificate pool.
	crtPool, err := x509.SystemCertPool()
	if err != nil {
		// Just log the error instead of a fatal. In many cases (e.g. when
		// running on windows) we won't be able to get the system cert pool
		// at all. Better to just log and attempt to use the default pool.
		log.Println("could not load system cert pool:", err)
	}
	if crtPool == nil {
		crtPool = x509.NewCertPool()
	}

	// Add the self-signed cert to the CA pool.
	if ok := crtPool.AppendCertsFromPEM(crt); !ok {
		log.Fatalln("could not append certificate to the pool")
	}

	xkp, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		log.Fatalln("could not load x509 key pair:", err)
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{xkp},
		RootCAs:      crtPool,
		ServerName:   serverName,
	}

	// We always make the requests to foobar.com as this is the FQDN that has
	// been used in the SSL certificate the server is using.
	svr := net.JoinHostPort(server, port)

	res, err := sniRequest(
		tlsCfg, http.MethodGet, fmt.Sprintf(pingURL, protocolHTTP, svr), nil,
	)
	if err != nil {
		log.Fatalln("ping HTTP request:", err)
	}
	log.Printf("response: protocol: %s status: %d", res.Proto, res.StatusCode)

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln("ping HTTP response:", err)
	}
	fmt.Println(string(b))

	fmt.Println()
	res, err = sniRequest(
		tlsCfg, http.MethodGet, fmt.Sprintf(pingURL, protocolHTTPS, svr), nil,
	)
	if err != nil {
		log.Fatalln("ping HTTPS request:", err)
	}
	log.Printf("response: protocol: %s status: %d", res.Proto, res.StatusCode)

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln("ping HTTPS response:", err)
	}
	fmt.Println(string(b))

	fmt.Println()
	fmt.Println("ping google just to prove the root certs still work")

	res, err = request(http.MethodGet, "https://www.google.com", nil)
	if err != nil {
		log.Fatalln("google request:", err)
	}
	log.Printf("response: protocol: %s status: %d", res.Proto, res.StatusCode)

	if res.StatusCode != http.StatusOK {
		log.Fatalln("could not ping google")
	}
	fmt.Println("pinged google")
}

func sniRequest(tlsCfg *tls.Config, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Host = serverName
	log.Printf("request: protocol: %s url: %s", req.Proto, req.URL)

	tr := &http.Transport{
		TLSClientConfig: tlsCfg,
	}

	// Because we created a custom TLSClientConfig, we have to opt-in to HTTP/2.
	// See https://github.com/golang/go/issues/14275
	if err = http2.ConfigureTransport(tr); err != nil {
		log.Fatalln("could not configure transport for HTTP/2:", err)
	}

	httpClient.Transport = tr
	return httpClient.Do(req)
}

func request(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	log.Printf("request: protocol: %s url: %s", req.Proto, req.URL)

	httpClient.Transport = http.DefaultTransport
	return httpClient.Do(req)
}
