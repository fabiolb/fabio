---
title: "Websockets"
since: "1.0.5"
---

fabio transparently supports Websocket connections by detecting the `Upgrade:
websocket` header in the incoming HTTP(S) request.

Websocket support has been implemented with the websocket library from
[golang.org/x/net/websocket](http://golang.org/x/net/websocket).

You can test the websocket support with the `demo/wsclient` and `demo/server` which
implements a simple echo server.

    ./server -addr 127.0.0.1:5000 -name ws-a -prefix /echo -proto ws
    ./wsclient -url ws://127.0.0.1:9999/echo

You can also run multiple web socket servers on different ports but the same endpoint.

fabio detects on whether to forward the request as HTTP or WS based on the
value of the `Upgrade` header. If the value is `websocket` it will attempt a
websocket connection to the target. Otherwise, it will fall back to HTTP.

One limitation of the current implementation is that the accepted set of
protocols has to be symmetric across all services handling it. Only the
following combinations will work reliably:

    svc-a and svc-b register /foo and accept only HTTP traffic there
    svc-a and svc-b register /foo and accept only WS traffic there
    svc-a and svc-b register /foo and accept both HTTP and WS traffic there

The following setup (or variations thereof) will not work reliably:

    svc-a registers /foo and accept only WS traffic there
    svc-b registers /foo and accept only HTTP traffic there

This is not a limitation of the routing itself but because the current
configuration does not provide fabio with enough information to make the
routing decision since the services do not advertise the protocols they handle
on a given endpoint.

This does not look like a big restriction but is also not difficult to extend
in a later version assuming there are use cases which require this behavior.
For now the services have to be symmetric in the protocols they accept.
