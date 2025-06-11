---
title: "metrics.prometheus.buckets"
---

`metrics.prometheus.buckets` configures the time buckets for use with histograms, measured in seconds.
for instance, .005 is equivalent to 5ms.  There is an implied "infinity" bucket tacked on at the end.

The default is
`metrics.prometheus.buckets = .005,.01,.025,.05,.1,.25,.5,1,2.5,5,10`
