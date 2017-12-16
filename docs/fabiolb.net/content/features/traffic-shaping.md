---
title: "Traffic Shaping"
since: "1.0"
---

fabio allows to control the amount of traffic a set of service instances will
receive. You can use this feature to direct a fixed percentage of traffic to a
newer version of an existing service for testing ("Canary testing"). See
[Manual Overrides](./Routing#manual-overrides) for a complete description of the `route
weight` command.

The following command will allocate 5% of traffic to `www.kjca.dev/auth/` to
all instances of `service-b` which match tags `version-15` and `dc-fra`. This
is independent of the number of actual instances running. The remaining 95%
of the traffic will be distributed evenly across the remaining instances
publishing the same prefix.

```
route weight service-b www.kjca.dev/auth/ weight 0.05 tags "version-15,dc-fra"
```

