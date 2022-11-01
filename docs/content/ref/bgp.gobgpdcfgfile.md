---
title: "bgp.gobgpdcfgfile"
---

`bgp.gobgpdcfgfile` is the optional file path to a 
gobgpd [config file](https://github.com/osrg/gobgp/blob/master/docs/sources/configuration.md).
This overrides the global config
items, such as bgp.routerid, bgp.asn etc.  This also skips automatically adding gobgpd 
[policies](https://github.com/osrg/gobgp/blob/master/docs/sources/policy.md)
that restrict / disallow accepting prefixes from neighbors. Only use 
this if you know what you're doing, this is to allow
for more flexibility than we expose directly with fabio.

The default value is

	bgp.gobgpdcfgfile =

