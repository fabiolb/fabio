---
title: "Test fabio with curl"
---

##### How do I send a request to fabio via `curl`?

```
curl -v -H 'Host: foo.com' 'http://localhost:9999/path'
```

The `-x` or `--proxy` options will most likely not work as you expect as they
send the full URL instead of just the request URI which usually does not match
any route but the default one - if configured.
