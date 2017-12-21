---
title: Request Tracing
---

#### How do I see which routes fabio is matching for a request?

To trace how a request is routed you can add a `Trace` header with an non-
empty value which is truncated at 16 characters to keep the log output short.

```
$ curl -v -H 'Trace: abc' -H 'Host: foo.com' 'http://localhost:9999/bar/baz'

2015/09/28 21:56:26 [TRACE] abc Tracing foo.com/bar/baz
2015/09/28 21:56:26 [TRACE] abc No match foo.com/bang
2015/09/28 21:56:26 [TRACE] abc Match foo.com/
2015/09/28 22:01:34 [TRACE] abc Routing to http://1.2.3.4:8080/
```

