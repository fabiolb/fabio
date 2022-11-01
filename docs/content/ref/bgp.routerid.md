---
title: "bgp.routerid"
---

`bgp.routerid` is the router id (ip address) of this router.  
This is required if bgp is enabled.  This should be the unique IP
address, not any anycast.  This will also be used as
the default nexthop address unless [bgp.nexthop](/ref/bgp.nexthop/)
is specified.

The default value is

	bgp.routerid =
