---
title: "proxy.strategy"
---

`proxy.strategy` configures the load balancing strategy.

* `rnd`: pseudo-random distribution
  configures a pseudo-random distribution by using the microsecond
  fraction of the time of the request.

* `rr`:  round-robin distribution
  configures a round-robin distribution.

The default is

    proxy.strategy = rnd

