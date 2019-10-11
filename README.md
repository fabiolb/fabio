<p align="center">
  <p align="center" style="width: 50%; height: 64px;">
    <img src="https://cdn.rawgit.com/fabiolb/fabio/015e999/fabio.svg" height="64"/>
  </p>
  <p align="center" style="margin-top: 16px">
    <a href="http://ebay.github.io/"><img src="https://cdn.rawgit.com/fabiolb/fabio/7a02e1f/ebay.png" height="32" style="padding-right: 4px"/></a>
    <a href="http://www.ebayclassifiedsgroup.com"><img src="https://cdn.rawgit.com/fabiolb/fabio/7a02e1f/ecg.png" height="32"/></a>
    <a href="http://www.mytaxi.de"><img src="https://cdn.rawgit.com/fabiolb/fabio/7a02e1f/mytaxi.png" height="32"/></a>
    <a href="http://www.classmarkets.com"><img src="https://cdn.rawgit.com/fabiolb/fabio/7a02e1f/classmarkets.png" height="32"/></a>
  </p>
  <p align="center" style="margin-top: 16px">
    <a href="https://github.com/fabiolb/fabio/releases/latest"><img alt="Release" src="https://img.shields.io/github/release/fabiolb/fabio.svg?style=flat-square"></a>
    <a href="https://raw.githubusercontent.com/fabiolb/fabio/master/LICENSE"><img alt="License MIT" src="https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square"></a>
    <a href="https://app.codeship.com/projects/222209"><img alt="Codeship CI Status" src="https://img.shields.io/codeship/3e8307d0-2426-0135-1183-6e6f38f65fc4/master.svg?label=codeship&style=flat-square"></a>
    <a href="https://github.com/fabiolb/fabio/releases"><img alt="Downloads" src="https://img.shields.io/github/downloads/fabiolb/fabio/total.svg?style=flat-square"></a>
    <a href="https://hub.docker.com/r/magiconair/fabio/"><img alt="Docker Pulls magiconair" src="https://img.shields.io/docker/pulls/magiconair/fabio.svg?style=flat-square&label=docker+pulls+magiconair"></a>
    <a href="https://hub.docker.com/r/fabiolb/fabio/"><img alt="Docker Pulls fabiolb" src="https://img.shields.io/docker/pulls/fabiolb/fabio.svg?style=flat-square&label=docker+pulls+fabiolb"></a>
    <a href="#backers"><img alt="Backers on Open Collective" src="https://opencollective.com/fabio/backers/badge.svg"></a>
    <a href="#sponsors"><img alt="Sponsors on Open Collective" src="https://opencollective.com/fabio/sponsors/badge.svg"></a>
  </p>
</p>

---

#### Notes

