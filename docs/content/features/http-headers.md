---
title: "HTTP Header Support"
since: "1.1.3"
---

In addition, to injecting the `Forwarded` and `X-Real-Ip` headers the
`X-Forwarded-For`, `X-Forwarded-Port` and `X-Forwarded-Proto` headers are added
to HTTP(S) and Websocket requests. Custom headers for the ip address and
protocol can be configured with the `proxy.header.clientip`, `proxy.header.tls`
and `proxy.header.tls.value` options.

Since version 1.5.3 fabio also sets the `X-Forwarded-Host` header.
