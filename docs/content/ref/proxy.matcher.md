---
title: "proxy.matcher"
---


`proxy.matcher` configures the path matching algorithm.

* `prefix`: prefix matching
* `glob`:  glob matching

When `prefix` matching is enabled then the route path must be a
prefix of the request URI, e.g. `/foo` matches `/foo`, `/foot` but
not `/fo`.

When `glob` matching is enabled the route is evaluated according to
globbing rules provided by the Go [`path.Match`](https://golang.org/pkg/path/#Match)
function.

For example, `/foo*` matches `/foo`, `/fool` and `/fools`. Also, `/foo/*/bar`
matches `/foo/x/bar`.

`iprefix` matching is similar to `prefix`, except it uses a case insensitive comparison

The default is

    proxy.matcher = prefix
