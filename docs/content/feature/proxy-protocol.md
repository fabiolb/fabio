---
title: "PROXY Protocol Support"
since: "1.1.3"
---

fabio transparently supports the HA Proxy
[PROXY protocol](http://www.haproxy.org/download/1.5/doc/proxy-protocol.txt) versions 1 and 2.
These protocol versions transmit the remote address and port of the client
without using headers. Version 1 is used by the Amazon Classic Load Balancer,
while version 2 is used by the Amazon Network Load Balancer (NLB).

You may control the behavior of PROXY protocol support with the following
options on the listener:

* `pxyproto`: When set to 'true' the listener will respect upstream PROXY
  protocol version 1 or version 2 headers.
  NOTE: PROXY protocol was on by default from 1.1.3 to 1.5.10.
  This changed to off when this option was introduced with
  the 1.5.11 release.
  For more information about the PROXY protocol, please see:
  http://www.haproxy.org/download/1.5/doc/proxy-protocol.txt

  This listener option is independent of the route option with the same name,
  which writes version 1 headers to TCP upstream connections.

* `pxytimeout`: Sets PROXY protocol header read timeout as a duration (e.g. '250ms').
  This defaults to 250ms if not set when 'pxyproto' is enabled.

See the comments in for `proxy.addr` in `fabio.properties` for more information.
