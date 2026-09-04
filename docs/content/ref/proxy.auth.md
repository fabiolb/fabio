---
title: "proxy.auth"
---

`proxy.auth` configures one or more authorization schemes.

Each authorization scheme is configured with a list of
key/value options. Each scheme must have a unique
name which can then be referred to in a routing rule.

    name=<name>;type=<type>;opt=arg;opt[=arg];...

The following types of authorization schemes are available:

#### Basic

The basic authorization scheme leverages [Http Basic Auth](https://en.wikipedia.org/wiki/Basic_access_authentication) and reads a [htpasswd](https://httpd.apache.org/docs/2.4/misc/password_encryptions.html) file at startup and credentials are cached until the service exits.

The `file` option contains the path to the htpasswd file. The `realm` parameter is optional (default is to use the `name`). The `refresh` option can set the htpasswd file refresh interval. Minimal refresh interval is `1s` to void busy loop. By default refresh is disabled i.e. set to zero.
Note: removing the htpasswd file will cause all requests to fail with HTTP status code 401 (Unauthorized) until the file is restored.

    name=<name>;type=basic;file=<file>;realm=<realm>;refresh=<interval>

Supported htpasswd formats are detailed [here](https://github.com/tg123/go-htpasswd)

#### External

The external authorization scheme sends a bodyless `GET` request to a fixed absolute HTTP or HTTPS `endpoint` before forwarding the protected request. The authorization request includes the original headers plus forwarding metadata such as `X-Forwarded-Method` and `X-Forwarded-Uri`.

All 2xx responses authorize the request. `set-auth-headers` selects response headers that replace the same headers on the upstream request, while `append-auth-headers` selects response headers whose values are appended. Quote comma-separated header lists in `proxy.auth` configuration.

Non-2xx responses, including redirects, are returned directly to the client without calling the protected backend. Fabio does not follow authorization endpoint redirects.

    name=<name>;type=external;endpoint=<absolute-http-or-https-url>;set-auth-headers=<headers>;append-auth-headers=<headers>

#### Examples

    # single basic auth scheme
    name=mybasicauth;type=basic;file=p/creds.file;

    # single basic auth scheme with refresh interval set to 30 seconds
    name=mybasicauth;type=basic;file=p/creds.htpasswd;refresh=30s

    # basic auth with multiple schemes
    proxy.auth = name=mybasicauth;type=basic;file=p/creds.htpasswd;refresh=30s,
                 name=myotherauth;type=basic;file=p/other-creds.htpasswd;realm=myrealm

    # oauth2-proxy authentication middleware
    proxy.auth = name=oauth;type=external;endpoint=http://oauth2-proxy:4180/oauth2/auth;set-auth-headers="X-Auth-Request-User,X-Auth-Request-Email";append-auth-headers=X-Auth-Request-Groups

The default is

    proxy.auth =
