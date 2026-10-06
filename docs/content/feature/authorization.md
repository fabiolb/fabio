---
title: "Authorization"
since: "1.5.11"
---

fabio supports basic and external HTTP authorization on a per-route basis.

<!--more-->

Authorization schemes are configured with the `proxy.auth` option.
You can configure one or multiple schemes.

Each authorization scheme is configured with a list of
key/value options.

    name=<name>;type=<type>;opt=arg;opt[=arg];...

Each scheme must have a **unique name** which is then
referenced in a route configuration.

    proxy.auth = name=myauth;type=...

When you configure the route, you must reference the unique name for the authorization scheme:

    route add svc / https://127.0.0.1:8080 auth=<name>

    urlprefix-/ proto=https auth=<name>

The following types of authorization schemes are available:

* [`basic`](#basic): legacy store for a single TLS and a set of client auth certificates
* [`external`](#external): delegate authorization to an HTTP endpoint

At the end you also find a list of [examples](#examples).

### Basic

The basic authorization scheme leverages [Http Basic Auth](https://en.wikipedia.org/wiki/Basic_access_authentication) and reads a [htpasswd](https://httpd.apache.org/docs/2.4/misc/password_encryptions.html) file at startup and credentials are cached until the service exits.

The `file` option contains the path to the htpasswd file. The `realm` parameter is optional (default is to use the `name`). The `refresh` option can set the htpasswd file refresh interval. Minimal refresh interval is `1s` to void busy loop. By default refresh is disabled i.e. set to zero.
Note: removing the htpasswd file will cause all requests to fail with HTTP status code 401 (Unauthorized) until the file is restored.

    name=<name>;type=basic;file=<file>;realm=<realm>;refresh=<interval>

Supported htpasswd formats are detailed [here](https://github.com/tg123/go-htpasswd)

### External

The external authorization scheme delegates each protected request to a fixed HTTP or HTTPS endpoint. Fabio sends a bodyless `GET` request to the configured `endpoint`, preserving the original request headers and adding `X-Forwarded-Method`, `X-Forwarded-Uri`, `X-Forwarded-Host`, `X-Forwarded-Proto`, `X-Forwarded-Port`, and client forwarding headers so the authorization service can evaluate the original request without consuming its body.

Any 2xx response authorizes the original request. `set-auth-headers` is a comma-separated list of response headers that replace the corresponding headers on the upstream request. `append-auth-headers` is a comma-separated list whose values are appended instead. Header lists containing commas must be quoted in `proxy.auth` configuration.

Non-2xx responses are returned to the client and the protected backend is not called. Redirects from the authorization endpoint are not followed by Fabio.

    name=<name>;type=external;endpoint=<absolute-http-or-https-url>;set-auth-headers=<headers>;append-auth-headers=<headers>

#### Examples

    # single basic auth scheme
    name=mybasicauth;type=basic;file=p/creds.htpasswd;

    # single basic auth scheme with refresh interval set to 30 seconds
    name=mybasicauth;type=basic;file=p/creds.htpasswd;refresh=30s

    # basic auth with multiple schemes
    proxy.auth = name=mybasicauth;type=basic;file=p/creds.htpasswd;refresh=30s,
                 name=myotherauth;type=basic;file=p/other-creds.htpasswd;realm=myrealm

    # oauth2-proxy authentication middleware
    proxy.auth = name=oauth;type=external;endpoint=http://oauth2-proxy:4180/oauth2/auth;set-auth-headers="X-Auth-Request-User,X-Auth-Request-Email";append-auth-headers=X-Auth-Request-Groups
    route add svc / http://127.0.0.1:8080 auth=oauth
