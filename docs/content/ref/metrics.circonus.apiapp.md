---
title: "metrics.circonus.apiapp"
---

`metrics.circonus.apiapp` configures the API token app to use when
submitting metrics to Circonus. See: https://login.circonus.com/user/tokens
This is optional when [metrics.target](/ref/metrics.target/) is set to `circonus`.

The default is

	metrics.circonus.apiapp = fabio
