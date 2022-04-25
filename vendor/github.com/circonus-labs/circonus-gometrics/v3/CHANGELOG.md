# v3.4.6

* upd: more stringent tag limit enforcement

# v3.4.5

* upd: set metric ts only when submit timestamp set by caller
* upd: dependencies
* fix: lint struct field alignment

# v3.4.4

* add: setting submission timestamp
* add: support timestamps on histograms with SerializeB64

# v3.4.3

* add: submit duration to results
* add: more per req logging on submit failures
* upd: optimize struct layout/size chkmgr

# v3.4.2

* upd: dependency (go-apiclient)
* upd: clean up/normalize http client for httptrap requests
* add: output full response from broker stats, filtered, and error if applicable in debug message
* add: ability to send to clustered brokers behind LB which are not using a cert for a single CN
* add: `SerialInit` option to initialize check serially
* add: `GetBrokerTLSConfig` to retrieve broker tls configuration
* add: `GetCheckBundle` for caching

# v3.4.1

* add: `DumpMetrics` config option, log payload sent to broker for debugging
* upd: dependencies

# v3.4.0

* fix: service tag file path
* fix: lint issues
* add: lint workflow
* upd: dependencies (go-apiclient,circonusllhst)

# v3.3.4

* upd: dependencies (cgm,circonusllhst,retryablehttp)

# v3.3.3

* add: custom metric support

# v3.3.2

* fix: put locking in snapshot handlers

# v3.3.1

* add: concurrent snapshot

# v3.3.0

* upd: add timestamps to metric output

# v3.2.2

* upd: add debug log when fetching cert from api
* fix: update internal cert copy for fallback
* upd: use manageMetrics to get check metric management
* add: manageMetrics state for non-metric filter checks

# v3.2.1

* fix: internal log interface satisfies rh log interface no need to cast to log.Logger any longer

# v3.2.0

* add: accept category only tags
* fix: tls config for self-signed certs (go1.15)

# v3.1.2

* fix: identify httptrap with subtype for submission url

# v3.1.1

* fix: quoting on error message in test

# v3.1.0

* upd: do not force tag values to lowercase

# v3.0.2

* add: method to flush metrics without resetting  (`FlushMetricsNoReset()`)

# v3.0.1

* upd: dependencies
* fix: send empty array for `check_bundle.metrics`, api errors on null now

# v3.0.0

* upd: stricter linting
* upd: dependencies
* upd: api submodule is deprecated (use github.com/circonus-labs/go-apiclient or older v2 branch of circonus-gometrics)

# v3.0.0-beta.4

* fix: verify at least one active check found when searching for checks
* upd: broker test IP and external host for match against submission url host

# v3.0.0-beta.3

* upd: go-apiclient for graph overlay attribute type fixes

# v3.0.0-beta.2

* fix: submit for breaking change in dependency patch release
* upd: dependencies

# v3.0.0-beta.1

* upd: merge tag helper methods, support logging invalid tags
* upd: allow manually formatted and base64 encoded tags
* upd: allow tag values to have embedded colons

# v3.0.0-beta

* add: log deprecation notice on api calls
* upd: dependency circonusllhist v0.1.2, go-apiclient v0.5.3
* upd: `snapHistograms()` method to use the histogram `Copy()` if `resetHistograms` is false, otherwise uses `CopyAndReset()`

# v3.0.0-alpha.5

* add: allow any log package with a `Printf` to be used
* upd: circonus-labs/go-apiclient v0.5.2 (for generic log support)
* upd: ensure only `Printf` is used for logging
* upd: migrate to errors package (`errors.Wrap` et al.)
* upd: error and log messages, remove explicit log level classifications from logging messages
* upd: OBSOLETE github.com/circonus-labs/v3/circonus-gometrics/api will be REMOVED --- USE **github.com/circonus-labs/go-apiclient**

# v3.0.0-alpha.4

* add: missing SetHistogramDurationWithTags
* upd: go-apiclient v0.5.1
* fix: remove cgm v2 dependency from DEPRECATED api package
* upd: retryablehttp v0.5.0

# v3.0.0-alpha.3

* add: RecordDuration, RecordDurationWithTags, SetHistogramDuration

# v3.0.0-alpha.2

* upd: circllhist v0.1.2

# v3.0.0-alpha.1

* fix: enable check management for add tags test
* fix: api.circonus.com hostname (accidentally changed during switch to apiclient)

# v3.0.0-alpha

* add: helper functions for metrics `*WithTags` e.g. `TimingWithTags(metricName,tagList,val)`
* upd: default new checks to use metric_filters
* add: metric_filters support
* upd: dependencies (circonusllhist v0.1.0)
* upd: change histograms from type 'n' to type 'h' in submissions
* upd: DEPRECATED github.com/circonus-labs/v3/circonus-gometrics/api
* upd: switch to using github.com/circonus-labs/go-apiclient
* upd: merge other metric tag functions into tags
* add: helper methods for handling tags (for new stream tags syntax and old check_bundle.metrics.metric.tags)
* upd: merge other metric output functions into metric_output
* upd: merge util into metric_output (methods in util are specifically for working with metric outputs)
