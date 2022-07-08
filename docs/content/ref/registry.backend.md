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
  "dst": "string",
  "weight": float,
  "tags": ["string"],
  "opts": {"string":"string"}
 }
]
```


The default is

	registry.backend = consul

Short description of the fields required for a custom backend

To configure routes Fabio uses a Config Language, specified [here](https://fabiolb.net/cfg/)

- cmd - the command to add, remove or change weight of a route. For example `route add` to add a new route mapping.
- service - the name that the service will show up in the UI.
- src - usually the prefix that will be used in the routing table.
- dst - the endpoint that will be used as the destination of the routing table. 
- weight - defines the weight of this path to perform routing. For example route A 90% and route B 10% for canary deployments.
- tags - a list of tags, provide a way to filter routes, making it easier to do operations like bulk deletes `route del tags "dev"`. 
- opts - a KV map of the config language list of options. for example `proto` or `prefix`
