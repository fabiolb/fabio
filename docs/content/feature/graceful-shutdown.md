---
title: "Graceful Shutdown"
since: "1.0"
---

fabio supports a graceful shutdown timeout during which new requests will
receive a `503 Service Unavailable` response while the active requests can
complete. See the `proxy.shutdownwait` option in the
[fabio.properties](https://github.com/eBay/fabio/blob/master/fabio.properties)
file.
