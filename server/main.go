package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/soheilhy/cmux"
)

var port, cert, key string

// handlePing writes a pong back to the HTTP response.
func handlePing(w http.ResponseWriter, r *http.Request) {
	log.Printf("request: protocol: %s url: %s\n", r.Proto, r.URL)
	if r.TLS != nil {
		log.Printf(
			"request: tls: version: %d handshake: %v",
			r.TLS.Version, r.TLS.HandshakeComplete,
		)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong"))
}

func serveHTTP(l net.Listener) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", handlePing)

	s := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	}
	if err := s.Serve(l); err != cmux.ErrListenerClosed {
		log.Fatal(err)
	}
}

func serveHTTPS(l net.Listener) {
	// Load certificates.
	crt, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		log.Panic(err)
	}

	// Create TLS listener.
	//
	// Config recommendations from https://blog.cloudflare.com/exposing-go-on-the-internet/
	tl := tls.NewListener(l, &tls.Config{
		// Causes servers to use Go's default ciphersuite preferences,
		// which are tuned to avoid attacks. Does nothing on clients.
		PreferServerCipherSuites: true,

		// Only use curves which have assembly implementations.
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519, // Go >= 1.8 only.
		},

		// Could cause compatibility issues with older clients.
		MinVersion: tls.VersionTLS12,

		// Ensure safer and faster cipher suites.
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Go >= 1.8 only.
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,   // Go >= 1.8 only.
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,

			// Best disabled, as they don't provide Forward Secrecy,
			// but might be necessary for some clients
			// tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			// tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		},

		// Add our self-signed certificate.
		Certificates: []tls.Certificate{crt},

		// Manually enable support for HTTP/2.
		NextProtos: []string{"h2", "http/1.1"},
	})

	// Serve HTTP over TLS.
	serveHTTP(tl)
}

func main() {
	flag.StringVar(&port, "port", "8080", "The port on which to serve")
	flag.StringVar(&cert, "cert", "ssl/server.crt", "Path to the SSL certificate for the server")
	flag.StringVar(&key, "key", "ssl/server.key", "Path to the SSL private key for the server")

	flag.Parse()

	// Create the TCP listener.
	l, err := net.Listen("tcp", net.JoinHostPort("", port))
	if err != nil {
		log.Fatalln("could not create tcp listener:", err)
	}

	// Create a mux.
	m := cmux.New(l)

	// We first match on HTTP 1.1 methods.
	hl := m.Match(cmux.HTTP1Fast())

	// If not matched, we assume that its TLS.
	//
	// Note that you can take this listener, do TLS handshake and
	// create another mux to multiplex the connections over TLS.
	tl := m.Match(cmux.Any())

	go serveHTTP(hl)
	go serveHTTPS(tl)

	fmt.Println("Serving on port:", port)
	log.Fatalln(m.Serve())
}
