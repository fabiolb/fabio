---
title: "runtime.gomaxprocs"
---

`runtime.gomaxprocs` configures GOMAXPROCS.

Setting `runtime.gomaxprocs` is equivalent to setting the `GOMAXPROCS`
environment variable which also takes precedence over
the value from the config file.

If `runtime.gomaxprocs` < 0 then all CPU cores are used.

The default is

	runtime.gomaxprocs = -1
