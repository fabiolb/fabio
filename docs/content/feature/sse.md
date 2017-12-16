---
title: "Server Sent Events (SSE)"
since: "1.3"
---

fabio detects [SSE](http://www.w3.org/TR/eventsource/) connections if the
`Accept` header is set to `text/event-stream` and enables automatic flushing of
the response buffer to forward data to the client. The default is set to `1s`
and can be configured with the `proxy.flushinterval` parameter.
