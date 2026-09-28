---
title: "proxy.header.clearclient"
---

`proxy.header.clearclient` clear headers from client.

Configure the proxy to clear source address headers supplied by the client. This will prevent clients from manipulating headers normally supplied by Fabio.

The default is

    proxy.header.clearclient = false

