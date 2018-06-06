---
title: "registry.consul.checksRequired"
---

`registry.consul.checksRequired` configures how many health checks 
must pass in order for fabio to consider a service available.

Possible values are:

* `one`: at least one health check must pass
* `all`: all health checks must pass

The default is

	registry.consul.checksRequired = one
