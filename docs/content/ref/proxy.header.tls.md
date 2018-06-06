---
title: "proxy.header.tls"
---

`proxy.header.tls` configures the header to set for TLS connections.

When set to a non-empty value the proxy will set this header on every
TLS request to the value of [proxy.header.tls.value](/ref/proxy.header.tls.value/)

The default is

    proxy.header.tls =
