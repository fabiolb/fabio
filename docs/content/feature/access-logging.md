---
title: "Access Logging"
since: "1.4.1"
---

Support for writing access logs for HTTP requests
in the [Common Log Format](https://en.wikipedia.org/wiki/Common_Log_Format)
or the [Combined Log Format](https://httpd.apache.org/docs/1.3/logs.html#combined)
or a custom format to stdout.

<!--more-->

By default, access logs are disabled. To enable them set `log.access.target=stdout`. This will
write access logs in the [Common Log Format](https://en.wikipedia.org/wiki/Common_Log_Format) to stdout. The
standard fabio logs are still written to stderr.

The log format can be controlled with the `log.access.format` parameter which
is either `common`, `combined` - which outputs the [Combined Log Format](https://httpd.apache.org/docs/1.3/logs.html#combined) - or a custom
format string which is fully described in
[fabio.properties](https://github.com/eBay/fabio/blob/master/fabio.properties#L374-L421).

```
# log.access.format configures the format of the access log.
#
# If the value is either 'common' or 'combined' then the logs are written in
# the Common Log Format or the Combined Log Format as defined below:
#
# 'common':   $remote_host - - [$time_common] "$request" $response_status $response_body_size
# 'combined': $remote_host - - [$time_common] "$request" $response_status $response_body_size "$header.Referer" "$header.User-Agent"
#
# Otherwise, the value is interpreted as a custom log format which is defined
# with the following parameters. Providing an empty format when logging is
# enabled is an error. To disable access logging leave the log.access.target
# value empty.
#
#   $header.<name>           - request http header (name: [a-zA-Z0-9-]+)
#   $remote_addr             - host:port of remote client
#   $remote_host             - host of remote client
#   $remote_port             - port of remote client
#   $request                 - request <method> <uri> <proto>
#   $request_args            - request query parameters
#   $request_host            - request host header (aka server name)
#   $request_method          - request method
#   $request_scheme          - request scheme
#   $request_uri             - request URI
#   $request_url             - request URL
#   $request_proto           - request protocol
#   $response_body_size      - response body size in bytes
#   $response_status         - response status code
#   $response_time_ms        - response time in S.sss format
#   $response_time_us        - response time in S.ssssss format
#   $response_time_ns        - response time in S.sssssssss format
#   $time_rfc3339            - log timestamp in YYYY-MM-DDTHH:MM:SSZ format
#   $time_rfc3339_ms         - log timestamp in YYYY-MM-DDTHH:MM:SS.sssZ format
#   $time_rfc3339_us         - log timestamp in YYYY-MM-DDTHH:MM:SS.ssssssZ format
#   $time_rfc3339_ns         - log timestamp in YYYY-MM-DDTHH:MM:SS.sssssssssZ format
#   $time_unix_ms            - log timestamp in unix epoch ms
#   $time_unix_us            - log timestamp in unix epoch us
#   $time_unix_ns            - log timestamp in unix epoch ns
#   $time_common             - log timestamp in DD/MMM/YYYY:HH:MM:SS -ZZZZ
#   $upstream_addr           - host:port of upstream server
#   $upstream_host           - host of upstream server
#   $upstream_port           - port of upstream server
#   $upstream_request_scheme - upstream request scheme
#   $upstream_request_uri    - upstream request URI
#   $upstream_request_url    - upstream request URL
#   $upstream_service        - name of the upstream service
#
# The default is
#
# log.access.format = common
```
