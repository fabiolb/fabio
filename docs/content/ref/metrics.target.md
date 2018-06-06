---
title: "metrics.target"
---

`metrics.target` configures the backend the metrics values are sent to.

Possible values are:

* `<empty>`:  do not report metrics
* `stdout`:   report metrics to stdout
* `graphite`: report metrics to Graphite on [metrics.graphite.addr](/ref/metrics.graphite.addr/)
* `statsd`: report metrics to StatsD on [metrics.statsd.addr](/ref/metrics.statsd.addr/)
* `circonus`: report metrics to Circonus (http://circonus.com/)

The default is

	metrics.target =
