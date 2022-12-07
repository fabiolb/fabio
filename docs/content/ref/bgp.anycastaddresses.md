---
title: "bgp.anycastaddresses"
---

`bgp.anycastaddresses` sets the anycast addresses we will advertise, 
separated by comma.  Technically this will advertise any route prefix.  
These should already be configured on the host probably hung off loopback.
 For example, 192.168.5.3/32.

The default value is

	bgp.anycastaddresses =

If bgp is enabled, this must be defined.
