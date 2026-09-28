---
title: "Amazon ELB and NLB"
weight: 400
---

You can deploy fabio behind an Amazon Classic Load Balancer and enable [PROXY protocol support](http://docs.aws.amazon.com/ElasticLoadBalancing/latest/DeveloperGuide/enable-proxy-protocol.html) to pass the remote address and port of the client using version 1. For an Amazon Network Load Balancer (NLB), enable PROXY protocol v2 on the target group and set `pxyproto=true` on the fabio listener.

The listener accepts either PROXY protocol version, so the same fabio routing configuration can be used with both load balancers.

<pre>
                                +- HTTP/TCP w/PROXY v1 or v2 -> fabio -+-> service-a (host-a)
                                |                               |
internet -- HTTP/HTTPS/TCP --> ELB or NLB -+- HTTP/TCP w/PROXY v1 or v2 -> fabio -+-> service-b (host-b)
                                |                               |
                                +- HTTP/TCP w/PROXY v1 or v2 -> fabio -+-> service-c (host-c)
</pre>
