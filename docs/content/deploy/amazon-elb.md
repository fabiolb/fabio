---
title: "Amazon ELB"
weight: 400
---

You can deploy fabio behind an Amazon ELB and enable [PROXY protocol support](http://docs.aws.amazon.com/ElasticLoadBalancing/latest/DeveloperGuide/enable-proxy-protocol.html) to get the remote address and port of the client.

<pre>
                                +- HTTP w/PROXY proto -> fabio -+-> service-a (host-a)
                                |                               |
internet -- HTTP/HTTPS --> ELB -+- HTTP w/PROXY proto -> fabio -+-> service-b (host-b)
                                |                               |
                                +- HTTP w/PROXY proto -> fabio -+-> service-c (host-c)
</pre>
