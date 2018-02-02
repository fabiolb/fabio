---
title: "registry.consul.kvpath"
---

`registry.consul.kvpath` configures the KV path for manual routes.

The Consul KV path is watched for changes which get appended to
the routing table. This allows for manual overrides and weighted
round-robin routes.

As of version 1.5.7 fabio will treat the kv path as a prefix and
combine the values of the key itself and all its subkeys in 
alphabetical order.

To see all updates you may want to set [`-log.routes.format`](/ref/log.routes.format/)
to `all`.

You can modify the content of the routes with the `consul` tool or via
the [Consul API](https://www.consul.io/api/index.html):

```
consul put fabio/config "route add svc /maint http://5.6.7.8:5000\nroute add svc / http://1.2.3.4:5000\n"

# fabio >= 1.5.7 supports prefix match
consul put fabio/config/maint    "route add svc /maint http://5.6.7.8:5000"
consul put fabio/config/catchall "route add svc / http://1.2.3.4:5000"

consul delete fabio/config/maint
```

The default is

	registry.consul.kvpath = /fabio/config
