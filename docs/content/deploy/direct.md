---
title: "Direct"
weight: 100
---

In the following setup fabio is configured to listen on the public ip(s)
where it can optionally terminate SSL traffic for one or more domains - one ip per domain.

<pre>
                                           +--> service-a
                                           |
internet -- HTTP/HTTPS --> fabio -- HTTP --+--> service-b
                                           |
                                           +--> service-c
</pre>

To scale fabio you can deploy it together with the frontend services which provides
high-availability and distributes the network bandwidth.

<pre>
           +- HTTP/HTTPS -> fabio -+- HTTP -> service-a (host-a)
           |                       |
internet --+- HTTP/HTTPS -> fabio -+- HTTP -> service-b (host-b)
           |                       |
           +- HTTP/HTTPS -> fabio -+- HTTP -> service-c (host-c)
</pre>
