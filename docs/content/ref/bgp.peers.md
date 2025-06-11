---
title: "bgp.peers"
---

`bgp.peers` sets the bgp peers we will advertise routes to.  This is required if bgp is enabled.
bgp.peers is specified as a comma separated list of neighboraddress and asn pairs, i.e.

    bgp.peers = address=1.2.3.4;asn=65001,address=5.6.7.8;asn=65002

valid parameters for peers are:
 
    address              - required
    port                 - optional, defaults to 179
    asn                  - required
    multihop             - optional, defaults to false
    multihoplength       - optional, defaults to 2
    password             - optional

The default value is

	bgp.peers =

