---
title: "Behind existing Gateway"
weight: 200
---

In the following setup fabio is configured receive all incoming traffic
from an existing gateway which also terminates SSL for one or more domains.

<pre>
                                                          +--> service-a
                                                          |
internet -- HTTP/HTTPS --> LB -- HTTP --> fabio -- HTTP --+--> service-b
                                                          |
                                                          +--> service-c
</pre>

Again, to scale fabio you can deploy it together with the frontend services
which provides high-availability and distributes the network bandwidth.

<pre>
                               +- HTTP -> fabio -+-> service-a (host-a)
                               |                 |
internet -- HTTP/HTTPS --> LB -+- HTTP -> fabio -+-> service-b (host-b)
                               |                 |
                               +- HTTP -> fabio -+-> service-c (host-c)
</pre>
