# v0.7.15

* fix: do not allow blank tags through on check bundle creation

# v0.7.14

* add: `MaxRetries`, `MinRetryDelay`, and `MaxRetryDelay` settings

# v0.7.13

* upd: dependencies

# v0.7.12

* fix: lint issues
* add: lint config and workflow
* upd: squelch empty data debug msg

# v0.7.11

* add: new `user_json` field support to rule_set
* upd: make timeout/retry tests optional (env var)

# v0.7.10

* upd: add 429 rate limit tests
* upd: dependency retryablehttp, to use Retry-After header on 429s
* upd: increase exp backoff range 1-60

# v0.7.9

* add: additional SMTP check attributes to support proxies

# v0.7.8

* add: `lookup_key` to rule_set

# v0.7.7

* fix: lint simplifications in tests
* upd: dependency
* add: WindowingMinDuration to RuleSetRule
* fix: remove derive from tests (deprecated)
* upd: remove RuleSet.Derive (deprecated)
* upd: remove Tags and Units from metrics (deprecated)

# v0.7.6

* fix: skip backoff for HTTP/400
* fix: change `Dashboard.Settings.ShowValue` to `*bool` to facilitate intentional `false` not being omitted

# v0.7.5

* fix: break, return error on 404 result with exponential backoff

# v0.7.4

* fix: `metric_type` field on dashboard state widget

# v0.7.3

* add: state widget to dashboard

# v0.7.2

* fix: `/rule_set_group` formulas, raise_severity api bug; mixed types - POST takes an int and returns a string. GET returns an int

# v0.7.1

* fix: typo in rule_set_group matching_severities

# v0.7.0

* fix: forecast gauge `flip` required, remove omitempty
* upd: range hi/low switched from `int` to `*int` so that 0 can be used, but common setting attribute still omitted for widgets which do not support the range settings

# v0.6.9

* fix: contact_group.`alert_formats`, individual fields should be omitted if not set (was `string|null`, now `string|omit`)
* add: contact_group.`always_send_clear` attribute, bool
* add: contact_group.`group_type` attribute, string

# v0.6.8

* upd: force logging of json being sent to api

# v0.6.7

* add: new rule_set attributes `_host`, `filter`, `metric_pattern`, and `name`.
* upd: go1.13

# v0.6.6

* fix: typo on struct attr 'omitempt'

# v0.6.5

* upd: dependencies
* upd: stricter linting
* add: `_reverse_urls` attribute to check object

# v0.6.4

* fix: graph.datapoint.alpha - doc:floating point number, api:string

# v0.6.3

* upd: remove tests for invalid cids
* fix: validate cids on prefix only to compensate for breaking change to rule_set cid in public v2 api

# v0.6.2

* upd: dependencies (retryablehttp)

# v0.6.1

* add: full overlay test suite to `examples/graph/overlays`
* fix: incorrect attribute types in graph overlays (docs vs what api actually returns)

# v0.6.0

* fix: graph structures incorrectly represented nesting of overlay sets

# v0.5.4

* add: `search` (`*string`) attribute to graph datapoint
* upd: `cluster_ip` (`*string`) can be string OR null
* add: `cluster_ip` attribute to broker details

# v0.5.3

* upd: use std log for retryablehttp until dependency releases Logger interface

# v0.5.2

* upd: support any logging package with a `Printf` method via `Logger` interface rather than forcing `log.Logger` from standard log package
* upd: remove explicit log level classifications from logging messages
* upd: switch to errors package (for `errors.Wrap` et al.)
* upd: clarify error messages
* upd: refactor tests
* fix: `SearchCheckBundles` to use `*SearchFilterType` as its second argument
* fix: remove `NewAlert` - not applicable, alerts are not created via the API
* add: ensure all `Delete*ByCID` methods have CID corrections so short CIDs can be passed

# v0.5.1

* upd: retryablehttp to start using versions that are now available instead of tracking master

# v0.5.0

* Initial - promoted from github.com/circonus-labs/circonus-gometrics/api to an independant package
