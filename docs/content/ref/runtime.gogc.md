---
title: "runtime.gogc"
---

`runtime.gogc` configures GOGC (the GC target percentage).

Setting `runtime.gogc` is equivalent to setting the `GOGC`
environment variable which also takes precedence over
the value from the config file.

Increasing this value means fewer but longer GC cycles
since there is more garbage to collect.

NOTE - the default for fabio up to 1.5.14 was 800.  This changed
to 100 in version 1.5.15

The default is

	runtime.gogc = 100
