---
title: "registry.consul.register.enabled"
---

`registry.consul.register.enabled` configures whether fabio registers itself in Consul.

Fabio will register itself in consul only if this value is set to `true` which
is the default. To disable registration set it to any other value, e.g. `false`

The default is

	registry.consul.register.enabled = true
