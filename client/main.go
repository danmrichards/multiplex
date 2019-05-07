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

var cert, key, server, port string

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

	// Use a custom HTTP transport with the extended root CA pool and the
	// key/pair we generated above. The key pair will be presented to the other
	// side of the connection and verified. We could avoid this if we had a
	// certificate from a root CA already trusted by the client and the server.
	// Or if you want to live dangerously by disabling verification.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{xkp},
			RootCAs:      crtPool,
		},
	}

	// Because we created a custom TLSClientConfig, we have to opt-in to HTTP/2.
	// See https://github.com/golang/go/issues/14275
	if err = http2.ConfigureTransport(tr); err != nil {
		log.Fatalln("could not configure transport for HTTP/2:", err)
	}

	client := &http.Client{
		Transport: tr,
	}

	svr := net.JoinHostPort(server, port)

	res, err := request(client, fmt.Sprintf(pingURL, protocolHTTP, svr), nil)
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
	res, err = request(client, fmt.Sprintf(pingURL, protocolHTTPS, svr), nil)
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

	res, err = request(client, "https://www.google.com", nil)
	if err != nil {
		log.Fatalln("google request:", err)
	}
	log.Printf("response: protocol: %s status: %d", res.Proto, res.StatusCode)

	if res.StatusCode != http.StatusOK {
		log.Fatalln("could not ping google")
	}
	fmt.Println("pinged google")
}

// request sends an HTTP request and returns an HTTP response for the given
// server, uri, port and protocol.
func request(client *http.Client, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, body)
	if err != nil {
		log.Fatalln("build request:", err)
	}
	log.Printf("request: protocol: %s url: %s", req.Proto, req.URL)

	return client.Do(req)
}
