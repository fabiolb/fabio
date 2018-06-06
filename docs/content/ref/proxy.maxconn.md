---
title: "proxy.maxconn"
---

`proxy.maxconn` configures the maximum number of cached
incoming and outgoing connections.

This configures the [MaxIdleConnsPerHost](https://golang.org/pkg/net/http/#Transport.MaxIdleConnsPerHost)
of the [http.Transport](https://golang.org/pkg/net/http/#Transport).

The default is

    proxy.maxconn = 10000

