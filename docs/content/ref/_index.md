---
title: "Reference"
weight: 600
---

All configuration options can be specified either 

* in the config file
* as environment variable
* as command line argument

and are evaluated in that order. 

```
# fabio.properties
metrics.target = stdout

# correspondig env var (no prefix)
metrics_target=stdout ./fabio

# env var with FABIO_ prefix (>= 1.2)
FABIO_metrics_target=stdout ./fabio

# env var with FABIO_ prefix (case-insensitive) (>= 1.2)
FABIO_METRICS_TARGET=stdout ./fabio

# command line argument (>= 1.2)
./fabio -metrics.target stdout
```