1. If you are confused about the commit order for the v1.5.11 release please
   check the [Release Notes](https://github.com/fabiolb/fabio/releases/tag/v1.5.11)
   for an explanation.

1. The 1.5.11 tag was wrongly pointing to commit 0297494e9a00f87d3e387b8c6ff0408c2f5db6a0
   instead of commit 446fbba59da42ed73df67c3d738b9945dbf0790a. I have updated the v1.5.11
   tag to point to the correct version and created v1.5.11-wrong tag to point to the
   old (wrong) version.

---

fabio is a fast, modern, zero-conf load balancing HTTP(S) and TCP router
for deploying applications managed by [consul](https://consul.io/).

Register your services in consul, provide a health check and fabio will start
routing traffic to them. No configuration required. Deployment, upgrading and
refactoring has never been easier.

fabio is developed and maintained by The Fabio Authors.

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

It supports ([Full feature list](https://fabiolb.net/feature/))

* [TLS termination with dynamic certificate stores](https://fabiolb.net/feature/certificate-stores/)
* [Raw TCP proxy](https://fabiolb.net/feature/tcp-proxy/)
* [TCP+SNI proxy for full end-to-end TLS](https://fabiolb.net/feature/tcp-sni-proxy/) without decryption
* [HTTPS upstream support](https://fabiolb.net/feature/https-upstream/)
* [Websockets](https://fabiolb.net/feature/websockets/) and
  [SSE](https://fabiolb.net/feature/sse/)
* [Dynamic reloading without restart](https://fabiolb.net/feature/dynamic-reloading/)
* [Traffic shaping](https://fabiolb.net/feature/traffic-shaping/) for "blue/green" deployments,
* [Circonus](https://fabiolb.net/feature/metrics/),
  [Graphite](https://fabiolb.net/feature/metrics/) and
  [StatsD/DataDog](https://fabiolb.net/feature/metrics/) metrics
* [WebUI](https://fabiolb.net/feature/web-ui/)

[Watch](https://www.youtube.com/watch?v=gf43TcWjBrE&list=PL81sUbsFNc5b-Gd59Lpz7BW0eHJBt0GvE&index=1)
Kelsey Hightower demo Consul, Nomad, Vault and fabio at HashiConf EU 2016.

The full documentation is on [fabiolb.net](https://fabiolb.net/)

## Getting started

1. Install from source, [binary](https://github.com/fabiolb/fabio/releases),
   [Docker](https://hub.docker.com/r/fabiolb/fabio/) or [Homebrew](http://brew.sh).
    ```shell
	# go 1.9 or higher is required
    go get github.com/fabiolb/fabio                     (>= go1.9)

    brew install fabio                                  (OSX/macOS stable)
    brew install --devel fabio                          (OSX/macOS devel)

    docker pull fabiolb/fabio                           (Docker)

    https://github.com/fabiolb/fabio/releases           (pre-built binaries)
    ```

2. Register your service in [consul](https://consul.io/).

   Make sure that each instance registers with a **unique ServiceID** and a service name **without spaces**.

3. Register a **health check** in consul as described [here](https://consul.io/docs/agent/checks.html).

   By default fabio only watches services which have a **passing** health check, unless overriden with [registry.consul.service.status](https://fabiolb.net/ref/registry.consul.service.status/).

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
   For TCP proxying see [TCP proxy](https://fabiolb.net/feature/tcp-proxy/).

7. Done

## Maintainers

* Frank Schroeder [@magiconair](https://twitter.com/magiconair)

### Contributors

This project exists thanks to all the people who contribute. [[Contribute](CONTRIBUTING.md)].
<a href="https://github.com/fabiolb/fabio/graphs/contributors"><img src="https://opencollective.com/fabio/contributors.svg?width=890" /></a>


### Backers

Thank you to all our backers! 🙏 [[Become a backer](https://opencollective.com/fabio#backer)]

<a href="https://opencollective.com/fabio#backers" target="_blank"><img src="https://opencollective.com/fabio/backers.svg?width=890"></a>


### Sponsors

Support this project by becoming a sponsor. Your logo will show up here with a link to your website. [[Become a sponsor](https://opencollective.com/fabio#sponsor)]

<a href="https://opencollective.com/fabio/sponsor/0/website" target="_blank"><img src="https://opencollective.com/fabio/sponsor/0/avatar.svg"></a>
<a href="https://opencollective.com/fabio/sponsor/1/website" target="_blank"><img src="https://opencollective.com/fabio/sponsor/1/avatar.svg"></a>
<a href="https://opencollective.com/fabio/sponsor/2/website" target="_blank"><img src="https://opencollective.com/fabio/sponsor/2/avatar.svg"></a>
<a href="https://opencollective.com/fabio/sponsor/3/website" target="_blank"><img src="https://opencollective.com/fabio/sponsor/3/avatar.svg"></a>
<a href="https://opencollective.com/fabio/sponsor/4/website" target="_blank"><img src="https://opencollective.com/fabio/sponsor/4/avatar.svg"></a>
<a href="https://opencollective.com/fabio/sponsor/5/website" target="_blank"><img src="https://opencollective.com/fabio/sponsor/5/avatar.svg"></a>
<a href="https://opencollective.com/fabio/sponsor/6/website" target="_blank"><img src="https://opencollective.com/fabio/sponsor/6/avatar.svg"></a>
<a href="https://opencollective.com/fabio/sponsor/7/website" target="_blank"><img src="https://opencollective.com/fabio/sponsor/7/avatar.svg"></a>
<a href="https://opencollective.com/fabio/sponsor/8/website" target="_blank"><img src="https://opencollective.com/fabio/sponsor/8/avatar.svg"></a>
<a href="https://opencollective.com/fabio/sponsor/9/website" target="_blank"><img src="https://opencollective.com/fabio/sponsor/9/avatar.svg"></a>


## License

* Contributions up to 14 Apr 2017 before [38f73da](https://github.com/fabiolb/fabio/commit/38f73da6413b68fed1631101ac1d0b79a2fac870)

  MIT Licensed
  Copyright (c) 2017 eBay Software Foundation. All rights reserved.

* Contributions after 14 Apr 2017 starting with  [38f73da](https://github.com/fabiolb/fabio/commit/38f73da6413b68fed1631101ac1d0b79a2fac870)

  MIT Licensed
  Copyright (c) 2017-2019 Frank Schroeder. All rights reserved.

See [LICENSE](https://github.com/fabiolb/fabio/blob/master/LICENSE) for details.

## Stargazers over Time

[![Stargazers over time](https://starcharts.herokuapp.com/fabiolb/fabio.svg)](https://starcharts.herokuapp.com/fabiolb/fabio)
