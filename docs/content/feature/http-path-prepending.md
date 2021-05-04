---
title: "HTTP Path Prepending"
since: "1.5.14"
---

fabio supports prepending a path to the incoming request. If you want to
forward `http://host/bar` as `http://host/foo/bar` you can add a `prepend=/foo`
option to the route options as `urlprefix-/bar prepend=/foo`.

Path prepending is done after path stripping. If you want to
forward `http://host/foo/bar` as `http://host/baz/bar` you can add
`prepend=/baz` and `strip=/foo` options to the route options as
`urlprefix-/bar prepend=/baz strip=/foo`.
