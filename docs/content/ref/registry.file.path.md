---
title: "registry.file.path"
---

`registry.file.path` configures a file based routing table.
The value configures the path to the file with the routing table.

#### Example
	registry.file.path = /home/zjj/route.txt
file content like is
```
route add svc / http://1.2.3.4:5000/
route add svc /test http://1.2.3.4:5001/
```
The default is

	registry.file.path =
