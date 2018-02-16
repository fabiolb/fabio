---
title: "fcgi.path.split"
---

`fcgi.path.split` specifies how to split the URL; the split value becomes the end of the first part
and anything in the URL after it becomes part of the `PATH_INFO` CGI variable.

Default value is

```
fcgi.path.split = .php
```
