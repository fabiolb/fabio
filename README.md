<div>
  <div style="width: 50%; height: 64px;">
    <img src="https://cdn.rawgit.com/fabiolb/fabio/015e999/fabio.svg" height="64"/>
  </div>
  <div style="width: 50%; height: 64px; margin-top: 16px;">
    <a href="http://ebay.github.io/"><img src="https://cdn.rawgit.com/fabiolb/fabio/7a02e1f/ebay.png" height="32" style="padding-right: 4px"/></a>
    <a href="http://www.ebayclassifiedsgroup.com"><img src="https://cdn.rawgit.com/fabiolb/fabio/7a02e1f/ecg.png" height="32"/></a>
    <a href="http://www.mytaxi.de"><img src="https://cdn.rawgit.com/fabiolb/fabio/7a02e1f/mytaxi.png" height="32"/></a>
    <a href="http://www.classmarkets.com"><img src="https://cdn.rawgit.com/fabiolb/fabio/7a02e1f/classmarkets.png" height="32"/></a>
  </div>
</div>

##### Current stable version: 1.5.3

[![Codeship CI Status](https://img.shields.io/codeship/3e8307d0-2426-0135-1183-6e6f38f65fc4/master.svg?label=codeship&style=flat-square)](https://app.codeship.com/projects/222209)
[![Travis CI Status](https://img.shields.io/travis/fabiolb/fabio.svg?label=travis-ci&style=flat-square)](https://travis-ci.org/fabiolb/fabio)
[![License MIT](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://raw.githubusercontent.com/fabiolb/fabio/master/LICENSE)
[![Downloads](https://img.shields.io/github/downloads/fabiolb/fabio/total.svg?style=flat-square)](https://github.com/fabiolb/fabio/releases)
[![Docker Pulls magiconair](https://img.shields.io/docker/pulls/magiconair/fabio.svg?style=flat-square&label=docker+pulls+magiconair)](https://hub.docker.com/r/magiconair/fabio/)
[![Docker Pulls fabiolb](https://img.shields.io/docker/pulls/fabiolb/fabio.svg?style=flat-square&label=docker+pulls+fabiolb)](https://hub.docker.com/r/fabiolb/fabio/)

fabio is a fast, modern, zero-conf load balancing HTTP(S) and TCP router
for deploying applications managed by [consul](https://consul.io/).

Register your services in consul, provide a health check and fabio will start
routing traffic to them. No configuration required. Deployment, upgrading and
refactoring has never been easier.

fabio is developed and maintained by [Frank Schroeder](https://twitter.com/magiconair).

It powers some of the largest websites in
The Netherlands ([marktplaats.nl](http://www.marktplaats.nl/)),
Australia ([gumtree.com.au](http://www.gumtree.com.au))
and Italy ([www.kijiji.it](http://www.kijiji.it/)).
It delivers 23.000 req/sec every day since Sep 2015 without problems.

It integrates with
[Consul](https://consul.io/),
[Vault](https://vaultproject.io/),
[Amazon ELB](https://aws.amazon.com/elasticloadbalancing),
[Amazon API Gateway](https://aws.amazon.com/api-gateway/)
and more.

It supports ([Full feature list](https://github.com/fabiolb/fabio/wiki/Features))

* [TLS termination with dynamic certificate stores](https://github.com/fabiolb/fabio/wiki/Features#certificate-stores)
* [Raw TCP proxy](https://github.com/fabiolb/fabio/wiki/Features#tcp-proxy-support)
* [TCP+SNI proxy for full end-to-end TLS](https://github.com/fabiolb/fabio/wiki/Features#tcpsni-proxy-support) without decryption
* [HTTPS upstream support](https://github.com/fabiolb/fabio/wiki/Features#https-upstream-support)
* [Websockets](https://github.com/fabiolb/fabio/wiki/Features#websocket-support) and
  [SSE](https://github.com/fabiolb/fabio/wiki/Features#sse---server-sent-events)
* [Dynamic reloading without restart](https://github.com/fabiolb/fabio/wiki/Features#dynamic-reloading)
* [Traffic shaping](https://github.com/fabiolb/fabio/wiki/Features#traffic-shaping) for "blue/green" deployments,
* [Circonus](https://github.com/fabiolb/fabio/wiki/Features#metrics-support),
  [Graphite](https://github.com/fabiolb/fabio/wiki/Features#metrics-support) and
  [StatsD/DataDog](https://github.com/fabiolb/fabio/wiki/Features#metrics-support) metrics
* [WebUI](https://github.com/fabiolb/fabio/wiki/Features#web-ui)

[Watch](https://www.youtube.com/watch?v=gf43TcWjBrE&list=PL81sUbsFNc5b-Gd59Lpz7BW0eHJBt0GvE&index=1)
Kelsey Hightower demo Consul, Nomad, Vault and fabio at HashiConf EU 2016.

The full documentation is on the [Wiki](https://github.com/fabiolb/fabio/wiki).

## Getting started

1. Install from source, [binary](https://github.com/fabiolb/fabio/releases),
   [Docker](https://hub.docker.com/r/fabiolb/fabio/) or [Homebrew](http://brew.sh).
    ```shell
	# go 1.8 or higher is required
    go get github.com/fabiolb/fabio                     (>= go1.8)

    brew install fabio                                  (OSX/macOS stable)
    brew install --devel fabio                          (OSX/macOS devel)

    docker pull fabiolb/fabio                           (Docker)

    https://github.com/fabiolb/fabio/releases           (pre-built binaries)
    ```

2. Register your service in [consul](https://consul.io/).

   Make sure that each instance registers with a **unique ServiceID** and a service name **without spaces**.

3. Register a **health check** in consul as described [here](https://consul.io/docs/agent/checks.html).

   Make sure the health check is **passing** since fabio will only watch services
   which have a passing health check.

4. Register one `urlprefix-` tag per `host/path` prefix it serves, e.g.:

```
# HTTP/S examples
urlprefix-/css                                     # path route
urlprefix-i.com/static                             # host specific path route
urlprefix-mysite.com/                              # host specific catch all route
urlprefix-/foo/bar strip=/foo                      # path stripping (forward '/bar' to upstream)
urlprefix-/foo/bar proto=https                     # HTTPS upstream
urlprefix-/foo/bar proto=https tlsskipverify=true  # HTTPS upstream and self-signed cert

# TCP examples
urlprefix-:3306 proto=tcp                          # route external port 3306
```

   Make sure the prefix for HTTP routes contains **at least one slash** (`/`).

   See the full list of options in the [Documentation](https://github.com/fabiolb/fabio/wiki/Routing#config-language).

5. Start fabio without a config file (assuming a running consul agent on `localhost:8500`)
   Watch the log output how fabio picks up the route to your service.
   Try starting/stopping your service to see how the routing table changes instantly.

6. Send all your HTTP traffic to fabio on port `9999`.
   For TCP proxying see [TCP proxy](https://github.com/fabiolb/fabio/wiki/Features#tcp-proxy-support).

7. Done

## Maintainers

* Frank Schroeder [@magiconair](https://twitter.com/magiconair)

## License

* Contributions up to 14 Apr 2017 before [38f73da](https://github.com/fabiolb/fabio/commit/38f73da6413b68fed1631101ac1d0b79a2fac870)

  MIT Licensed
  Copyright (c) 2017 eBay Software Foundation. All rights reserved.

* Contributions after 14 Apr 2017 starting with  [38f73da](https://github.com/fabiolb/fabio/commit/38f73da6413b68fed1631101ac1d0b79a2fac870)

  MIT Licensed
  Copyright (c) 2017 Frank Schroeder. All rights reserved.

See [LICENSE](https://github.com/fabiolb/fabio/blob/master/LICENSE) for details.

## Stargazers over Time

[![Stargazers over time](https://starcharts.herokuapp.com/fabiolb/fabio.svg)](https://starcharts.herokuapp.com/fabiolb/fabio)
