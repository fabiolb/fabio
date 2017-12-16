---
title: "Request Debugging"
since: "1.0"
---

To send a request from the command line via the fabio using `curl`
you should send it as follows:

```
curl -v -H 'Host: foo.com' 'http://localhost:9999/path'
```

The `-x` or `--proxy` options will most likely not work as you expect as they
send the full URL instead of just the request URI which usually does not match
any route but the default one - if configured.
