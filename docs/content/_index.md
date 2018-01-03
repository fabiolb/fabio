---
title: "Overview"
---

Fabio is an HTTP and TCP reverse proxy that configures itself with data from
[Consul](https://consul.io/). 

Traditional load balancers and reverse proxies need to be configured with a
config file. The configuration contains the hostnames and paths the proxy is
forwarding to upstream services. This process can be automated with tools like
[consul-template](https://github.com/hashicorp/consul-template) that generate
config files and trigger a reload.

Fabio works differently since it updates its routing table directly from the
data stored in [Consul](https://consul.io/) as soon as there is a change and
without restart or reloading.

When you register a service in Consul all you need to add is a tag that
announces the paths the upstream service accepts, e.g. `urlprefix-/user` or
`urlprefix-/order` and fabio will do the rest.

### Maintainer

Fabio was developed and is maintained by Frank Schröder and the great community.

It was originally developed at the [eBay Classifieds Group](https://www.ebayclassifiedsgroup.com/) in Amsterdam, The Netherlands.
