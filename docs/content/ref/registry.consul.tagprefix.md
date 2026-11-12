---
title: "registry.consul.tagprefix"
---

`registry.consul.tagprefix` configures the prefix for both legacy service tags
and service metadata keys which define routes.

For new routes, define each host/path route in a Consul service metadata
**value**. Its key must start with the configured prefix and be unique per
service, for example:

```json
{
  "tags": ["fabio", "prod"],
  "meta": {
    "urlprefix-0": "/users",
    "urlprefix-1": "example.com/ proto=https"
  }
}
```

The metadata keys use a unique suffix (`0`, `1`, etc., after the configured
prefix); keep routes and their options in the values. Consul metadata keys
allow ASCII letters, numbers, underscores, and hyphens (up to 128 characters).
Values are limited to 512 characters, with up to 64 metadata pairs per service.

The exact `fabio` service tag is needed to discover services that define
routes only in metadata: Consul health-state checks include service tags
but not service metadata. A service without `fabio` or any prefixed route
tag is not discovered.

Existing prefixed service tags continue to work and may coexist with metadata
routes. For example, `urlprefix-/users` still defines a route without the
`fabio` discovery tag. Other service tags remain route labels.

If metadata-only changes do not advance the Consul health-state query index,
set `registry.consul.pollinterval = 5s` (or another nonzero duration) to
poll for updates periodically.

The default is

	registry.consul.tagprefix = urlprefix-
