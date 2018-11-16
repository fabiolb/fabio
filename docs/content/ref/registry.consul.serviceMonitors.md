---
title: "registry.consul.serviceMonitors"
---

`registry.consul.serviceMonitors` configures the concurrency for
route updates. Fabio will make up to the configured number of
concurrent calls to Consul to fetch status data for route
updates.

The default is

	registry.consul.serviceMonitors = 1
