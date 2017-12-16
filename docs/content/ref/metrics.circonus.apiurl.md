---
title: "metrics.circonus.apiurl"
---

`metrics.circonus.apiurl` configures the API URL to use when
submitting metrics to Circonus. https://api.circonus.com/v2/
will be used if no specific URL is provided.
This is optional when [metrics.target](/ref/metrics.target/) is set to `circonus`.

The default is

	metrics.circonus.apiurl =
