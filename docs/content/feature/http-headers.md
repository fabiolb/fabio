---
title: "HTTP Header Support"
since: "1.1.3"
---

In addition to injecting the `Forwarded` and `X-Real-Ip` headers, the
`X-Forwarded-For`, `X-Forwarded-Port` and `X-Forwarded-Proto` headers are added
to HTTP(S) and Websocket requests. The optional `proxy.header.clientip` setting
allows another client IP header name, for example `X-Mycdn-Ip`. Fabio preserves
a non-empty incoming value of that header, filling in the immediate connection
peer IP only when absent or empty. `X-Real-Ip` remains independent and retains
its original preserve-or-populate behavior.

Only trust incoming client IP values from a controlled upstream (such as
CloudFront/Lambda@Edge). When `proxy.header.clearclient = true`, Fabio clears
the incoming protected header before inserting the immediate peer address.
`X-Forwarded-For` retains its existing forwarding behavior. TLS header
options are `proxy.header.tls` and `proxy.header.tls.value`.

Since version 1.5.3 fabio also sets the `X-Forwarded-Host` header.
