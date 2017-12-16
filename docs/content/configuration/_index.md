---
title: "Configuration"
weight: 150
---

One of the main objectives of fabio is that you do not have to configure it to do its job.

If you run fabio next to a [Consul](https://consul.io/) agent and your services are
[configured properly](/#quickstart) all you have to do is start it and forget about it.

By default fabio listens on port `9999` for HTTP traffic and uses
[Consul](https://consul.io/) on `localhost:8500` as the default registry backend.

Depending on your environment or requirements you may want to configure additional
listeners, different backends, enable metrics reporting or change other configuration
parameters. 

The full set of configurable options can be found in the 
[fabio.properties](https://raw.githubusercontent.com/fabiolb/fabio/master/fabio.properties)
file.

Each option can be configured through a properties file, an environment variable
or via command line arguments and are evaluated in that order.

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

