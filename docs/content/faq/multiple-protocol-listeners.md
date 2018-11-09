---
title: "Handling Multiple Protocols"
---

It is quite possible for a single fabio instance to serve multiple protocols 
via distinct listeners.

In this example:
```
proxy.addr = 172.16.20.11:80;proto=http;rt=60s;wt=30s,\
             172.16.20.11:443;proto=https;rt=60s;wt=30s;cs=all;tlsmin=10, \
             172.16.20.11:8443;proto=tcp+sni
```

We are telling fabio to bind to `172.16.20.11` on three different ports 
(`80`, `443`, and `8443`) using three distinct protocols 
(`HTTP`, `HTTPS`, `TCP+SNI`).  You are free to bind to as many address, 
port, and protocol combinations as needed within a single instance.

See [#490](https://github.com/fabiolb/fabio/issues/490) for context.
