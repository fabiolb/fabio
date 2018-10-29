---
title: "Access Control"
since: "1.5.8"
---

fabio supports basic ip centric access control per route.  You may
specify one of `allow` or `deny` options per route to control access.
Currently only source ip control is available.

<!--more-->

To allow access to a route from clients within the `192.168.1.0/24`
and `fe80::/10` subnet you would add the following option:

```
allow=ip:192.168.1.0/24,ip:fe80::/10
```

With this specified only clients sourced from those two subnets will
be allowed.  All other requests to that route will be denied.


Inversely, to deny a specific set of clients you can use the
following option syntax:

```
deny=ip:fe80::1234,100.123.0.0/16
```

With this configuration access will be denied to any clients with
the `fe80::1234` address or coming from the `100.123.0.0/16` network.

Single host addresses (addresses without a prefix) will have a
`/32` prefix, for IPv4, or a `/128` prefix, for IPv6, added automatically.
That means `1.2.3.4` is equivalent to `1.2.3.4/32` and `fe80::1234`
is equivalent to `fe80::1234/128` when specifying
address blocks for `allow` or `deny` rules.

The source ip used for validation against the defined ruleset is
taken from information available in the request.

For `HTTP` requests the client `RemoteAddr` is always validated
followed by all elements of the `X-Forwarded-For` header, if
present.  When all of these elements match an `allow` the request
will be allowed; similarly when any element matches a `deny` the
request will be denied.

For `TCP` requests the source address of the network socket
is used as the sole paramater for validation.

If the inbound connection uses the [PROXY protocol](https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt)
to transmit the true source address of the client then it will
be used for both `HTTP` and `TCP` connections for validating access.

