---
title: "registry.consul.register.checkTimeout"
---

`registry.consul.register.checkTimeout` configures the timeout for the health check.

Fabio registers an http health check on http(s)://[ui.addr](/ref/ui.addr)/health
and this value tells Consul how long to wait for a response.

The default is

	registry.consul.register.checkTimeout = 3s
