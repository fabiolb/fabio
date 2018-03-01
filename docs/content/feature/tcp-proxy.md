---
title: "TCP Proxy"
since: "1.4"
---

fabio can run a transparent TCP proxy which dynamically forwards an incoming
connection on a given port to services which advertise that port. To use TCP
proxy support the service needs to advertise `urlprefix-:1234 proto=tcp` in
Consul. In addition, fabio needs to be configured to listen on that port:

```
fabio -proxy.addr ':1234;proto=tcp'
```

TCP proxy support can be combined with [Certificate Stores](/feature/certificate-stores/) to provide TLS termination on fabio.

```
fabio -proxy.cs 'cs=ssl;type=path;path=/etc/ssl' -proxy.addr ':1234;proto=tcp;cs=ssl'
```

