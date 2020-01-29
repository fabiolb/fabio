---
title: "registry.consul.register.deregisterCriticalServiceAfter"
---

This option is deprecated and has no effect in versions after 1.5.11. Services
are now always deregistered shortly after fabio exits for any reason.

In versions up to and including 1.5.11
`registry.consul.register.deregisterCriticalServiceAfter` configures the
duration after which registered services are removed from Consul after fabio
exits abruptly (services are always deregistered immediately when fabio exits
normally).

At the time of this writing, Consul enforces a minimum value of one minute and runs
its reaper process every thirty seconds. 

The default for fabio <= 1.5.11 is

	registry.consul.register.deregisterCriticalServiceAfter = 90m
