# fabio

fabio is a fast, modern, zero-conf load balancing HTTP router for deploying
microservices. Services provide one or more host/path prefixes they serve and
fabio updates the routing table every time a service becomes (un-)available
without restart.

fabio was developed at the [eBay Classifieds Group](http://www.ebayclassifiedsgroup.com)
in Amsterdam and is currently used to route traffic for
[marktplaats.nl](http://www.makrtplaats.nl) and [kijiji.it](http://www.kijiji.it).
Marktplaats is running all of its traffic through fabio which is
several thousand requests per second distributed over several fabio
instances.

## Features

* Single binary in Go. No external dependencies.
* Zero-conf
* Hot-reloading of routing table through backend watchers
* Round robin and random distribution
* [Traffic Shaping](#Traffic Shaping) (send 5% of traffic to new instances)
* Graphite metrics
* Request tracing
* WebUI
* Fast

fabio listens on a single HTTP port for incoming requests and routes
them to the registered services.

## Installation

To install fabio run (you need Go 1.4 or higher)

    go get github.com/eBay/fabio

To start fabio run

    ./fabio

which will run it with the default configuration which is described
in `fabio.properties`. To run it with a config file run it
with

    ./fabio -cfg cfgfile

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
check the time fabio spends in the GC with `GODEBUG=gotrace=1`.

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
route add service host/path targetURL [weight <weight>] [tags "tag1,tag2,..."]
	- Add a new route for host/path to targetURL

route del service
	- Remove all routes for service

route del service host/path
	- Remove all routes for host/path for this service only

route del service host/path targetURL
	- Remove only this route

route weight service host/path weight n tags "tag1,tag2"
  - Route n% of traffic to services matching service, host/path and tags
    n is a float > 0 describing a percentage, e.g. 0.5 == 50%
    n <= 0: means no fixed weighting. Traffic is evenly distributed
    n > 0: route will receive n% of traffic. If sum(n) > 1 then n is normalized.
    sum(n) >= 1: only matching services will receive traffic

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

## Example

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

fabio allows to control the amount of traffic a set of service
instances will receive. You can use this feature to direct a fixed percentage
of traffic to a newer version of an existing service for testing ("Canary
testing").

The following command will allocate 5% of traffic to `www.kjca.dev/auth/` to
all instances of `service-b` which match tags `version-15` and `dc-fra`. This
is independent of the number of actual instances running. The remaining 95%
of the traffic will be distributed evenly across the remaining instances
publishing the same prefix.

```
route weight service-b www.kjca.dev/auth/ weight 0.05 tags "version-15,dc-fra"
```

### Traffic shaping with multiple active fabio instances

The percentage calculation is currently local to the fabio instance.
That means that each fabio will send N percent of traffic to a
service for which traffic shaping is enabled. Therefore, if you want to
send 10% of traffic to a service and have two fabio instances
running you need to set the percentage to 5%.

This will change in a later version when fabio registers itself in
consul and can adapt the percentages automatically depending on the number
of active fabio instances.

## Debugging

To send a request from the command line via the fabio using `curl`
you should send it as follows:

```
curl -v -H 'Host: foo.com' 'http://localhost:9999/path'
```

The `-x` or `--proxy` options will most likely not work as you expect as they
send the full URL instead of just the request URI which usually does not match
any route but the default one - if configured.

### Tracing a request

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

fabio contains a (very) simple web ui to examine the routing
table. By default it is accessible on `http://localhost:9998/`

## Roadmap

The following features are planned to be added next.

* HTTP/2 support
* Correct traffic shaping with multiple fabio instances

## License

MIT licensed

