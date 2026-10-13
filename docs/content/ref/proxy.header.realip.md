---
title: "proxy.header.realip"
---

`proxy.header.realip` configures the name of the real client IP header. Fabio
preserves its existing non-empty value and sets it to the connection peer IP
from [http.Request.RemoteAddr](https://pkg.go.dev/net/http#Request.RemoteAddr)
only when the header is missing or empty.

The default is

    proxy.header.realip = X-Real-Ip

For a trusted CloudFront/Lambda@Edge header, configure:

    proxy.header.realip = X-Mycdn-Ip

The backend must read `X-Mycdn-Ip`; Fabio does **not** copy it into `X-Real-Ip`.
With the default `proxy.header.clearclient = false`, incoming real-IP values are
trusted, so allow only a trusted upstream to set them. Setting
`proxy.header.clearclient = true` discards incoming protected values and uses
the immediate connection peer IP instead.
