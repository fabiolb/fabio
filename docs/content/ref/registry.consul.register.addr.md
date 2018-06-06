---
title: "registry.consul.register.addr"
---

`registry.consul.register.addr` configures the address for the service registration.

Fabio registers itself in consul with this `host:port` address.
It must point to the UI/API endpoint configured by [ui.addr](/ref/ui.addr/) and defaults to its
value.

The default is

	registry.consul.register.addr = :9998
