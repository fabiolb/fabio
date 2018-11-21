---
title: "Metrics"
since: "1.0.0 (Graphite), 1.2.1 (StatsD, DataDog, Circonus)"
---

fabio collects metrics per route and service instance as well as running totals
to avoid computing large amounts of metrics. The metrics can be send to
[Circonus](http://www.circonus.com), [Graphite](https://graphiteapp.org),
[StatsD](https://github.com/etsy/statsd), [DataDog](https://www.datadoghq.com)
(via statsd) or stdout. See the `metrics.*` options in the
[fabio.properties](https://github.com/eBay/fabio/blob/master/fabio.properties)
file.

Fabio reports the following metrics:

Name                        | Type     | Description
--------------------------- | -------- | -------------
`{route}.rx`                | timer    | Number of bytes received by fabio for TCP target
`{route}.tx`                | timer    | Number of bytes transmitted by fabio for TCP target
`{route}`                   | timer    | Average response time for a route
`http.status.code.{code}`   | timer    | Average response time for all HTTP(S) requests per status code
`notfound`                  | counter  | Number of failed HTTP route lookups
`requests`                  | timer    | Average response time for all HTTP(S) requests
`grpc.requests`             | timer    | Average response time for all GRPC(S) requests
`grpc.noroute`              | counter  | Number of failed GRPC route lookups
`grpc.conn`                 | counter  | Number of established GRPC proxy connections
`grpc.status.{code}`        | timer    | Average response time for all GRPC(S) requests per status code
`tcp.conn`                  | counter  | Number of established TCP proxy connections
`tcp.connfail`              | counter  | Number of TCP upstream connection failures
`tcp.noroute`               | counter  | Number of failed TCP upstream route lookups
`tcp_sni.conn`              | counter  | Number of established TCP+SNI proxy connections
`tcp_sni.connfail`          | counter  | Number of failed TCP+SNI proxy connections
`tcp_sni.noroute`           | counter  | Number of failed TCP+SNI upstream route lookups
`ws.conn`                   | gauge    | Number of actively open websocket connections


### Legend

#### timer

A timer counts events and provides an average throughput and latency number.
Depending on the metrics provider the aggregation happens either in the metrics library
(go-metrics: statsd, graphite) or in the system of the metrics provider (Circonus)

#### counter

A counter counts events and provides an monotonically increasing value.

#### gauge

A gauge provides a current value.

#### {code}

`{code}` is the three digit HTTP status code like `200`.

#### {route}

`{route}` is a shorthand for the metrics name generated for a route
with the `metrics.names` template defined in
[fabio.properties](https://github.com/fabiolb/fabio/blob/master/fabio.properties)

