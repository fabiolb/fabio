---
title: "proxy.grpcshutdowntimeout"
---

`proxy.grpcshutdowntimeout` configures the amount of time fabio will wait to attempt
to close the connection while waiting for grpc traffic to finish to a backend that's been
deregistered.  The default value is

    proxy.grpcshutdowntimeout = 2s
