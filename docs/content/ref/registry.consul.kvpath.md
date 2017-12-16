---
title: "registry.consul.kvpath"
---

`registry.consul.kvpath` configures the KV path for manual routes.

The consul KV path is watched for changes which get appended to
the routing table. This allows for manual overrides and weighted
round-robin routes.

The default is

	registry.consul.kvpath = /fabio/config
