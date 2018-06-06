---
title: "HTTPS Upstream"
since: "1.4.2"
---

To support HTTPS upstream servers add the `proto=https` option to the
`urlprefix-` tag. The current implementation requires that upstream
certificates need to be in the system root CA list. To disable certificate
validation for a target set the `tlsskipverify=true` option.

```
urlprefix-/foo proto=https
urlprefix-/foo proto=https tlsskipverify=true
```

