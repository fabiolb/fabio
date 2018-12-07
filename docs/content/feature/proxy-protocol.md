---
title: "PROXY Protocol Support"
since: "1.1.3"
---

fabio transparently supports the HA Proxy
[PROXY protocol](http://www.haproxy.org/download/1.5/doc/proxy-protocol.txt) version 1
which is used by HA Proxy,
[Amazon ELB](http://docs.aws.amazon.com/ElasticLoadBalancing/latest/DeveloperGuide/enable-proxy-protocol.html)
and others to transmit the remote address and port of the client without using headers.

You may control the behavior of PROXY protocol support with the following
options on the listener:

* `pxyproto`: When set to 'true' the listener will respect upstream v1
  PROXY protocol headers.
  NOTE: PROXY protocol was on by default from 1.1.3 to 1.5.10.
  This changed to off when this option was introduced with
  the 1.5.11 release.
  For more information about the PROXY protocol, please see:
  http://www.haproxy.org/download/1.5/doc/proxy-protocol.txt

* `pxytimeout`: Sets PROXY protocol header read timeout as a duration (e.g. '250ms').
  This defaults to 250ms if not set when 'pxyproto' is enabled.

See the comments in for `proxy.addr` in `fabio.properties` for more information.
