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

### Vault Example

[Vault](https://www.vaultproject.io) is a tool by [HashiCorp](https://www.hashicorp.com/) for managing secrets and protecting sensitive data. When running in HA mode, Vault will have a single active node which is responsible for responding the API requests. Fabio can be used to ensure traffic is routed to the correct server via traffic shaping.

The following command will allocate 100% of traffic to `vault.company.com` to the instance of `vault` which is registered with the tag `active`.

```
route weight vault vault.company.com weight 1.00 tags "active"
```
