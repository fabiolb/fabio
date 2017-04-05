# ![./fabio](https://github.com/eBay/fabio/blob/master/fabio.png)

##### Current stable version: 1.4.1

[![Build Status](https://travis-ci.org/eBay/fabio.svg?branch=master)](https://travis-ci.org/eBay/fabio)
[![License MIT](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/eBay/fabio/master/LICENSE)
[![Downloads](https://img.shields.io/github/downloads/eBay/fabio/total.svg)](https://github.com/eBay/fabio/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/magiconair/fabio.svg)](https://hub.docker.com/r/magiconair/fabio/)

fabio is a fast, modern, zero-conf load balancing HTTP(S) and TCP router
for deploying applications managed by [consul](https://consul.io/).

Register your services in consul, provide a health check and fabio will start routing traffic to them. No configuration required. Deployment, upgrading and refactoring has never been easier.

fabio is developed and maintained by [Frank Schroeder](https://twitter.com/magiconair) at [eBay in Amsterdam](http://www.ebayclassifiedsgroup.com/).
It powers some of the largest websites in The Netherlands ([marktplaats.nl](http://www.marktplaats.nl/)), Australia ([gumtree.com.au](http://www.gumtree.com.au)) and Italy ([www.kijiji.it](http://www.kijiji.it/)).
It delivers 23.000 req/sec every day since Sep 2015 without problems.

It integrates with
[Consul](https://consul.io/),
[Vault](https://vaultproject.io/),
[Amazon ELB](https://aws.amazon.com/elasticloadbalancing),
[Amazon API Gateway](https://aws.amazon.com/api-gateway/)
and more.

It supports
SSL,
[TCP proxy](https://github.com/eBay/fabio/wiki/Features#tcp-proxy-support),
[TCP+SNI proxy (full end-to-end TLS)](https://github.com/eBay/fabio/wiki/Features#tcpsni-proxy-support),
[HTTPS upstream](https://github.com/eBay/fabio/wiki/Features#https-upstream-support),
[Websockets](https://github.com/eBay/fabio/wiki/Features#websocket-support),
[SSE](https://github.com/eBay/fabio/wiki/Features#sse---server-sent-events),
[dynamic reloading](https://github.com/eBay/fabio/wiki/Features#dynamic-reloading),
[traffic shaping](https://github.com/eBay/fabio/wiki/Features#traffic-shaping) for "blue/green" deployments,
[Circonus metrics](https://github.com/eBay/fabio/wiki/Features#metrics-support),
[Graphite metrics](https://github.com/eBay/fabio/wiki/Features#metrics-support),
[StatsD/DataDog metrics](https://github.com/eBay/fabio/wiki/Features#metrics-support),
[WebUI](https://github.com/eBay/fabio/wiki/Features#web-ui)
and [more](https://github.com/eBay/fabio/wiki/Features).

[Watch](https://www.youtube.com/watch?v=gf43TcWjBrE&list=PL81sUbsFNc5b-Gd59Lpz7BW0eHJBt0GvE&index=1) Kelsey Hightower demo Consul, Nomad, Vault and fabio at HashiConf EU 2016.

The full documentation is on the [Wiki](https://github.com/eBay/fabio/wiki).

## Getting started

1. Install from source, [binary](https://github.com/eBay/fabio/releases), [Docker](https://hub.docker.com/r/magiconair/fabio/) or [Homebrew](http://brew.sh).
    ```
	# go 1.8 or higher is required
    go get github.com/eBay/fabio                        (>= go1.8)

    brew install fabio                                  (OSX/macOS stable)
    brew install --devel fabio                          (OSX/macOS devel)

    docker pull magiconair/fabio                        (Docker)

    https://github.com/eBay/fabio/releases              (pre-built binaries)
    ```

2. Register your service in [consul](https://consul.io/).

   Make sure that each instance registers with a **unique ServiceID**.

3. Register a **health check** in consul as described [here](https://consul.io/docs/agent/checks.html).

   Make sure the health check is **passing** since fabio will only watch services
   which have a passing health check.

4. Register one `urlprefix-` tag per `host/path` prefix it serves, e.g.:

```
# HTTP/S examples
urlprefix-/css                   # path route
urlprefix-i.com/static           # host specific path route
urlprefix-mysite.com/            # host specific catch all route
urlprefix-/foo/bar strip=/foo    # route with path stripping (forward only '/bar' to upstream)

# TCP examples
urlprefix-:3306 proto=tcp        # route external port 3306
```

   Make sure the prefix for HTTP routes contains **at least one slash** (`/`).

5. Start fabio without a config file (assuming a running consul agent on `localhost:8500`)
   Watch the log output how fabio picks up the route to your service.
   Try starting/stopping your service to see how the routing table changes instantly.

6. Send all your HTTP traffic to fabio on port `9999`. 
   For TCP proxying see [TCP proxy](https://github.com/eBay/fabio/wiki/Features#tcp-proxy-support).

7. Done

## License

MIT licensed
