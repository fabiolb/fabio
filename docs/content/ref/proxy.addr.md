---
title: "proxy.addr"
---


`proxy.addr` configures listeners.

Each listener is configured with and address and a
list of optional arguments in the form of

    [host]:port;opt=arg;opt[=arg];...

Each listener has a protocol which is configured
with the `proto` option for which it routes and
forwards traffic.

The supported protocols are:

* `http` for HTTP based protocols
* `https` for HTTPS based protocols
* `grpc` for GRPC based protocols
* `grpcs` for GRPC+TLS based protocols
* `tcp` for a raw TCP proxy with or witout TLS support
* `tcp+sni` for an SNI aware TCP proxy

If no `proto` option is specified then the protocol
is either `http` or `https` depending on whether a
certificate source is configured via the `cs` option
which contains the name of the certificate source.

The TCP+SNI proxy analyzes the `ClientHello` message
of TLS connections to extract the server name
extension and then forwards the encrypted traffic
to the destination without decrypting the traffic.

#### General options

* `rt`: Sets the read timeout as a duration value (e.g. `3s`)

* `wt`: Sets the write timeout as a duration value (e.g. `3s`)

* `strictmatch`: When set to `true` the certificate source must provide
  a certificate that matches the hostname for the connection
  to be established. Otherwise, the first certificate is used
  if no matching certificate was found. This matches the default
  behavior of the Go TLS server implementation.

* `pxyproto`: When set to 'true' the listener will respect upstream v1
  PROXY protocol headers.
  NOTE: PROXY protocol was on by default from 1.1.3 to 1.5.10.
  This changed to off when this option was introduced with
  the 1.5.11 release.
  For more information about the PROXY protocol, please see:
  http://www.haproxy.org/download/1.5/doc/proxy-protocol.txt

* `pxytimeout`: Sets PROXY protocol header read timeout as a duration (e.g. '250ms').
  This defaults to 250ms if not set when 'pxyproto' is enabled.

#### TLS options

* `tlsmin`: Sets the minimum TLS version for the handshake. This value
  is one of `ssl30`, `tls10`, `tls11`, `tls12` or the corresponding
  version number from https://golang.org/pkg/crypto/tls/#pkg-constants

* `tlsmax`: Sets the maximum TLS version for the handshake. See `tlsmin`
  for the format.

* `tlsciphers`: Sets the list of allowed ciphers for the handshake. The value
  is a quoted comma-separated list of the hex cipher values or
  the constant names from https://golang.org/pkg/crypto/tls/#pkg-constants,
  e.g. `"0xc00a,0xc02b"` or `"TLS_RSA_WITH_RC4_128_SHA,TLS_RSA_WITH_AES_128_CBC_SHA"`

#### Examples

    # HTTP listener on port 9999
    proxy.addr = :9999

    # HTTP listener on IPv4 with read timeout
    proxy.addr = 1.2.3.4:9999;rt=3s

    # HTTP listener on IPv6 with write timeout
    proxy.addr = [2001:DB8::A/32]:9999;wt=5s

    # Multiple listeners
    proxy.addr = 1.2.3.4:9999;rt=3s,[2001:DB8::A/32]:9999;wt=5s

    # HTTPS listener on port 443 with certificate source
    proxy.addr = :443;cs=some-name

    # HTTPS listener on port 443 with certificate source and TLS options
    proxy.addr = :443;cs=some-name;tlsmin=tls10;tlsmax=tls11;tlsciphers="0xc00a,0xc02b"
    
    # GRPC listener on port 8888 
    proxy.addr = :8888;proto=grpc
    
    # GRPCS listener on port 8888 with certificate source
    proxy.addr = :8888;proto=grpcs;cs=some-name

    # TCP listener on port 1234 with port routing
    proxy.addr = :1234;proto=tcp

    # TCP listener on port 443 with SNI routing
    proxy.addr = :443;proto=tcp+sni

The default is

    proxy.addr = :9999
