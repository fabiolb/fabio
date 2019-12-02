---
title: "TCP Dynamic Proxy"
---

The TCP dynamic proxy is similar to the TCP Proxy, but the listener is started from the Consul urlprefix tag.
Also, the service is defined with IP and port, so that multiple services can be defined on the load balancer using
the same TCP port.  Connections are forwarded to services based on the combination of ip:port

To use TCP Dynamic proxy support the service needs to advertise `urlprefix-127.0.0.1:1234 proto=tcp` in
Consul. In addition, fabio needs to be configured with a placeholder for the proxy.addr.:

```
fabio -proxy.addr '0.0.0.0:0;proto=tcp-dynamic;refresh=5s'
```

The TCP listener is started for the given TCP ports.  To use IP addressing to separate the services, matching IP
addressed would need to be added to the loopback interface on the host.