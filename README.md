# ![./fabio](https://github.com/eBay/fabio/blob/master/fabio.png)

##### Current stable version: 1.3.2

[![Build Status](https://travis-ci.org/eBay/fabio.svg?branch=master)](https://travis-ci.org/eBay/fabio)
[![License MIT](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/eBay/fabio/master/LICENSE)
[![Join the chat at https://gitter.im/eBay/fabio](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/ebay/fabio?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

fabio is a fast, modern, zero-conf load balancing HTTP(S) router
for deploying applications managed by [consul](https://consul.io/).

Register your services in consul, provide a health check and fabio will start routing traffic to them. No configuration required. Deployment, upgrading and refactoring has never been easier.

fabio was developed by [eBay in Amsterdam](http://www.ebayclassifiedsgroup.com) and runs some of the largest websites in The Netherlands and Italy. It delivers 15.000 req/sec every day since Sep 2015 without problems.

It integrates with
[Consul](https://consul.io/),
[Vault](https://vaultproject.io/),
[Amazon ELB](https://aws.amazon.com/elasticloadbalancing),
[Amazon API Gateway](https://aws.amazon.com/api-gateway/)
and more.

It supports
SSL,
[TCP+SNI proxy (full end-to-end TLS)](https://github.com/eBay/fabio/wiki/Features#tcpsni-proxy-support),
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
    GO15VENDOREXPERIMENT=1 go get github.com/eBay/fabio (go1.5.x)

    go get github.com/eBay/fabio                        (>= go1.6)

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

   * `urlprefix-/css`
   * `urlprefix-i.com/static`
   * `urlprefix-mysite.com/`

   Make sure the prefix contains **at least one slash** (`/`).

5. Start fabio without a config file (assuming a running consul agent on `localhost:8500`)
   Watch the log output how fabio picks up the route to your service.
   Try starting/stopping your service to see how the routing table changes instantly.

6. Send all your HTTP traffic to fabio on port `9999`

7. Done

## License

MIT licensed
