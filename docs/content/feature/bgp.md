---
title: "BGP"
since: "1.6.3"
---

NOTE: This feature does not work on Windows at present since the gobgp project
does not support windows.


This feature integrates the functionality of [gobgpd](https://github.com/osrg/gobgp)
with fabio.  This is particularly useful in the scenario where we are using
anycast IP addresses and want to dynamically advertise to upstream routers
when we're ready to receive traffic.  In the past, we've used external router
packages such as quagga or frr to handle this for us, but it's potentially
messy to make sure that the route advertisement stops if fabio goes down,
and the bgp daemon is started back up once fabio is running again.
By integrating the bgp advertisement with the proxy server, we've made
sure that when fabio goes down, the route is no longer advertised and 
traffic can be sent to other fabio instances accordingly.  When fabio is back up,
the route is advertised again.  

Further, the gobgp [command line client](https://github.com/osrg/gobgp/blob/master/docs/sources/cli-command-syntax.md) 
is fully supported by enabling 
the [bgp.enablegrpc](/ref/bgp.enablegrpc/) option.

[Multihop](/ref/bgp.multihop/) is supported, where fabio may not be
on the same subnet as neighbor.

To enable BGP, you must at a minimum:
* Set up an anycast interface on the host with a /32 address.  On linux, the dummy interface type is a good option 
  since it's supported using network manager.  Another option is hanging this address off of loopback.
* Configure the neighbor / peer / upstream router to allow us to peer, and to allow our anycast as a prefix it will 
  accept
* Set [bgp.enabled](/ref/bgp.enabled/)=true
* Configure the [bgp.asn](/ref/bgp.asn/) to be our router's Asn - probably use a private ASN here
* Configure the [bgp.routerid](/ref/bgp.routerid/) to be our router's IP address (i.e., not the anycast address, 
  something unique).  This will be the default nexthop of all routes we publish.
* Configure the [bgp.peers](/ref/bgp.peers/) for at least one nieghbor.
* Configure the [bgp.anycastaddresses](/ref/bgp.anycastaddresses) for at least one anycast address.

This will embed a gobgpd instance inside of fabio on startup and it will publish the configure anycast addresses.  
It will also configure a [gobgpd policy](https://github.com/osrg/gobgp/blob/master/docs/sources/policy.md)
that will reject all incoming prefixes from neighbors.

Alternatively, for more advanced use cases, you can reference an [external gobgpd config file](/ref/bgp.gobgpdcfgfile/) 
that will override many of the options set in the fabio config, including the policy
blocking us from accepting prefixes from neighbors.  You still need to specify the bgp.grpc
options from the fabio config since there is no analog in the gobgpd config file. 
You may still specify bgp anycastaddresses or bgp.peers from 
the fabio config, but we ignore anything
that would be specified in the global section of the gobgpd config file, including router ID and
the ASN.  Even If the bgp.gobgpdcfgfile value is set, fabio will still honor any values 
configured for bgp.anycastaddresses or bgp.peers.  These will be processed after the config
file is processed.


### Note
For situations where multiple fabio instances are running with the same anycast address
in the same datacenter, or in any other situation where the path distance
is the same and load balancing across multiple fabio instances is desired, 
the details of ECMP configuration is outside the scope of
this document as configuration would vary greatly depending on the 
details of the upstream router.

