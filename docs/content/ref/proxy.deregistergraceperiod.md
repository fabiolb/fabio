---
title: "proxy.deregistergraceperiod"
---

`proxy.deregistergraceperiod` configures the time to wait before 
shutting down the proxies de-registering from the service registry.

After a signal is caught Fabio will immediately de-register from the
service registry and wait for `proxy.deregistergraceperiod` letting
in-flight requests finish after which it will continue with shutting
down the proxy.

The default is

    proxy.deregistergraceperiod = 0s
