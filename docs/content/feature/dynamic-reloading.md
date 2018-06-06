---
title: "Dynamic Reloading"
since: "1.0"
---

fabio builds the routing table from the Consul service registrations, health
check status and the user provided `route` commands stored in the Consul KV
store. This is **the** core feature of fabio - the reason it exists.

The cluster wide state is stored in the Consul Raft log which provides a
consistent view of the available and healthy services in the cluster. 

When the Raft log changes fabio is notified and downloads the list of
healthy services and the user defined routes from the KV store and re-builds
the routing table.

Once the new routing table has been built it is atomically swapped with the
active routing table without any service interruption. Existing connections
remain open and running requests are served even if the new routing table no
longer contains that route. 

Registering or de-registering a service, setting a node to maintenance mode,
failing or passing of a health check for a service, or writing data into the
Consul KV store all trigger an automatic reload of the fabio routing table for
all fabio nodes in the cluster.

This all happens automatically, with no downtime, or manual intervention.
