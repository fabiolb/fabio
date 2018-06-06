---
title: "ui.addr"
---

`ui.addr` configures the address the UI is listening on.

The listener uses the same syntax as [proxy.addr](/ref/proxy.addr/) but
supports only a single listener. 

To enable HTTPS configure a certificate source. You should use a different
certificate source than the one you use for the external connections, e.g.
`cs=ui`.

The default is

	ui.addr = :9998
