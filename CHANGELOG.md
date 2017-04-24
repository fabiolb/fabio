## Changelog

### Unreleased

#### Improvements

 * The listener code no longer swallows the errors and exits if it cannot create
   a listening socket.

### [v1.4.3](https://github.com/fabiolb/fabio/releases/tag/v1.4.3) - 24 Apr 2017

#### Bug Fixes

 * [Issue #269](https://github.com/fabiolb/fabio/issue/269): Access log cannot be disabled

   The access logging feature that was added in v1.4.1 did not allow to disable the access logging
   output and all fabio instances were writing an access log by default. Also, the logging setup
   code would leave fabio registered in consul in case of a failure.

#### Improvements

 * [PR #268](https://github.com/fabiolb/fabio/pull/268): Add support for TLSSkipVerify for https consul fabio check

   When the fabio admin port is configured to use HTTPS then the consul health check has
   to use HTTPS as well. The new `registry.consul.register.checkTLSSkipVerify` option allows
   to disable TLS certificate validation for this check. This requires consul 0.7.2 or higher.

   Thanks to @Ginja for providing this patch.

 * Demo server supports HTTPS
 
   The `demo/server/server` now supports `https` and `wss` to test the 
   HTTPS upstream support. To run an HTTPS server run the following

   ```shell
   # generate some test certs
   cd $GOPATH/src/github.com/fabiolb/fabio
   build/issue-225-gen-cert.bash

   # build and run the demo server
   cd demo/server
   go build
   ./server -certFile ../cert/server/server-cert.pem -keyFile ../cert/server/server-key.pem -proto https -prefix "/foo tlsskipverify=true"
   ```
   
 * Add route options to UI
 
   The UI now shows the combined options from all targets for a route.

 * Add fabio logo to UI
 
   The Fabio logo is displayed on all UI pages.

### [v1.4.2](https://github.com/fabiolb/fabio/releases/tag/v1.4.2) - 10 Apr 2017

The vault tests do not yet pass with vault 0.7.0 and support for vault 0.7.0 has yet to be confirmed.
fabio is known to work with vault 0.6.4.

#### Features

 * [PR #257](https://github.com/fabiolb/fabio/pull/257), [Issue #181](https://github.com/eBay/fabio/issues/181): Add HTTPS Upstream Support

   Upstream servers can now be served via HTTPS. To enable this for a route add the `proto=https` option
   to the `urlprefix-` tag. The upstream certificate needs to be in the system certificate chain for the
   certificate validation to succeed. To disable certificate validation for upstream requests add the
   `tlsskipverify=true` option. Support for certificate stores for upstream servers may come at a later
   point.

   Thanks to @shadowfax-chc for providing this patch.

   See: https://github.com/fabiolb/fabio/wiki/Features#https-upstream-support

 * [PR #258](https://github.com/fabiolb/fabio/pull/258): Allow UI/API to be served over HTTPS

   The UI/API endpoint can now be served via HTTPS. To enable this configure the `ui.addr` property
   with a `cs=<cert store>` option like the `proxy.addr` listeners.

   Thanks to @shadowfax-chc for providing this patch.

#### Improvements

 * Upgrade to go1.8.1
 * Run tests with consul 0.8.0
 * Improve CHANGELOG

### [v1.4.1](https://github.com/fabiolb/fabio/releases/tag/v1.4.1) - 4 Apr 2017

#### Features

 * [Issue #80](https://github.com/fabiolb/fabio/issues/80): Add support for access logging

   fabio now supports configurable access logging. By default access logging is disabled and can
   be enabled with `log.access.target=stdout`. The default format is the
   [Common Log Format](https://en.wikipedia.org/wiki/Common_Log_Format) but can be changed
   to either the [Combined Log Format](https://httpd.apache.org/docs/1.3/logs.html#combined)
   or a custom log format by setting `log.access.format`

   Thanks to @beyondblog for providing the initial patch.

   See: https://github.com/fabiolb/fabio/wiki/Features#access-logging

### [v1.4](https://github.com/fabiolb/fabio/releases/tag/v1.4) - 25 Mar 2017

#### Features

 * [Issue #1](https://github.com/fabiolb/fabio/issues/1), [Issue #179](https://github.com/eBay/fabio/issues/179): Add generic TCP Proxy support

   fabio now supports raw TCP proxying support by setting the `proto=tcp` option on the
   `urlprefix-` tag. The target needs to be the external port of the service, e.g.
   `urlprefix-:3306` for a MySQL proxy. fabio needs to have a TCP listener configured for
   that port through the `proxy.addr` option, e.g. `proxy.addr=:3306;proto=tcp`.

   The TCP proxy also supports TLS which is configured through the `cs=<cert store>`
   option like the HTTPS listeners.

 * [Issue #163](https://github.com/fabiolb/fabio/issues/163): Support glob host matching

   This patch adds support for glob host matching the hostname in routes like
   `urlprefix-*.foo.com/bar`.

#### Improvements

 * Upgrade to Go 1.8 and drop support for Go 1.7

 * [Issue #178](https://github.com/fabiolb/fabio/issues/178): Add tests and timeouts to TCP+SNI proxy

   Add full integration tests and support for read/write timeouts through the `rt=` and `wt=`
   options on the listener config for the TCP+SNI proxy. The initial implementation was only
   tested manually.

 * [Issue #248](https://github.com/fabiolb/fabio/issues/248): Start listener after routing table is initialized

   fabio now waits for the first routing table before serving requests. This should remove
   503s during restarts on heavily loaded sites.

### [v1.3.8](https://github.com/fabiolb/fabio/releases/tag/v1.3.8) - 14 Feb 2017

#### Features

 * [Issue #219](https://github.com/fabiolb/fabio/issues/219): Support absolute URLs

#### Improvements

 * Upgrade to Go 1.7.5
 * [Issue #238](https://github.com/fabiolb/fabio/issues/238): Make route update logging format configurable. Log delta by default
 * [Issue #240](https://github.com/fabiolb/fabio/issues/240): Retry registry during startup

### [v1.3.7](https://github.com/fabiolb/fabio/releases/tag/v1.3.7) - 19 Jan 2017

#### Features

 * [Issue #44, #124, #164](https://github.com/fabiolb/fabio/issues/44, https://github.com/eBay/fabio/issues/124, https://github.com/eBay/fabio/issues/164): Support path stripping
 * [Issue #201](https://github.com/fabiolb/fabio/issues/201): Support deleting routes by tag

#### Bug Fixes

 * [Issue #207](https://github.com/fabiolb/fabio/issues/207): Bad statsd mean metric format
 * [Issue #217](https://github.com/fabiolb/fabio/issues/217): fabio 1.3.6 UI displays host and path as 'undefined' in the routes page
 * [Issue #218](https://github.com/fabiolb/fabio/issues/218): requests and notfound metric missing

### [v1.3.6](https://github.com/fabiolb/fabio/releases/tag/v1.3.6) - 17 Jan 2017

#### Improvements

 * Upgrade to Go 1.7.4
 * [Issue #111](https://github.com/fabiolb/fabio/issues/111): Refactor urlprefix tags (step 1: options and new parser)
 * [Issue #186](https://github.com/fabiolb/fabio/issues/186): runtime error: integer divide by zero
 * [Issue #199](https://github.com/fabiolb/fabio/issues/199): Refactor config loader tests
 * [Issue #215](https://github.com/fabiolb/fabio/issues/215): Re-enable HTTP/2 support

### [v1.3.5](https://github.com/fabiolb/fabio/releases/tag/v1.3.5) - 30 Nov 2016

#### Improvements

 * [Issue #182](https://github.com/fabiolb/fabio/issues/182): Initialize Vault client better
 * [Issue #183](https://github.com/fabiolb/fabio/issues/183): Websocket header casing
 * [Issue #189](https://github.com/fabiolb/fabio/issues/189): missing 'cs' in map
 * [Issue #194](https://github.com/fabiolb/fabio/issues/194): Remove proxy.header.tls header from inbound request
 * [Issue #197](https://github.com/fabiolb/fabio/issues/197): Add support for '--version'

### [v1.3.4](https://github.com/fabiolb/fabio/releases/tag/v1.3.4) - 28 Oct 2016

#### Features

 * [Issue #119](https://github.com/fabiolb/fabio/issues/119): Transparent response body compression

#### Improvements

 * Upgrade to Go 1.7.3

### [v1.3.3](https://github.com/fabiolb/fabio/releases/tag/v1.3.3) - 12 Oct 2016

#### Improvements 

 * Drop support for Go 1.6 since tests now use `t.Run()`
 * [PR #167](https://github.com/fabiolb/fabio/pull/167): Use Go's net.JoinHostPort which will auto-detect ipv6
 * [Issue #177](https://github.com/fabiolb/fabio/issues/177): TCP+SNI proxy does not work with PROXY protocol

#### Bug Fixes

 * [Issue #172](https://github.com/fabiolb/fabio/issues/172): Consul cert store URL with token not parsed correctly

### [v1.3.2](https://github.com/fabiolb/fabio/releases/tag/v1.3.2) - 11 Sep 2016

#### Bug Fixes

 * [Issue #159](https://github.com/fabiolb/fabio/issues/159): Panic on invalid response

### [v1.3.1](https://github.com/fabiolb/fabio/releases/tag/v1.3.1) - 9 Sep 2016

#### Bug Fixes

 * [Issue #157](https://github.com/fabiolb/fabio/issues/157): ParseListen may set the wrong protocol

### [v1.3](https://github.com/fabiolb/fabio/releases/tag/v1.3) - 9 Sep 2016

#### Features

 * [Issue #1](https://github.com/fabiolb/fabio/issues/1): Add TCP proxy with SNI support (EXPERIMENTAL)
 * [Issue #138](https://github.com/fabiolb/fabio/issues/138): Add option to disable cert fallback
 * [Issue #147](https://github.com/fabiolb/fabio/issues/147): Support multiple metrics libraries
 * [Issue #151](https://github.com/fabiolb/fabio/issues/151)/[PR #150](https://github.com/eBay/fabio/pull/150): Add support for Circonus metrics

#### Improvements

 * [Issue #125](https://github.com/fabiolb/fabio/issues/125): Extended metrics
 * [Issue #134](https://github.com/fabiolb/fabio/issues/134): Vault token should not require 'root' or 'sudo' privileges
 * [PR #154](https://github.com/fabiolb/fabio/pull/154): Make route metric names configurable

### [v1.2.1](https://github.com/fabiolb/fabio/releases/tag/v1.2.1) - 25 Aug 2016

#### Features

 * [Issue #73](https://github.com/fabiolb/fabio/pull/73)/[PR #139](https://github.com/eBay/fabio/pull/139): Add statsd support
 * [Issue #129](https://github.com/fabiolb/fabio/issues/129): Server-sent events support

#### Improvements

 * [Issue #136](https://github.com/fabiolb/fabio/issues/136): Always deregister from consul
 * [PR #143](https://github.com/fabiolb/fabio/pull/143): Improve error message on missing trailing slash

#### Bug Fixes

 * [Issue #146](https://github.com/fabiolb/fabio/issues/146): fabio fails to start with "[FATAL] 1.2. missing 'cs' in cs"

### [v1.2](https://github.com/fabiolb/fabio/releases/tag/v1.2) - 16 Jul 2016

#### Features

 * [Issue #27](https://github.com/fabiolb/fabio/issues/27): Change certificates via API
 * [Issue #70](https://github.com/fabiolb/fabio/issues/70): Support Vault
 * [Issue #85](https://github.com/fabiolb/fabio/issues/85): SNI support

#### Improvements

 * [Issue #28](https://github.com/fabiolb/fabio/issues/28): Refactor listener config
 * [Issue #79](https://github.com/fabiolb/fabio/issues/79): Refactor config loading to use flag sets

### [v1.1.6](https://github.com/fabiolb/fabio/releases/tag/v1.1.6) - 12 Jul 2016

#### Bug Fixes

 * [Issue #108](https://github.com/fabiolb/fabio/issues/108): TLS handshake error: failed to verify client's certificate
 * [Issue #122](https://github.com/fabiolb/fabio/issues/122): X-Forwarded-Port should use local port

### [v1.1.5](https://github.com/fabiolb/fabio/releases/tag/v1.1.5) - 23 Jun 2016

#### Improvements

 * [PR #117](https://github.com/fabiolb/fabio/pull/117): Allow routes to a service in warning status

### [v1.1.4](https://github.com/fabiolb/fabio/releases/tag/v1.1.4) - 15 Jun 2016

#### Improvements

 * [Issue #99](https://github.com/fabiolb/fabio/issues/99): Disable fabio health check in consul
 * [Issue #100](https://github.com/fabiolb/fabio/issues/100): Keep fabio registered in consul
 * [Issue #107](https://github.com/fabiolb/fabio/issues/107): Custom status code when no route found

### [v1.1.3](https://github.com/fabiolb/fabio/releases/tag/v1.1.3) - 20 May 2016

#### Features

 * [Issue #95](https://github.com/fabiolb/fabio/issues/95): Expand experimental HTTP API
 * [Issue #97](https://github.com/fabiolb/fabio/issues/97): Support PROXY protocol
 * [PR #93](https://github.com/fabiolb/fabio/pull/93): Add glob path matching

#### Improvements

 * Drop support for Go 1.5
 * [Issue #55](https://github.com/fabiolb/fabio/issues/55): Expand ${DC} to consul datacenter
 * [Issue #96](https://github.com/fabiolb/fabio/issues/96): Allow tags for fabio service registration
 * [Issue #98](https://github.com/fabiolb/fabio/issues/98): Improve forward header
 * [Issue #103](https://github.com/fabiolb/fabio/issues/103): Trim whitespace around tag
 * [Issue #104](https://github.com/fabiolb/fabio/issues/104): Keep sort order in UI stable

### [v1.1.2](https://github.com/fabiolb/fabio/releases/tag/v1.1.2) - 27 Apr 2016

#### Improvements

 * Upgrade to Go 1.5.4 and Go 1.6.2
 * [PR #74](https://github.com/fabiolb/fabio/pull/74): Improve forward header handling
 * [Issue #77](https://github.com/fabiolb/fabio/issues/77): Fix registry.consul.register.addr example in properties
 * [Issue #88](https://github.com/fabiolb/fabio/issues/88): Use consul node address
 * [Issue #90](https://github.com/fabiolb/fabio/issues/90): Drop default port from request

### [v1.1.1](https://github.com/fabiolb/fabio/releases/tag/v1.1.1) - 22 Feb 2016

#### Improvements

 * [Issue #57](https://github.com/fabiolb/fabio/issues/57): Deleted routes hide visible routes
 * [Issue #59](https://github.com/fabiolb/fabio/issues/59): Latest fabio docker image fails consul check
 * [PR #58](https://github.com/fabiolb/fabio/pull/58): Fix use of local ip in consul service registration

### [v1.1](https://github.com/fabiolb/fabio/releases/tag/v1.1) - 18 Feb 2016

#### Features

 * [Issue #12](https://github.com/fabiolb/fabio/issues/12): Support additional backends
 * [Issue #32](https://github.com/fabiolb/fabio/issues/32): HTTP2 support with latest Go
 * [Issue #43](https://github.com/fabiolb/fabio/issues/43): Allow configuration via env vars

#### Improvements

 * Drop support for Go 1.4 and build for both Go 1.5.3 and Go 1.6
 * [Issue #37](https://github.com/fabiolb/fabio/issues/37): Add support for consul ACL token to demo server
 * [Issue #41](https://github.com/fabiolb/fabio/issues/41): Cleanup metrics for deleted routes
 * [Issue #47](https://github.com/fabiolb/fabio/issues/47): Move dependencies to vendor path
 * [Issue #48](https://github.com/fabiolb/fabio/issues/48): Allow configuration of serviceip used during consul registration
 * [PR #49](https://github.com/fabiolb/fabio/pull/49): Fix up use of addr in service registration

### [v1.0.9](https://github.com/fabiolb/fabio/releases/branch/v1.0.9) - 16 Jan 2016

#### Improvements

 * [Issue #53](https://github.com/fabiolb/fabio/issues/53): Make read and write timeout configurable

### [v1.0.8](https://github.com/fabiolb/fabio/releases/tag/v1.0.8) - 14 Jan 2016

#### Features

 * [Issue #36](https://github.com/fabiolb/fabio/issues/36): Add support for consul ACL token

#### Improvements

 * Upgrade to Go 1.5.3
 * [Issue #29](https://github.com/fabiolb/fabio/issues/29): Include service with check ids other than 'service:*'
 * [Issue #30](https://github.com/fabiolb/fabio/issues/30): Register fabio with local ip address as fallback

### [v1.0.7](https://github.com/fabiolb/fabio/releases/tag/v1.0.7) - 13 Dec 2015

#### Improvements

 * [Issue #22](https://github.com/fabiolb/fabio/issues/22): fabio route not removed after consul deregister
 * [Issue #23](https://github.com/fabiolb/fabio/issues/23): routes not removed when passing empty string
 * [Issue #26](https://github.com/fabiolb/fabio/issues/26): Detect when consul agent is down
 * Allow to override title and color UI

### [v1.0.6](https://github.com/fabiolb/fabio/releases/tag/v1.0.6) - 01 Dec 2015

#### Improvements

 * [Issue #9](https://github.com/fabiolb/fabio/issues/9): Enabled raw websocket proxy by default
 * [Issue #15](https://github.com/fabiolb/fabio/issues/15): Traffic shaping now matches on service
 * [Issue #16](https://github.com/fabiolb/fabio/issues/16): Improved Web UI with better filtering
 * [Issue #18](https://github.com/fabiolb/fabio/issues/18): Manage manual overrides via ui

### [v1.0.5](https://github.com/fabiolb/fabio/releases/tag/v1.0.5) - 11 Nov 2015

#### Features

 * [Issue #9](https://github.com/fabiolb/fabio/issues/9): Add experimental support for web sockets
 * [Issue #10](https://github.com/fabiolb/fabio/issues/10): Add support for `Forwarded` and `X-Forwarded-For` header

#### Improvements

 * Add `proxy.localip` to set proxy ip address for headers

### [v1.0.4](https://github.com/fabiolb/fabio/releases/tag/v1.0.4) - 03 Nov 2015

#### Features

 * [Issue #8](https://github.com/fabiolb/fabio/issues/8): Add support for SSL client certificate authentication

### [v1.0.3](https://github.com/fabiolb/fabio/releases/tag/v1.0.3) - 25 Oct 2015

#### Improvements

 * Add Docker support and official Docker image `magiconair/fabio`

### [v1.0.2](https://github.com/fabiolb/fabio/releases/tag/v1.0.2) - 23 Oct 2015

#### Improvements

 * [PR #3](https://github.com/fabiolb/fabio/pull/3): Honor consul.url and consul.addr from config file (@jeinwag)

### [v1.0.1](https://github.com/fabiolb/fabio/releases/tag/v1.0.1) - 21 Oct 2015

#### Improvements

 * Honor maintenance mode for both services and nodes

### [v1.0.0](https://github.com/fabiolb/fabio/releases/tag/v1.0.0) - 16 Oct 2015

 * Initial open-source release

