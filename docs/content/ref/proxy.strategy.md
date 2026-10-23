---
title: "proxy.strategy"
---

`proxy.strategy` configures the load balancing strategy.

* `rnd`: pseudo-random distribution
  configures a pseudo-random distribution by using the microsecond
  fraction of the time of the request.

* `rr`:  round-robin distribution
  configures a round-robin distribution.

* `lconn`: least active upstream requests/connections
  selects the least busy target within the group chosen by route weights.
  Ties rotate among targets. An active operation lasts from upstream handling
  until its HTTP request, gRPC RPC/stream, WebSocket/SSE session, or TCP session
  finishes. Idle keep-alive sockets do not count. Each Fabio process tracks its
  own active operations; counts are not shared across Fabio instances.

The default is

    proxy.strategy = rnd

