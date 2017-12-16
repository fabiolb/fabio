---
title: "metrics.circonus.brokerid"
---

`metrics.circonus.brokerid` configures a specific broker to use when
creating a check for submitting metrics to Circonus.

This is optional when [metrics.target](/ref/metrics.target/) is set to `circonus`.

Optional for public brokers, required for Inside brokers.
Only applicable if a check is being created.

The default is

	metrics.circonus.brokerid =
