# Multiplex
A proof-of-concept set of Golang applications to illustrate self-signed
SSL certificates with a multiplexing server (i.e. serving HTTP and HTTPS
on the same port). Uses the [cmux](https://github.com/soheilhy/cmux)
package for multiplexing.

## Summary
Within this repo are 2 applications:
1. The multiplexing server
2. A client

The client simple sends a "ping" to the server and expects to get a
"pong" back again.

The client makes the request once over HTTP and once over HTTPS. Both
requests are sent to the same host and port, which will verify if the
server multiplexing is working as expected.

You will notice that the HTTPS request is automatically upgraded to
HTTP/2, This is a nice side benefit of Go's built in HTTP client and
server packages (although because we're using self signed certs some
extra config has had to be made).

## Usage
Generate the certificate and key:
```bash
$ openssl req \
    -x509 \
    -nodes \
    -newkey rsa:2048 \
    -keyout ssl/server.key \
    -out ssl/server.crt \
    -days 3650 \
    -subj "/C=GB/ST=Bournemouth/L=Bournemouth/O=FooBar/OU=Turbo Encabulator/CN=*"
```

Build the binaries:
```bash
$ make
```

### Client
```bash
Usage of ./bin/client-linux-amd64:
  -cert string
    	Path to the SSL certificate file for the server (default "ssl/server.crt")
  -key string
    	Path to the SSL private key file for the server (default "ssl/server.key")
  -port string
    	The port on which to ping the server (default "8080")
  -server string
    	Server for the server to ping (default "localhost")
```

### Server
```bash
Usage of ./bin/server-linux-amd64:
  -cert string
    	Path to the SSL certificate for the server (default "ssl/server.crt")
  -key string
    	Path to the SSL private key for the server (default "ssl/server.key")
  -port string
    	The port on which to serve (default "8080")
```
