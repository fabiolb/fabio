---
title: "proxy.shutdownwait"
---

`proxy.shutdownwait` configures the time for a graceful shutdown.

After a signal is caught the proxy will immediately suspend
routing traffic and respond with a `503 Service Unavailable`
for the duration of the given period.

The default is

    proxy.shutdownwait = 0s
