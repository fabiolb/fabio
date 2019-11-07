# proxyproto

This library provides the `proxyproto` package which can be used for servers
listening behind HAProxy of Amazon ELB load balancers. Those load balancers
support the use of a proxy protocol (https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt),
which provides a simple mechansim for the server to get the address of the client
instead of the load balancer.

This library provides both a net.Listener and net.Conn implementation that
can be used to handle situation in which you may be using the proxy protocol.
Proxy protocol version 1 and 2 supported. Also this library supported mode as TLS server,

The only caveat is that we check for the header prefix to determine if the protocol
is being used. If that string may occur as part of your input, then it is ambiguous
if the protocol is being used and you may have problems.

# proxy protocol version 2

Initial support for proxy protocol version 2, need more work but it's working.

# Documentation

Full documentation can be found [here](http://godoc.org/github.com/armon/go-proxyproto).

# Examples

Using the library in TCP mode:

```

// Create a listener
list, err := net.Listen("tcp", "...")

// Wrap listener in a proxyproto listener (TCP mode)
proxyList := &proxyproto.Listener{Listener: list}
conn, err := proxyList.Accept()

...
```

Using the library in TLS mode:

```

// Create a TLS server config
cer, err := tls.X509KeyPair([]byte(serverCert), []byte(serverKey))
if err != nil {
	log.Fatal(err)
}
config := &tls.Config{Certificates: []tls.Certificate{cer}}

// Create a listener
list, err := net.Listen("tcp", "...")

// Wrap listener in a proxyproto listener (TLS mode)
proxyList := &proxyproto.Listener{Listener: list, TLSConfig: config}
conn, err := proxyList.Accept()

...
```
