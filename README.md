# ![./fabio](https://github.com/eBay/fabio/blob/master/fabio.png)

##### Current stable version: 1.1

[![Build Status](https://travis-ci.org/eBay/fabio.svg?branch=master)](https://travis-ci.org/eBay/fabio) [![License MIT](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/eBay/fabio/master/LICENSE)

fabio is a fast, modern, zero-conf load balancing HTTP(S) router
for deploying microservices managed by consul.

## Why build another HTTP router?

Both hardware and software routers like Citrix Netscaler, F5 Big IP, haproxy,
varnish, nginx or apache require some form of configuration - the routing
table - to route incoming traffic to services which can handle them. The
routing table has to be kept in sync with the actual deployed set of services
and instances during each deployment and each outage on every environment.
This makes the routing table a crucial part of the configuration without the
application cannot function.

Managing the routing table can be automated via API calls or tools like
[consul-template](https://github.com/hashicorp/consul-template) but that also
requires configuration and/or tools. In the case of consul-template the config
file template itself has to be kept in sync with the actual setup of the
application. Finally, updating the routing table without loss of existing
connections can be [challenging](http://engineeringblog.yelp.com/2015/04/true-
zero-downtime-haproxy-reloads.html).

fabio solves this problem by making the services themselves responsible for
updating the routing table. Services already know which routes they serve since
they have handlers who can handle requests for them. Once services push the
routes they handle into the service registry (in this case consul) fabio can
build the routing table and can re-configure itself on every change
automatically without restart and without the loss of existing connections.

The motivation is also outlined in the presentation I've given at the dotGo EU
pre-party in Paris  on 9 Nov 2015. You can watch it
[here](https://www.youtube.com/watch?v=82UAB3qEe54).

fabio was developed at the [eBay Classifieds Group](http://www.ebayclassifiedsgroup.com)
in Amsterdam and routes in total roughly 15.000 req/sec for the following sites without
any measurable latency impact.

* [marktplaats.nl](http://www.marktplaats.nl)
* [admarkt.marktplaats.nl](http://admarkt.marktplaats.nl)
* [topannoncer.dbabusiness.dk](http://topannoncer.dbabusiness.dk)
* [cas.kijiji.ca](http://cas.kijiji.ca)
* [www.kijiji.ij](http://www.kijiji.it)

(drop me a note if you want to have your site listed here)

## Features

* Single binary in Go. No external dependencies.
* Zero-conf
* Hot-reloading of routing table through backend watchers
* Round robin and random distribution
* [Traffic Shaping](#traffic-shaping) (send 5% of traffic to new instances)
* Graphite metrics
* Request tracing
* WebUI
* Fast
* v1.0.4: SSL client certificate authentication support (see `proxy.addr` in [fabio.properties](https://raw.githubusercontent.com/eBay/fabio/master/fabio.properties))
* v1.0.5: `X-Forwarded-For` and `Forwarded` header support
* v1.0.5: Websocket support (experimental)
* v1.0.6: Raw websocket support as default
* v1.0.6: Experimental HTTP API
* v1.0.6: Improved UI
* v1.0.6: fabio registers itself in consul
* v1.0.8: Support consul ACL token
* v1.0.9: Make read and write timeout configurable

## Documentation

* [Quickstart](#quickstart)
* [Installation](#installation)
* [Contribution](#contribute-to-fabio)
* [Configuration](#configuration)
* [Deployment](#deployment)
* [Performance](#performance)
* [Service configuration](#service-configuration)
* [Manual overrides](#manual-overrides)
* [Routing](#routing)
* [Traffic shaping](#traffic-shaping)
* [Websockets](#websockets)
* [Debugging](#debugging)
* [Request tracing](#request-tracing)
* [Web UI](#web-ui)
* [Changelog](https://github.com/eBay/fabio/blob/master/CHANGELOG.md)
* [License](#license)

## Quickstart

This is how you use fabio in your setup:

1. Register your service in consul
2. Register a **health check** in consul as described [here](https://consul.io/docs/agent/checks.html).
   Make sure the health check is **passing** since fabio will only watch services
   which have a passing health check.
3. Register one `urlprefix-` tag per `host/path` prefix it serves,
   e.g. `urlprefix-/css`, `urlprefix-/static`, `urlprefix-mysite.com/`
4. Start fabio without a config file (assuming a consul agent on `localhost:8500`)
   Watch the log output how fabio picks up the route to your service.
   Try starting/stopping your service to see how the routing table changes instantly.
5. Send all your HTTP traffic to fabio on port `9999`
6. Done

To start a sample server to test the routing run the `demo/server` like this:

    ./server -addr 127.0.0.1:5000 -name svc-a -prefix /foo

and access the server direct and via fabio

    curl 127.0.0.1:5000/foo   # direct
    curl 127.0.0.1:9999/foo   # via fabio

If you want fabio to handle SSL as well set the `proxy.addr` along with the
public/private key files in
[fabio.properties](https://github.com/eBay/fabio/blob/master/fabio.properties)
and run `fabio -cfg fabio.properties`. You might also want to set the
`proxy.header.clientip`, `proxy.header.tls` and `proxy.header.tls.value`
options.

Check the [Debugging](#debugging) section to see how to test fabio with `curl`.

See fabio in action

[![fabio demo](http://i.imgur.com/aivFAKl.png)](https://www.youtube.com/watch?v=gvxxu0PLevs"fabio demo - Click to Watch!")

The `fabio-example` project is now in the `demo/server` directory.

## Installation

To install fabio you need Go 1.5.3 or higher. Run

    GO15VENDOREXPERIMENT=1 go get github.com/eBay/fabio

To start fabio run

    ./fabio

which will run it with the default configuration which is described
in `fabio.properties`. To run it with a config file run it
with

    ./fabio -cfg fabio.properties

or use the official Docker image and mount your own config file to `/etc/fabio/fabio.properties`

    docker run -d -p 9999:9999 -p 9998:9998 -v $PWD/fabio/fabio.properties:/etc/fabio/fabio.properties magiconair/fabio

If you want to run the Docker image with one or more SSL certificates then
you can store your configuration and certificates in `/etc/fabio` and mount
the entire directory, e.g.

    $ cat ~/fabio/fabio.properties
    proxy.addr=:443;/etc/fabio/ssl/mycert.pem;/etc/fabio/ssl/mykey.pem

    docker run -d -p 443:443 -p 9998:9998 -v $PWD/fabio:/etc/fabio magiconair/fabio

The official Docker image contains the root CA certificates from a recent and updated
Ubuntu 12.04.5 LTS installation.

## Contribute to fabio

Contributions to fabio of any kind are welcome including documentation, examples,
feature requests, bug reports, discussions, helping with issues, etc.

If you have a question on how or what to contribute just open an issue and
indicate that it is a question.

### Code Contribution Guideline

Your contribution is welcome. To make merging code as seamless as possible
we ask for the following:

* For small changes and bug fixes go ahead, fork the project, make your changes
  and send a pull request.
* Larger changes should start with a proposal in an issue. This should ensure
  that the requested change is in line with the project and similar work is not
  already underway.
* Only add libraries if they provide significant value. Consider copying the code
  (attribution) or writing it yourself.
* Manage dependencies in the vendor path via `govendor` as separate commits per library.
  Please make sure your commit message has the following format:

```
Vendoring in version <git hash> of <pkgname>
```

Once you are ready to send in a pull request, be sure to:

* Sign the [CLA](https://cla-assistant.io/eBay/fabio)
* Provide test cases for the critical code which test correctness. If your code
  is in a performance critical path make sure you have performed some real world
  measurements to ensure that performance is not degregated.
* `go fmt` and `make test` your code
* Squash your change into a single commit with the exception of additional libraries.
* Write a good commit message.

## Configuration

fabio is configured to listen on port 9999 for HTTP traffic and uses
consul on `localhost:8500` as the default registry backend. To configure
additional listeners, different backends, enable metrics reporting or
change other configuration parameters please check the well documented
[fabio.properties](https://raw.githubusercontent.com/eBay/fabio/master/fabio.properties)
file. Each property value can also be configured via a corresponding
environment variable which has the dots replaced with underscores.

Example:

```
# fabio.properties
metrics.target = stdout

# correspondig env var
metrics_target=stdout ./fabio
```

## Deployment

The main use-case for fabio is to distribute incoming HTTP(S) requests
from the internet to frontend (FE) services which can handle these requests.
In this scenario the FE services then use the service discovery feature in
consul to find backend (BE) services they need in order to fulfil the
request.

That means that fabio is currently not used as an FE-BE or BE-BE router to
route traffic among the services themselves since the service discovery of
consul already solves that problem. Having said that, there is nothing that
inherently prevents fabio from being used that way. It just means that we
are not doing it.

### Direct

In the following setup fabio is configured to listen on the public ip(s)
where it can optionally terminate SSL traffic for one or more domains - one ip per domain.

```
                                           +--> service-a
                                           |
internet -- HTTP/HTTPS --> fabio -- HTTP --+--> service-b
                                           |
                                           +--> service-c
```

To scale fabio you can deploy it together with the frontend services which provides
high-availability and distributes the network bandwidth.

```
           +- HTTP/HTTPS -> fabio -+- HTTP -> service-a (host-a)
           |                       |
internet --+- HTTP/HTTPS -> fabio -+- HTTP -> service-b (host-b)
           |                       |
           +- HTTP/HTTPS -> fabio -+- HTTP -> service-c (host-c)
```

### Behind an existing LB/Gateway

In the following setup fabio is configured receive all incoming traffic
from an existing gateway which also terminates SSL for one or more domains.
fabio supports SSL Client Certificate Authentication to support the
[Amazon API Gateway](https://aws.amazon.com/api-gateway/)

```
                                                          +--> service-a
                                                          |
internet -- HTTP/HTTPS --> LB -- HTTP --> fabio -- HTTP --+--> service-b
                                                          |
                                                          +--> service-c
```

Again, to scale fabio you can deploy it together with the frontend services
which provides high-availability and distributes the network bandwidth.

```
                               +- HTTP -> fabio -+-> service-a (host-a)
                               |                 |
internet -- HTTP/HTTPS --> LB -+- HTTP -> fabio -+-> service-b (host-b)
                               |                 |
                               +- HTTP -> fabio -+-> service-c (host-c)
```


## Performance

fabio has been tested to deliver up to 15.000 req/sec on a single 16
core host with moderate memory requirements (~ 60 MB).

To achieve the performance fabio sets the following defaults which
can be overwritten with the environment variables:

* `GOMAXPROCS` is set to `runtime.NumCPU()` since this is not the
  default for Go 1.4 and before
* `GOGC=800` is set to reduce the pressure on the garbage collector

When fabio is compiled with Go 1.5 and run with default settings it can be up
to 40% slower  than the same version compiled with Go 1.4. The `GOGC=100`
default puts more pressure on the Go 1.5 GC which makes the fabio spend 10% of
the time in the GC. With `GOGC=800` this drops back to 1-2%. Higher values
don't provide higher gains.

As usual, don't rely on these numbers and perform your own benchmarks. You can
check the time fabio spends in the GC with `GODEBUG=gctrace=1`.

## Service configuration

Each service can register one or more URL prefixes for which it serves
traffic. A URL prefix is a `host/path` combination without a scheme since SSL
has already been terminated and all traffic is expected to be HTTP. To
register a URL prefix add a tag `urlprefix-host/path` to the service
definition.

By default, traffic is distributed evenly across all service instances which
register a URL prefix but you can set the amount of traffic a set of instances
will receive ("Canary testing"). See [Traffic Shaping](#Traffic Shaping)
below.

A background process watches for service definition and health status changes
in consul. When a change is detected a new routing table is constructed using
the commands described in [Config Commands](#Config Commands).

## Manual overrides

Since an automatically generated routing table can only be changed with a
service deployment additional routing commands can be stored manually in the
consul KV store which get appended to the automatically generated routing
table. This allows fine-tuning and fixing of problems without a deployment.

The [Traffic Shaping](#Traffic Shaping) commands are also stored in the KV
store.

## Routing Table Configuration

The routing table is configured with the following commands:

```
route add <svc> <src> <dst> weight <w> tags "<t1>,<t2>,..."
  - Add route for service svc from src to dst and assign weight and tags

route add <svc> <src> <dst> weight <w>
  - Add route for service svc from src to dst and assign weight

route add <svc> <src> <dst> tags "<t1>,<t2>,..."
  - Add route for service svc from src to dst and assign tags

route add <svc> <src> <dst>
  - Add route for service svc from src to dst

route del <svc> <src> <dst>
  - Remove route matching svc, src and dst

route del <svc> <src>
  - Remove all routes of services matching svc and src

route del <svc>
  - Remove all routes of service matching svc

route weight <svc> <src> weight <w> tags "<t1>,<t2>,..."
  - Route w% of traffic to all services matching svc, src and tags

route weight <src> weight <w> tags "<t1>,<t2>,..."
  - Route w% of traffic to all services matching src and tags

route weight <svc> <src> weight <w>
  - Route w% of traffic to all services matching svc and src

route weight service host/path weight w tags "tag1,tag2"
  - Route w% of traffic to all services matching service, host/path and tags

    w is a float > 0 describing a percentage, e.g. 0.5 == 50%
    w <= 0: means no fixed weighting. Traffic is evenly distributed
    w > 0: route will receive n% of traffic. If sum(w) > 1 then w is normalized.
    sum(w) >= 1: only matching services will receive traffic

    Note that the total sum of traffic sent to all matching routes is w%.

```

The order of commands matters but routes are always ordered from most to least
specific by prefix length.

## Routing

The routing table contains first all routes with a host sorted by prefix
length in descending order and then all routes without a host again sorted by
prefix length in descending order.

For each incoming request the routing table is searched top to bottom for a
matching route. A route matches if either `host/path` or - if there was no
match - just `/path` matches.

The matching route determines the target URL depending on the configured
strategy. `rnd` and `rr` are available with `rnd` being the default.

### Example

The auto-generated routing table is

```
route add service-a www.mp.dev/accounts/ http://host-a:11050/ tags "a,b"
route add service-a www.kjca.dev/accounts/ http://host-a:11050/ tags "a,b"
route add service-a www.dba.dev/accounts/ http://host-a:11050/ tags "a,b"
route add service-b www.mp.dev/auth/ http://host-b:11080/ tags "a,b"
route add service-b www.kjca.dev/auth/ http://host-b:11080/ tags "a,b"
route add service-b www.dba.dev/auth/ http://host-b:11080/ tags "a,b"
```

The manual configuration under `/fabio/config` is

```
route del service-b www.dba.dev/auth/
route add service-c www.somedomain.com/ http://host-z:12345/
```

The complete routing table then is

```
route add service-a www.mp.dev/accounts/ http://host-a:11050/ tags "a,b"
route add service-a www.kjca.dev/accounts/ http://host-a:11050/ tags "a,b"
route add service-a www.dba.dev/accounts/ http://host-a:11050/ tags "a,b"
route add service-b www.mp.dev/auth/ http://host-b:11080/ tags "a,b"
route add service-b www.kjca.dev/auth/ http://host-b:11080/ tags "a,b"
route add service-c www.somedomain.com/ http://host-z:12345/ tags "a,b"
```

## Traffic Shaping

fabio allows to control the amount of traffic a set of service instances will
receive. You can use this feature to direct a fixed percentage of traffic to a
newer version of an existing service for testing ("Canary testing"). See
[Manual Overrides](#manual-overrides) for a complete description of the `route
weight` command.

The following command will allocate 5% of traffic to `www.kjca.dev/auth/` to
all instances of `service-b` which match tags `version-15` and `dc-fra`. This
is independent of the number of actual instances running. The remaining 95%
of the traffic will be distributed evenly across the remaining instances
publishing the same prefix.

```
route weight service-b www.kjca.dev/auth/ weight 0.05 tags "version-15,dc-fra"
```

## Websockets

Websocket support works but is considered experimental since I don't have an
in-house use case for it at the moment. I would like to hear from users whether it
works in their environments beyond my simple test case before I declare it stable.
It has been implemented with the websocket library from
[golang.org/x/net/websocket](http://golang.org/x/net/websocket)

You can test the websocket support with the `demo/wsclient` and `demo/server` which
implements a simple echo server.

    ./server -addr 127.0.0.1:5000 -name ws-a -prefix /echo -proto ws
    ./wsclient -url ws://127.0.0.1:9999/echo

You can also run multiple web socket servers on different ports but the same endpoint.

fabio detects on whether to forward the request as HTTP or WS based on the
value of the `Upgrade` header. If the value is `websocket` it will attempt a
websocket connection to the target. Otherwise, it will fall back to HTTP.

One limitation of the current implementation is that the accepted set of
protocols has to be symmetric across all services handling it. Only the
following combinations will work reliably:

    svc-a and svc-b register /foo and accept only HTTP traffic there
    svc-a and svc-b register /foo and accept only WS traffic there
    svc-a and svc-b register /foo and accept both HTTP and WS traffic there

The following setup (or variations thereof) will not work reliably:

    svc-a registers /foo and accept only WS traffic there
    svc-b registers /foo and accept only HTTP traffic there

This is not a limitation of the routing itself but because the current
configuration does not provide fabio with enough information to make the
routing decision since the services do not advertise the protocols they handle
on a given endpoint.

This does not look like a big restriction but is also not difficult to extend
in a later version assuming there are use cases which require this behavior.
For now the services have to be symmetric in the protocols they accept.

## Debugging

To send a request from the command line via the fabio using `curl`
you should send it as follows:

```
curl -v -H 'Host: foo.com' 'http://localhost:9999/path'
```

The `-x` or `--proxy` options will most likely not work as you expect as they
send the full URL instead of just the request URI which usually does not match
any route but the default one - if configured.

## Request tracing

To trace how a request is routed you can add a `Trace` header with an non-
empty value which is truncated at 16 characters to keep the log output short.

```
$ curl -v -H 'Trace: abc' -H 'Host: foo.com' 'http://localhost:9999/bar/baz'

2015/09/28 21:56:26 [TRACE] abc Tracing foo.com/bar/baz
2015/09/28 21:56:26 [TRACE] abc No match foo.com/bang
2015/09/28 21:56:26 [TRACE] abc Match foo.com/
2015/09/28 22:01:34 [TRACE] abc Routing to http://1.2.3.4:8080/
```

## Web UI

fabio contains a simple web ui to examine the routing table and manage the
manual overrides. By default it is accessible on `http://localhost:9998/`

## License

MIT licensed
