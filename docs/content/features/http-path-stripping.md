---
title: "HTTP Path Stripping"
since: "1.3.7"
---

fabio supports stripping a path from the incoming request. If you want to
forward `http://host/foo/bar` as `http://host/bar` you can add a `strip=/foo`
option to the route options as `urlprefix-/foo/bar strip=/foo`.
