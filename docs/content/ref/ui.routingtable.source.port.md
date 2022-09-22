---
title: "ui.routingtable.source.port"
---

`ui.routingtable.source.port` configures an optional port for the routing table source column link.  This is used in conjunction with the [scheme](/ref/ui.routingtable.source.scheme/) and [host](/ref/ui.routingtable.source.host/).  

If the source is not a separate server (does not begin with '/', e.g. 'dev.google.net'), and the [host](/ref/ui.routingtable.source.host/) is set, this will use the port that is set, or default to the current scheme protocol port (80 for http or 443 for https).  

This is only applicable if the [linkenabled](/ref/ui.routingtable.source.linkenabled/) is set to true.

The default is

    ui.routingtable.source.port = 
