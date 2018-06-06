---
title: "runtime.gogc"
---

`runtime.gogc` configures GOGC (the GC target percentage).

Setting `runtime.gogc` is equivalent to setting the `GOGC`
environment variable which also takes precedence over
the value from the config file.

Increasing this value means fewer but longer GC cycles
since there is more garbage to collect.

The default of `GOGC=100` works for Go 1.4 but shows
a significant performance drop for Go 1.5 since the
concurrent GC kicks in more often.

During benchmarking I have found the following values
to work for my setup and for now I consider them sane
defaults for both Go 1.4 and Go 1.5.

	GOGC=100: Go 1.5 40% slower than Go 1.4
	GOGC=200: Go 1.5 == Go 1.4 with GOGC=100 (default)
	GOGC=800: both Go 1.4 and 1.5 significantly faster (40%/go1.4, 100%/go1.5)

The default is

	runtime.gogc = 800
