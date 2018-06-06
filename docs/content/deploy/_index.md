---
title: "Deployment"
weight: 300
---

The main use-case for fabio is to distribute incoming HTTP(S) and TCP requests
from the internet to frontend (FE) services which can handle these requests.
In this scenario the FE services then use the service discovery feature in
[Consul](https://consul.io/) to find backend (BE) services they need in order
to serve the request.

That means that fabio is currently not used as an FE-BE or BE-BE router to
route traffic among the services themselves since the service discovery of
[Consul](https://consul.io/) already solves that problem. Having said that,
there is nothing that inherently prevents fabio from being used that way.
It just means that we are not doing it.

