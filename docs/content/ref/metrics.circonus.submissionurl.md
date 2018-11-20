---
title: "metrics.circonus.submissionurl"
---

`metrics.circonus.submissionurl` configures a specific check submission url
for a Check API object of a previously created HTTPTRAP check.

This is optional when [metrics.target](/ref/metrics.target/) is set to `circonus`
but [metrics.circonus.apikey](/ref/metrics.circonus.apikey/) is specified}.

#### Example

`http://127.0.0.1:2609/write/fabio`

The default is

	metrics.circonus.submissionurl =
