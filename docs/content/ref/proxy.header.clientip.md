---
title: "proxy.header.clientip"
---

`proxy.header.clientip` configures the name of an optional client IP header.
Fabio preserves a non-empty incoming value of this header and fills in the
immediate connection peer IP when it is missing or empty. The peer address is
taken from [http.Request.RemoteAddr](https://pkg.go.dev/net/http#Request.RemoteAddr).

For an upstream CloudFront/Lambda@Edge header, configure:

    proxy.header.clientip = X-Mycdn-Ip

Only trust values supplied by a controlled upstream proxy. With the default
`proxy.header.clearclient = false`, an incoming value is forwarded unchanged;
setting `proxy.header.clearclient = true` discards it and fills in the peer IP.
The separate `X-Real-Ip` header is still populated or preserved independently.
`X-Forwarded-For` continues to use its normal forwarding rules.

The default is

    proxy.header.clientip =
