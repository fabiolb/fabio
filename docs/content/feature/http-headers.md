---
title: "HTTP Header Support"
since: "1.1.3"
---

In addition to injecting the `Forwarded` and (by default) `X-Real-Ip` headers, the
`X-Forwarded-For`, `X-Forwarded-Port` and `X-Forwarded-Proto` headers are added
to HTTP(S) and Websocket requests. The `proxy.header.realip` option changes the
real-IP header name (default `X-Real-Ip`), for example to `X-Mycdn-Ip`. Fabio
preserves a non-empty incoming value of that header, filling in the immediate
connection peer IP only when absent or empty. The backend must read the new
header name; Fabio does not also populate `X-Real-Ip` when a custom name is used.

Only trust incoming real-IP values from a controlled upstream (such as
CloudFront/Lambda@Edge). When `proxy.header.clearclient = true`, Fabio clears
the incoming protected header before inserting the peer address. The separate
`proxy.header.clientip` option writes the immediate peer address to another
header. TLS header options are `proxy.header.tls` and `proxy.header.tls.value`.

Since version 1.5.3 fabio also sets the `X-Forwarded-Host` header.
