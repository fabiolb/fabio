---
title: "FastCGI Upstream"
---

To support FastCGI upstream add `proto=fcgi` option to the `urlprefix-` tag.

FastCGI upstreams support following configuration options:

 - `index`: Used to specify the index file that should be used if the request URL does not contain a
     file.
 - `root`: Document root of the FastCGI server.

Note that `index` and `root` can also be set in Fabio configuration as global default.

