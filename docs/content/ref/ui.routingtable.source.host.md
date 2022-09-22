---
title: "ui.routingtable.source.host"
---

`ui.routingtable.source.host` configures an optional host or base address for the link in the source column.

This is only used when the source is not a separate server (does not begin with '/', e.g. 'dev.google.net'). If source is subdirectory it will set the link for the source to this host.
If this is not set, and the source link is enabled, the link will default to current host.  

This is only applicable if the [linkenabled](/ref/ui.routingtable.source.linkenabled/) is set to true.

The default is

    ui.routingtable.source.host =
