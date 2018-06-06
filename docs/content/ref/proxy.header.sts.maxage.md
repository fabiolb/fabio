---
title: "proxy.header.sts.maxage"
---

`proxy.header.sts.maxage` enables and configures the max-age of HSTS for TLS requests.
When set greater than zero this enables the Strict-Transport-Security header
and sets the max-age value in the header.

The default is

    proxy.header.sts.maxage = 0
