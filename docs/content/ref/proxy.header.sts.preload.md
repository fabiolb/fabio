---
title: "proxy.header.sts.preload"
---

`proxy.header.sts.preload` instructs HSTS to include the preload directive.
When set to true, the 'preload' option will be added to the
Strict-Transport-Security header.

Sending the preload directive from your site can have PERMANENT CONSEQUENCES
and prevent users from accessing your site and any of its subdomains if you
find you need to switch back to HTTP. Please read the details at
[https://hstspreload.org/#removal](https://hstspreload.org/#removal)
before sending the header with "preload".

The default is

    proxy.header.sts.preload = false
