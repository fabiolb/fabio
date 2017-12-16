---
title: "registry.consul.tagprefix"
---

`registry.consul.tagprefix` configures the prefix for tags which define routes.

Services which define routes publish one or more tags with host/path
routes which they serve. These tags must have this prefix to be
recognized as routes.

The default is

	registry.consul.tagprefix = urlprefix-
