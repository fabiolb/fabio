## Intro
This library is based on the cyberdelia [graphite library](https://github.com/cyberdelia/go-metrics-graphite) just modifying the wireformat for statsd

This is a reporter for the [go-metrics](https://github.com/rcrowley/go-metrics)
library which will post the metrics to Statsd. It was originally part of the
`go-metrics` library itself, but has been split off to make maintenance of
both the core library and the client easier.

### Usage

```go
import "github.com/pubnub/go-metrics-statsd"


go statsd.StatsD(metrics.DefaultRegistry,
  1*time.Second, "some.prefix", addr)
```

