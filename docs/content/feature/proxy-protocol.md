---
title: "PROXY Protocol Support"
since: "1.1.3"
---

fabio transparently supports the HA Proxy
[PROXY protocol](http://www.haproxy.org/download/1.5/doc/proxy-protocol.txt) version 1
which is used by HA Proxy,
[Amazon ELB](http://docs.aws.amazon.com/ElasticLoadBalancing/latest/DeveloperGuide/enable-proxy-protocol.html)
and others to transmit the remote address and port of the client without using headers.
