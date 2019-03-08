---
title: "registry.backend"
---

`registry.backend` configures which backend is used.
Supported backends are: `consul`, `static`, `file`, `custom`. If custom is used fabio makes an api 
call to a remote system expecting the below json response

```json
[
 {
  "cmd": "string",
  "service": "string",
  "src": "string",
  "dest": "string",
  "weight": float,
  "tags": ["string"],
  "opts": {"string":"string"}
 }
]
```


The default is

	registry.backend = consul
