## Changelog

### [v1.5.11](https://github.com/fabiolb/fabio/releases/tag/v1.5.11) - 25 Feb 2019

#### Breaking Changes

* [Issue #524](https://github.com/fabiolb/fabio/issues/524): Fix TCP proxy

  Adding access control in 1.5.8 broke the TCP proxy for clients which wait
  for a server handshake before they send data themself, e.g. MySQL client.
  Calling `RemoteAddr` to check the access list is a blocking call on the
  underlying listener. The Go team fixed this for HTTP in
  https://github.com/golang/go/issues/12943 but fabio was still affected.

  This patch **disables** the PROXY protocol by default and you have to
  **enable** it if you need it. You can control this with the `pxyproto` and
  `pxytimeout` options which allow to enable/disable the protocol and add a
  timeout for auto-detection if it is enabled.

  Thanks to [@pschultz](https://github.com/pschultz) and
  [@leprechau](https://github.com/leprechau) for the debugging and patch.

#### Bug Fixes

* [PR #577](https://github.com/fabiolb/fabio/pull/577): Fix ip access rules within tcp proxy

  Access rules were not being evaluated in the `tcp` proxy.

  Thanks to [@KEZHwMlXV1vFzs6QvY8v5WjX5](https://github.com/KEZHwMlXV1vFzs6QvY8v5WjX5) for identifying the issue, providing a solution, and testing.

* [PR #588](https://github.com/fabiolb/fabio/pull/588): Fix XSS vulnerability in UI

  Thanks to [@pschultz](https://github.com/pschultz).

#### Improvements

* [PR #564](https://github.com/fabiolb/fabio/pull/564): refactor consul service monitor

  This patch addresses [Issue #558](https://github.com/fabiolb/fabio/issues/558) where
  route updates are delayed when there are a large number of services registered in
  Consul. This patch adds a new option `registry.consul.serviceMonitors` to configure
  the concurrency for routing table updates.

  Thanks to [@galen0624](https://github.com/galen0624) for the issue report and the initial
  patch.

* [PR #574](https://github.com/fabiolb/fabio/pull/574): add support for circonus.submissionurl

  Thanks to [@stack72](https://github.com/stack72) for the patch.

* [PR #583](https://github.com/fabiolb/fabio/pull/583): Make dest column clickable

  This patch updates the documentation around the PROXY protocol.

  Thanks to [@pschultz](https://github.com/pschultz).

* [PR #587](https://github.com/fabiolb/fabio/pull/587): Make dest column clickable

  This patch makes the dest column in the fabio UI clickable.

  Thanks to [@kneufeld](https://github.com/kneufeld).

#### Features

* [PR #429](https://github.com/fabiolb/fabio/issues/429): Support for opentracing

   This patch adds support for opentracing.

   Thanks to Jeremy White, Kristina Fischer, Micheal Murphz, Nathan West,
   Austin Hartzheim and Jacob Hansen for this patch!

* [PR #553](https://github.com/fabiolb/fabio/issues/553): Support for case-insensitive matching

  This patch adds a new `iprefix` option to the `proxy.matcher` to support case-insensitive
  path prefix matching.

  Thanks to [@herbrandson](https://github.com/herbrandson) for the patch.

* [PR #573](https://github.com/fabiolb/fabio/pull/573): add http-basic auth reading from a file

  This patch adds support for HTTP Basic Authentication via an `htpasswd` file.

  Thanks to [@andyroyle](https://github.com/andyroyle) for the patch.

* [PR #575](https://github.com/fabiolb/fabio/pull/575): Add GRPC proxy support

  This patch adds proper GRPC proxy support, including TLS upstream and TLS termination.

  Thanks to [@andyroyle](https://github.com/andyroyle) for the patch.

### [v1.5.10](https://github.com/fabiolb/fabio/releases/tag/v1.5.10) - 25 Oct 2018

#### Breaking Changes

#### Bug Fixes

 * [Issue #530](https://github.com/fabiolb/fabio/issues/530): Memory leak in go-metrics library

   When metrics collection was enabled within fabio instances with very dynamic route changes memory usage quickly
   ramped above expected levels.  Research done by [@galen0624](https://github.com/galen0624) identified the issue
   and lead to the discovery of a fix in an updated version of the go-metrics library used by fabio.

 * [Issue #506](https://github.com/fabiolb/fabio/issues/506): Wrong route for multiple matching host glob patterns

   When multiple host glob patterns match an incoming request fabio can pick the wrong backend for the request.
   This is because the sorting code that should sort the matching patterns from most specific to least specific
   does not take into account that doamin names have their most specific part at the front. This has been fixed
   by reversing the domain names before sorting.

#### Improvements

 * The default Docker image is now based on alpine:3.8 and runs the full test suite during build. It also sets
   `/usr/bin/fabio` as `ENTRYPOINT` with `-cfg /etc/fabio/fabio.properties` as default command line arguments.
   The previous image was built on `scratch`.

 * [PR #497](https://github.com/fabiolb/fabio/pull/497): Make tests pass with latest Consul and Vault versions

   Thanks to [@pschultz](https://github.com/pschultz) for the patch.

 * [PR #531](https://github.com/fabiolb/fabio/pull/531): Set flush buffer interval for non-SSE requests

   This PR adds a `proxy.globalflushinterval` option to configure an interval when the HTTP Response
   Buffer is flushed.

   Thanks to [@samm-git](https://github.com/samm-git) for the patch.

 * [Issue #542](https://github.com/fabiolb/fabio/issues/542): Ignore host case when adding and matching routes

  Fabio was forcing hostnames in routes added via Consul tags to lowercase.  This caused problems
  with table lookups that were not case-insensitive.  The patch applied in #543 forces all routes added
  via consul tags or the internal `addRoute` to have lower case hostnames in addition to forcing
  hostnames to lowercase before performing table lookups.  This means that the host portion of routes and
  host based table lookups in fabio are no longer case sensitive.

  Thanks to [@shantanugadgil](https://github.com/shantanugadgil) for the patch.

 * [Issue #548](https://github.com/fabiolb/fabio/issues/548): Slow glob matching with large number of services

   This patch adds the new `glob.matching.disabled` option which controls whether glob matching is enabled for
   route lookups. If the number of routes is large then the glob matching can have a performance impact and
   disabling it may help.

  Thanks to [@galen0624](https://github.com/galen0624) for the patch and
  [@leprechau](https://github.com/leprechau) for the review.

#### Features

 * [Issue #544](https://github.com/fabiolb/fabio/issues/544): Add $host pseudo variable

  This PR added support for `$host` pseudo variable that behaves similarly to the `$path` variable.
  You should now be able to create a global redirect for requests received on any host to the same or different
  request host on the same or different path when combined with the `$path` variable.  This allows for a truly global
  protocol redirect of HTTP -> HTTPS traffic irrespective of host and path.

  Thanks to [@holtwilkins](https://github.com/holtwilkins) for the patch.

### [v1.5.9](https://github.com/fabiolb/fabio/releases/tag/v1.5.9) - 16 May 2018

#### Notes

 * [Issue #494](https://github.com/fabiolb/fabio/issues/494): Tests fail with Vault > 0.9.6 and Consul > 1.0.6

   Needs more investigation.

#### Breaking Changes

 * None

#### Bug Fixes

 * [Issue #460](https://github.com/fabiolb/fabio/issues/460): Fix access logging when gzip is enabled

   Fabio was not writing access logs when the gzip compression was enabled.

   Thanks to [@tino](https://github.com/tino) for finding this and providing
   and initial patch.

 * [PR #468](https://github.com/fabiolb/fabio/pull/468): Fix the regex of the example proxy.gzip.contenttype

   The example regexp for `proxy.gzip.contenttype` in `fabio.properties` was not properly escaped.

   Thanks to [@tino](https://github.com/tino) for the patch.

 * [Issue #421](https://github.com/fabiolb/fabio/issues/421): Fabio routing to wrong backend

   Fabio does not close websocket connections if the connection upgrade fails. This can lead to
   connections being routed to the wrong backend if there is another HTTP router like nginx in
   front of fabio. The failed websocket connection creates a direct TCP tunnel to the original
   backend server and that connection is not closed properly.

   The patches detect an unsuccessful handshake and close the connection properly.

   Thanks to [@craigday](https://github.com/craigday) for the original reporting and debugging.

#### Improvements

 * [Issue #427](https://github.com/fabiolb/fabio/issues/427): Fabio does not remove service when one of the registered health-checks fail

   If a service has more than one health check then the behavior in whether the
   service is available differs between Consul and Fabio. Consul requires that
   all health checks for a service need to pass in order to return a positive
   DNS result. Fabio requires only one of the health checks to pass.

   A new config option `registry.consul.checksRequired` has been added which
   defaults to the current fabio behavior of `one` passing health check for the
   service to be added to the routing table. To make fabio behave like Consul
   you can set the option to `all`.

   Fabio will make `all` the default as of version 1.6.

   Thanks to [@systemfreund](https://github.com/systemfreund) for the patch.

 * [Issue #448](https://github.com/fabiolb/fabio/issues/448): Redirect http to https on the same destination

   Fabio will now handle redirecting from http to https on the same destination
   without a redirect loop.

   Thanks to [@leprechau](https://github.com/leprechau) for the patch and to
   [@atillamas](https://github.com/atillamas) for the original PR and the
   discussion.

 * [PR #453](https://github.com/fabiolb/fabio/pull/453): Handle proxy chains of any length

   Fabio will now validate that all elements of the `X-Forwarded-For` header
   are allowed by the given ACL of the route. See discussion in
   [PR #449](https://github.com/fabiolb/fabio/pull/449) for details.

   Thanks to [@leprechau](https://github.com/leprechau) for the patch and to
   [@atillamas](https://github.com/atillamas) for the original PR and the
   discussion.

 * [Issue #452](https://github.com/fabiolb/fabio/issues/452): Add improved glob matcher

   Fabio now uses the `github.com/gobaws/glob` package for glob matching which
   allows more complex patterns.

   Thanks to [@sharbov](https://github.com/sharbov) for the patch.

#### Features

 * None

### [v1.5.8](https://github.com/fabiolb/fabio/releases/tag/v1.5.8) - 18 Feb 2018

#### Breaking Changes

 * None

#### Bug Fixes

 * Fix windows build.

   fabio 1.5.7 broke the Windows build but this wasn't detected since the new
   build process did not build the Windows binaries. This has been fixed.

 * [Issue #438](https://github.com/fabiolb/fabio/pull/438): Do not add separator to `noroute.html` page

   fabio 1.5.7 added support for multiple routing tables in Consul and added a
   comment which described the origin to the output. The same comment was added
   to the `noroute.html` page since the same code is used to fetch it. This
   returned an invalid HTML page which has been fixed.

#### Improvements

 * [PR #423](https://github.com/fabiolb/fabio/pull/423): TCP+SNI support arbitrary large Client Hello

   With this patch fabio correctly parses `ClientHello` messages on TLS
   connections up to their maximum size.

   Thanks to [@DanSipola](https://github.com/DanSipola) for the patch.

#### Features

 * [PR #426](https://github.com/fabiolb/fabio/pull/426): Add option to allow Fabio to register frontend services in Consul on behalf of user services

   With this patch fabio can register itself multiple times under different
   names in Consul. By adding the `register=name` option to a route fabio will
   register itself under that name as well.

   Thanks to [@rileyje](https://github.com/rileyje) for the patch.

 * [PR #442](https://github.com/fabiolb/fabio/pull/442): Add basic ip centric access control on routes

   With this patch fabio adds an `allow` and `deny` option to the routes which
   allows for basic ip white and black listing of IPv4 and IPv6 addresses. See
   http://fabiolb.net/feature/access-control/ for more details.

   Thanks to [@leprechau](https://github.com/leprechau) for the patch and
   [@microadam](https://github.com/microadam) for the testing.

### [v1.5.7](https://github.com/fabiolb/fabio/releases/tag/v1.5.7) - 6 Feb 2018

#### Breaking Changes

 * None

#### Bug Fixes

 * [Issue #434](https://github.com/fabiolb/fabio/issues/434): VaultPKI tests fail with go1.10rc1

   All unit tests pass now on go1.10rc1.

#### Improvements

 * [Issue #369](https://github.com/fabiolb/fabio/issues/369): Warn if fabio is run as root

   fabio 1.5.7 emits a recurring warning when run as root. This can be disabled when using
   the new `-insecure` flag which also provides a link to alternatives.

 * [Issue #433](https://github.com/fabiolb/fabio/issues/433): `proxy.noroutestatus` must be three digit code

   go1.10 will enforce that HTTP status codes must be three digit values `[100,1000)` and
   and otherwise the handler will panic. This change enforces that the `proxy.noroutestatus`
   has a valid status code value.

#### Features

 * [Issue #396](https://github.com/fabiolb/fabio/issues/396): treat `registry.consul.kvpath` as prefix

   This patch allows fabio to have multiple manual routing tables stored in consul, e.g.
   under `fabio/config/foo` and `fabio/config/bar`. The routing table fragments are
   concatenated in lexicographical order of the keys and the log output contains comments
   to indicate to which key the segment belongs.

 * [PR #425](https://github.com/fabiolb/fabio/pull/425): Add support for HSTS headers

   fabio has now support for adding HSTS headers to the response.

   Thanks to [@leprechau](https://github.com/leprechau) for the patch.

### [v1.5.6](https://github.com/fabiolb/fabio/releases/tag/v1.5.6) - 5 Jan 2018

#### Breaking Changes

 * None

#### Improvements

 * [Issue #216](https://github.com/fabiolb/fabio/issues/216)/[Issue #383](https://github.com/fabiolb/fabio/issues/383)/[PR #414](https://github.com/fabiolb/fabio/pull/414): Do not require globally unique service IDs

	Since version 1.0 fabio required all service ids in Consul to be globally
	unique although service ids only have to be unique per Consul agent. This patch fixes this.

	Thanks to [@dropje86](https://github.com/dropje86) and [@alvaroaleman](https://github.com/alvaroaleman) for the patch!

 * [Issue #408](https://github.com/fabiolb/fabio/issues/408): Log Consul state changes as DEBUG

   `Health changed to xxx` and similar log messages will be logged as `DEBUG`.

 * [PR #415](https://github.com/fabiolb/fabio/pull/415): Honor the `-version` flag

   `fabio -version` does now what you would expect it to do.

### [v1.5.5](https://github.com/fabiolb/fabio/releases/tag/v1.5.5) - 21 Dec 2017

#### Breaking Changes

 * None

#### Features

 * [PR #398](https://github.com/fabiolb/fabio/pull/398): Add custom no route HTML page

   This patch adds support for a custom HTML template stored in Consul or on the file system which will be returned when
   there is no route.

   Thanks to [@tino](https://github.com/tino) for the patch!

### [v1.5.4](https://github.com/fabiolb/fabio/releases/tag/v1.5.4) - 10 Dec 2017

#### Breaking Changes

 * None

#### Features

 * [Issue #87](https://github.com/fabiolb/fabio/issues/87)/[PR #395](https://github.com/fabiolb/fabio/pull/395): Add redirect support

   This patch adds support to redirect a request for a matching route to
   another URL. If the `redirect=<code>` option is set on a route fabio will
   send a redirect response to the dst address with the given code.

   The syntax for the `urlprefix-` tag is slightly different since the
   destination address is usually generated from the service registration
   stored in Consul.

   The `$path` pseudo-variable can be used to include the original request URI
   in the destination target.

   Thanks to [@ctlajoie](https://github.com/ctlajoie) for providing this patch!

```
# redirect /foo to https://www.foo.com/
route add svc /foo https://www.foo.com/ opts "redirect=301"

# redirect /foo to https://www.foo.com/
urlprefix-/foo redirect=301,https://www.foo.com/

# redirect /foo to https://www.foo.com/foo
urlprefix-/foo redirect=301,https://www.foo.com$path
```

#### Bug Fixes

 * [Issue #385](https://github.com/fabiolb/fabio/issues/385): opts with host= with multiple routes does not work as expected

   When multiple routes for the same path had different `host` options then only the one set on the
   first route worked. This has been fixed so that the `Host` header is now set according to the
   selected target.

 * [Issue #389](https://github.com/fabiolb/fabio/issues/389): match exact host before glob matches

   When there is an exact match and a glob match for a hostname then the exact match
   is preferred.

#### Improvements

 * [PR #380](https://github.com/fabiolb/fabio/pull/380): Set X-Forwared-Host header if not present

   Fabio now sets the `X-Forwarded-Host` header if it isn't present.

 * [Issue #400](https://github.com/fabiolb/fabio/issues/400): Do not exit on SIGHUP

   Fabio will now ignore the `SIGHUP` signal. Additionally, the caught signal is logged with the action (exit or ignore).

### [v1.5.3](https://github.com/fabiolb/fabio/releases/tag/v1.5.3) - 3 Nov 2017

#### Breaking Changes

 * None

#### Features

 * [PR #315](https://github.com/fabiolb/fabio/pull/315)/[Issue #135](https://github.com/fabiolb/fabio/issues/135): Vault PKI cert source

   This adds support for using [Vault](https://vaultproject.io/) as a PKI cert source.

   Thanks to [@pschultz](https://github.com/pschultz) for providing this patch!

#### Bug Fixes

 * [Issue #306](https://github.com/fabiolb/fabio/issues/306): Add metrics for TCP and TCP+SNI proxy

   fabio now reports metrics for TCP and TCP+SNI connections.

 * [Issue #330](https://github.com/fabiolb/fabio/issues/330): Strip option has no effect on websockets

   The `strip=/prefix` option now works correctly on web sockets

 * [Issue #350](https://github.com/fabiolb/fabio/issues/350): statsd - unable to parse line - gf metric

   fabio now correctly reports mean values for timers as gauge values to statsd.

#### Improvements

 * [Issue #320](https://github.com/fabiolb/fabio/issues/320): FATAL error when metrics cannot be delivered

   fabio adds a `metrics.timeout` and a `metrics.retry` config parameter to control when the
   the metrics backend should become available and changes the default behavior to retry for
   some time before giving up.

 * [PR #366](https://github.com/fabiolb/fabio/pull/366): add leveled logging

   Add a `-log-level` parameter which allows to control the log level.

 * [Issue #367](https://github.com/fabiolb/fabio/issues/367): nodes and services in maintenance can cause excessive logging

   Notifications about nodes and services in maintenance mode are now logged as DEBUG and therefore
   filtered out by default.

 * [Issue #375](https://github.com/fabiolb/fabio/issues/375): `host` option allows to set `Host` header

   The `host` option now allows to set the `Host` header to the provided value in addition to the special `dst` value.

### [v1.5.2](https://github.com/fabiolb/fabio/releases/tag/v1.5.2) - 24 Jul 2017

#### Breaking Changes

 * None

#### Bug Fixes

 * [Issue #305](https://github.com/fabiolb/fabio/issues/305): 1.5.0 config compatibility problem

   In fabio 1.5.0 the key/value parsing was refactored and that introduced a bug where a second `=`
   failed to parse correctly and prevented fabio from starting.

#### Improvements

 * [PR #321](https://github.com/fabiolb/fabio/pull/321): Cleanup TCP proxy connections

   This patch updates the internal connection map when a connection is closed.

   Thanks to [@crypto89](https://github.com/crypto89) for this patch.

### [v1.5.1](https://github.com/fabiolb/fabio/releases/tag/v1.5.1) - 6 Jul 2017

#### Improvements

 * Added Code of Conduct

 * Add support for `detail` format for `log.routes.format`

   The `detail` format prints the routing table with more detail than the other formats
   and it isn't intended to be machine readable.

```
./fabio -log.routes.format detail
2017/06/19 11:51:14 [INFO] Updated config to
+-- host=:3306
|   +-- path=
|       |-- addr=127.0.0.1:5001 weight 0.20 slots 2000/10000
|       +-- addr=127.0.0.1:5000 weight 0.80 slots 8000/10000
+-- host=:3307
    +-- path=
        +-- addr=127.0.0.1:5002 weight 1.00 slots 1/1
```

 * [Issue #42](https://github.com/fabiolb/fabio/issues/42): Add support for 'weight=f' option in urlprefix tag

   This allows to specify a manual weight on the `urlprefix-` tag. This can be used to
   manually distribute the load between multiple TCP endpoints or to have an active/standby
   setup by setting `weight=1` on the active and `weight=0` on the standby server.

 * [Issue #274](https://github.com/fabiolb/fabio/issues/274)/[PR #314](https://github.com/fabiolb/fabio/pull/314): Avoid premature Vault token renewal

   Non-renewable tokens are no longer renewed. In addition, the token TTL is honored for token that can
   be renewed.

   Thanks to [@pschultz](https://github.com/pschultz) for this patch.

 * [PR #313](https://github.com/fabiolb/fabio/pull/313): Tests work now with Vault 0.7.x

   Thanks to [@pschultz](https://github.com/pschultz) for this patch.

### [v1.5.0](https://github.com/fabiolb/fabio/releases/tag/v1.5.0) - 7 Jun 2017

#### Breaking Changes

 * Support for the deprecated `proxy.addr` format `:port;certfile;keyfile;cafile` has been dropped.
   Please use instead `proxy.addr` in combination with a
   [certificate store](https://github.com/fabiolb/fabio/wiki/Features#certificate-stores).

#### Bug Fixes

#### Improvements

 * Upgrade to go1.8.3

 * [Issue #133](https://github.com/fabiolb/fabio/issues/133): websockets failing with 500 on rancher

   Rancher is a Java application which uses `java.net.URL` to compose
   the original request URL from the `X-Forwarded-Proto` and other
   headers. The `java.net.URL` class does not support the `ws` or `wss`
   protocol without a matching `java.net.URLStreamHandler`
   implementation. Java code should use the `java.net.URI` class for
   these types of URLs instead. However, the `X-Forwarded-Proto` header
   isn't specified as the `Forwarded` header ([RFC
   7239](https://tools.ietf.org/html/rfc7239#section-5.4)) and the
   common usage is to only use either `http` or `https` for websocket
   connections. In order not to break existing applications fabio now
   sets the `X-Forwarded-Proto` header to `http` for `ws` and to `https`
   for `wss` connections.

 * [PR #292](https://github.com/fabiolb/fabio/pull/292): Add unique request id

   fabio can now add a unique request id in form of a UUIDv4 to each request as a header.
   The name of the header is configurable and the value of the header can be logged
   to the access log.

   Thanks to [@bkmit](https://github.com/bkmit) for this patch.

 * [Issue #249](https://github.com/fabiolb/fabio/issues/249): Make TLS version and cipher suites configurable

   fabio now allows to configure the TLS parameters for the handshake as part of the
   `proxy.addr` configuration. See `fabio.properties` for details.

 * [Issue #280](https://github.com/fabiolb/fabio/issues/280): Add protocol data to `Forwarded` header

   fabio adds `httpproto`, `tlsver` and `tlsciphers` to the `Forwarded` header.

 * [Issue #290](https://github.com/fabiolb/fabio/issues/290): Add profiling support

   fabio now supports optional memory, cpu, mutex and block (contention) profiling.
   Profiling is enabled through the `profile.mode` flag which determines the mode.
   The `profile.path` flag determines the output path.

 * [Issue #294](https://github.com/fabiolb/fabio/issues/294): Use upstream host name for request

    Add support for a `host=dst` option on the route to trigger fabio to
	use the target hostname for the outgoing request instead of the
	host name provided by the original request.

 * [Issue #296](https://github.com/fabiolb/fabio/issues/296): Sync X-Forwarded-Proto and Forwarded header when possible

   The X-Forwarded-Proto header and the proto value of the Forwarded
   header can get out of sync when an upstream load balancer sets the
   one but not the other header. Fabio would then not touch the existing
   header and derive the value for the unset header based on the
   connection.

   This patch changes this behavior so that the value for the missing
   header is derived from the other one. When both headers are set they are
   both left untouched since it cannot be decided which one is the source
   of truth.

 * [Issue #300](https://github.com/fabiolb/fabio/issues/300): Support Gzip encoding for websockets

   Setting the `Accept-Encoding` header to `gzip` and enabling gzip compression
   triggered a bug in fabio which prevented the use of gzip compression on
   websocket connections.

 * [Issue #302](https://github.com/fabiolb/fabio/issues/302): Add support for read-only UI

   The `ui.access` parameter can be used to configure the ui endpoint to
   be in either read-write or read-only mode.

 * [Issue #304](https://github.com/fabiolb/fabio/issues/304): Add support for X-Forwarded-Prefix header

   The `X-Forwarded-Prefix` header is added when the `strip=/foo` option
   is used on a route and contains the path that was stripped (e.g.
   `/foo`).

### [v1.4.4](https://github.com/fabiolb/fabio/releases/tag/v1.4.4) - 8 May 2017

#### Bug Fixes

 * [Issue #271](https://github.com/fabiolb/fabio/issues/271): Support websocket for HTTPS upstream

   This patch fixes that websocket connections are not forwarded to an HTTPS upstream server.

 * [Issue #279](https://github.com/fabiolb/fabio/issues/279): fabio does not start with multiple listeners

   Commit [5a23cb1](https://github.com/fabiolb/fabio/commit/5a23cb19dc64a30ee40c42bd3ec1dde289a91033)
   found in [#265](https://github.com/fabiolb/fabio/issues/265) added code for
   not swallowing the errors but did not capture the loop variable for the go
   routines when starting listeners. This prevented fabio from starting up
   properly when being configured with more than one listener.

 * [Issue #289](https://github.com/fabiolb/fabio/issues/289): Fabio does not advertise http/1.1 on TLS connections

   This patch makes fabio announce both `h2` and `http/1.1` as application level protocols
   on TLS connections.

#### Improvements

 * The listener code no longer swallows the errors and exits if it cannot create
   a listening socket.

 * [Issue #278](https://github.com/fabiolb/fabio/issues/278): Add service name to access log fields

   Add `$upstream_service` which contains the service name of the selected target
   to the available access log fields.

### [v1.4.3](https://github.com/fabiolb/fabio/releases/tag/v1.4.3) - 24 Apr 2017

#### Bug Fixes

 * [Issue #269](https://github.com/fabiolb/fabio/issues/269): Access log cannot be disabled

   The access logging feature that was added in v1.4.1 did not allow to disable the access logging
   output and all fabio instances were writing an access log by default. Also, the logging setup
   code would leave fabio registered in consul in case of a failure.

#### Improvements

 * [PR #268](https://github.com/fabiolb/fabio/pull/268): Add support for TLSSkipVerify for https consul fabio check

   When the fabio admin port is configured to use HTTPS then the consul health check has
   to use HTTPS as well. The new `registry.consul.register.checkTLSSkipVerify` option allows
   to disable TLS certificate validation for this check. This requires consul 0.7.2 or higher.

   Thanks to [@Ginja](https://github.com/Ginja) for providing this patch.

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

 * [PR #257](https://github.com/fabiolb/fabio/pull/257), [Issue #181](https://github.com/fabiolb/fabio/issues/181): Add HTTPS Upstream Support

   Upstream servers can now be served via HTTPS. To enable this for a route add the `proto=https` option
   to the `urlprefix-` tag. The upstream certificate needs to be in the system certificate chain for the
   certificate validation to succeed. To disable certificate validation for upstream requests add the
   `tlsskipverify=true` option. Support for certificate stores for upstream servers may come at a later
   point.

   Thanks to [@shadowfax-chc](https://github.com/shadowfax) for providing this patch.

   See: https://github.com/fabiolb/fabio/wiki/Features#https-upstream-support

 * [PR #258](https://github.com/fabiolb/fabio/pull/258): Allow UI/API to be served over HTTPS

   The UI/API endpoint can now be served via HTTPS. To enable this configure the `ui.addr` property
   with a `cs=<cert store>` option like the `proxy.addr` listeners.

   Thanks to [@shadowfax-chc](https://github.com/shadowfax) for providing this patch.

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

   Thanks to [@beyondblog](https://github.com/beyondblog) for providing the initial patch.

   See: https://github.com/fabiolb/fabio/wiki/Features#access-logging

### [v1.4](https://github.com/fabiolb/fabio/releases/tag/v1.4) - 25 Mar 2017

#### Features

 * [Issue #1](https://github.com/fabiolb/fabio/issues/1), [Issue #179](https://github.com/fabiolb/fabio/issues/179): Add generic TCP Proxy support

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

 * [Issue #44](https://github.com/fabiolb/fabio/issues/44), [Issue #124](https://github.com/fabiolb/fabio/issues/124), [Issue #164](https://github.com/fabiolb/fabio/issues/164): Support path stripping
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
 * [Issue #151](https://github.com/fabiolb/fabio/issues/151)/[PR #150](https://github.com/fabiolb/fabio/pull/150): Add support for Circonus metrics

#### Improvements

 * [Issue #125](https://github.com/fabiolb/fabio/issues/125): Extended metrics
 * [Issue #134](https://github.com/fabiolb/fabio/issues/134): Vault token should not require 'root' or 'sudo' privileges
 * [PR #154](https://github.com/fabiolb/fabio/pull/154): Make route metric names configurable

### [v1.2.1](https://github.com/fabiolb/fabio/releases/tag/v1.2.1) - 25 Aug 2016

#### Features

 * [Issue #73](https://github.com/fabiolb/fabio/pull/73)/[PR #139](https://github.com/fabiolb/fabio/pull/139): Add statsd support
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

 * [PR #5](https://github.com/fabiolb/fabio/pull/5): Fix typo

### [v1.0.2](https://github.com/fabiolb/fabio/releases/tag/v1.0.2) - 23 Oct 2015

#### Improvements

 * [PR #3](https://github.com/fabiolb/fabio/pull/3): Honor consul.url and consul.addr from config file ([@jeinwag](https://github.com/jeinwag))

### [v1.0.1](https://github.com/fabiolb/fabio/releases/tag/v1.0.1) - 21 Oct 2015

#### Improvements

 * Honor maintenance mode for both services and nodes

### [v1.0.0](https://github.com/fabiolb/fabio/releases/tag/v1.0.0) - 16 Oct 2015

 * Initial open-source release
