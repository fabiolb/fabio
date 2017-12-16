---
title: "registry.consul.register.checkInterval"
---

`registry.consul.register.checkInterval` configures the interval for the health check.

Fabio registers an http health check on http(s)://[ui.addr](/ref/ui.addr)/health
and this value tells consul how often to check it.

The default is

	registry.consul.register.checkInterval = 1s
