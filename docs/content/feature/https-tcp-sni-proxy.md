---
title: "HTTPS TCP-SNI Proxy"
since: "1.5.14"
---

fabio can run a TCP+SNI routing proxy on a listener, and have fallback to https functionality.
 This is effectively an amalgam of the TCP-SNI Proxy and the HTTPS functionality.
 
 To enable this feature configure a listener as follows:
 
 ```
 fabio -proxy.addr=':443;proto=https+tcp+sni;cs=somecertstore'
 ```
 
For host matches that are proto=tcp or have a scheme of tcp://, this will proxy TCP using SNI.

You would register your service in [Consul](https://consul.io) with a `urlprefix-` tag that
matches the host from the SNI extension for any services that should be proxied TCP (TLS
terminated by upstream).  If the upstream service you'd like to proxy TCP responds to
`https://foo.com/...` then you should register a `urlprefix-foo.com/ proto=tcp` tag for this
service.

For path based matching, you would do the typical `urlprefix-/path/` and this would cause
fabio to terminate TLS using the cs= line specified in the config.
