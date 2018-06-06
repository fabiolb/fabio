---
title: "log.routes.format"
---

`log.routes.format` configures the log output format of routing table updates.

Changes to the routing table are written to the standard log. This option
configures the output format:

* `detail`:   detailed routing table as ascii tree
* `delta`:    additions and deletions in config language
* `all`:      complete routing table in config language

The default is

	log.routes.format = delta

