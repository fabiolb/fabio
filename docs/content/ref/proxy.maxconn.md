---
title: "proxy.maxconn"
---

`proxy.maxconn` configures the maximum number of cached
incoming and outgoing connections.

This configures the [MaxConnsPerHost](https://golang.org/pkg/net/http/#Transport.MaxConnsPerHost)
of the [http.Transport](https://golang.org/pkg/net/http/#Transport).

#### Default

    proxy.maxconn = 10000

