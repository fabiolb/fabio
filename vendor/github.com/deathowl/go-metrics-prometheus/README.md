**go-metrics-prometheus**
[![Build Status](https://api.travis-ci.org/deathowl/go-metrics-prometheus.svg)](https://travis-ci.org/deathowl/go-metrics-prometheus)

This is a reporter for the go-metrics library which will post the metrics to the prometheus client registry . It just updates the registry, taking care of exporting the metrics is still your responsibility.


Usage:

```

	import "github.com/deathowl/go-metrics-prometheus"
	import "github.com/prometheus/client_golang/prometheus"

        metricsRegistry := metrics.NewRegistry()
	prometheusClient := prometheusmetrics.NewPrometheusProvider(
	   metrics.DefaultRegistry, "whatever","something",prometheus.DefaultRegisterer, 1*time.Second)
        go prometheusClient.UpdatePrometheusMetrics()
```

