---
title: "proxy.flushinterval"
---

`proxy.flushinterval` configures periodic flushing of the
response buffer for SSE (server-sent events) connections.
They are detected when the `Accept` header is
`text/event-stream`.

The default is

    proxy.flushinterval = 1s
