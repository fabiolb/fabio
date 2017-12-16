---
title: "registry.consul.register.checkTLSSkipVerify"
---

`registry.consul.register.checkTLSSkipVerify` configures TLS verification for the health check.

Fabio registers an http health check on http(s)://[ui.addr](/ref/ui.addr)/health
and this value tells consul to skip TLS certificate validation for
https checks.

The default is

	registry.consul.register.checkTLSSkipVerify = false
