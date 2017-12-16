---
title: "proxy.header.clientip"
---

`proxy.header.clientip` configures the header for the request ip.

The remote ip address is taken from [http.Request.RemoteAddr](https://golang.org/pkg/net/http/#Request.RemoteAddr).

The default is

    proxy.header.clientip =

