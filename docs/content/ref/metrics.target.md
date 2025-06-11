---
title: "metrics.target"
---

`metrics.target` configures the backend the metrics values are sent to.

Possible values are:

* `<empty>`:  do not report metrics
* `stdout`:   report metrics to stdout
* `graphite`: report metrics to Graphite on [metrics.graphite.addr](/ref/metrics.graphite.addr/)
* `statsd`: legacy statsd support, used in v1.5.5 and lower - removed in v1.6
* `statsd_raw`: report metrics to StatsD on [metrics.statsd.addr](/ref/metrics.statsd.addr/) - this was 
  intentionally renamed because anyone upgrading to 1.6 will need to revisit their configuration anyway due to 
  rewrite of this backend.  It was quite broken before, the counters never reset, it did not follow the spec so the info was 
  likely wrong or people using this were doing some workarounds they'll need to remove anyway.
* `circonus`: report metrics to Circonus (https://circonus.com/)
* `prometheus`: use prometheus metrics. (https://prometheus.io)  Must be used in conjuction with a prometheus 
  listener in [proxy.addr](/ref/proxy.addr/)
* `dogstatsd`: use with datadog dogstatsd (https://www.datadoghq.com/)

The default is

	metrics.target =

Multiple metrics targets can be defined separated by comma.
