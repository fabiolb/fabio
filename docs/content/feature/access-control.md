---
title: "Access Control"
since: "1.5.8"
---

fabio supports basic ip centric access control per route.  You may
specify one of `allow` or `deny` options per route to control access.
Currently only source ip control is available.

<!--more-->

To allow access to a route from clients within the `192.168.1.0/24`
and `10.0.0.0/8` subnet you would add the following option:

```
allow=ip:192.168.1.0/24,ip:10.0.0.0/8
```

With this specified only clients sourced from those two subnets will
be allowed.  All other requests to that route will be denied.


Inversely, to only deny a specific set of clients you can use the
following option syntax:

```
deny=ip:1.2.3.4/32,100.123.0.0/16
```

With this configuration access will be denied to any clients with
the `1.2.3.4` address or coming from the `100.123.0.0/16` network.
