---
title: "metrics.circonus.checkid"
---

`metrics.circonus.checkid` configures a specific check to use when
submitting metrics to Circonus.

This is optional when [metrics.target](/ref/metrics.target/) is set to `circonus`.

An attempt will be made to search for a previously created check,
if no applicable check is found, one will be created.

The default is

	metrics.circonus.checkid =
