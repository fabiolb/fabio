---
title: "proxy.header.trust.xff"
---

`proxy.header.trust.xff` instructs fabio to trust the X-Forwarded-For header
sent by the upstream client.

This should only be enabled (true) when fabio is running in a trusted
environment behind another upstream proxy.
When fabio is the authoritative internet facing load balancer you are
strongly encourged to set this to false allowing fabio to set the
value of the X-Forwarded-For header for the connecting client.

The legacy default is

    proxy.header.trust.xff = true
