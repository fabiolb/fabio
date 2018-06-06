---
title: "proxy.localip"
---

`proxy.localip` configures the ip address of the proxy which is added
to the Header configured by [`header.clientip`](/ref/header.clientip/) and to the `Forwarded: by=` attribute.

The local non-loopback address is detected during startup
but can be overwritten with this property.

The default is

    proxy.localip =
