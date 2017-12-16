---
title: "registry.static.routes"
---

`registry.static.routes` configures a static routing table.

#### Example

	registry.static.routes = \
		route add svc / http://1.2.3.4:5000/

The default is

	registry.static.routes =
