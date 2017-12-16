---
title: "TCP-SNI Proxy"
since: "1.3"
---

fabio can run a transparent TCP proxy with SNI support which can forward any TLS connection
**without re-encrypting the traffic**. fabio captures the `ClientHello` packet which is the
first packet of the TLS handshake and extracts the server name from the SNI extension and
uses it for finding the upstream server to forward the connection to. It then replays the
`ClientHello` packet and then transparently forwards all traffic between client and server
as a byte stream.

To enable this feature configure a listener as follows:

```
fabio -proxy.addr=':443;proto=tcp+sni'
```

to listen to more than 1 port separate with comma's (like if you want to do tcp and http listening):
```
fabio -proxy.addr ':9999,:19587;proto=tcp
```
This will do normal fabio http(s) routing on port 9999 and TCP proxy on port 19587.

and register your services in [Consul](https://consul.io/) with a `urlprefix-` tag that
matches the host from the SNI extension. If your server responds to `https://foo.com/...`
then you should register a `urlprefix-foo.com/` tag for this service. Note that the tag
should only contain  `<host>/` since path-based routing is not possible with this approach.
