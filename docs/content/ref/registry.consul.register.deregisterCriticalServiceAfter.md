---
title: "registry.consul.register.deregisterCriticalServiceAfter"
---

`registry.consul.register.deregisterCriticalServiceAfter` configures the time the
service health check is allowed to be in state `critical` until Consul automatically
deregisters it.
If fabio is still running, the service will be re-registered almost immediately after
being deleted by Consul.

At the time of this writing, Consul enforces a minimum value of one minute and runs
its reaper process every thirty seconds. 

The default is

	registry.consul.register.deregisterCriticalServiceAfter = 90m