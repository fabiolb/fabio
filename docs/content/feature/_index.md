---
title: "Features"
weight: 200
---

The following list provides a list of features supported by fabio. 

 * [Access Logging](/feature/access-logging/) - customizable access logs
 * [Access Control](/feature/access-control/) - route specific access control
 * [Certificate Stores](/feature/certificate-stores/) - dynamic certificate stores like file system, HTTP server, [Consul](https://consul.io/) and [Vault](https://vaultproject.io/)
 * [Compression](/feature/compression/) - GZIP compression for HTTP responses
 * [Docker Support](/feature/docker/) - Official Docker image, Registrator and Docker Compose example
 * [Dynamic Reloading](/feature/dynamic-reloading/) - hot reloading of the routing table without downtime
 * [Graceful Shutdown](/feature/graceful-shutdown/) - wait until requests have completed before shutting down
 * [HTTP Header Support](/feature/http-headers/) - inject some HTTP headers into upstream requests
 * [HTTPS Upstreams](/feature/https-upstream/) - forward requests to HTTPS upstream servers
 * [Metrics Support](/feature/metrics/) - support for Graphite, StatsD/DataDog and Circonus
 * [PROXY Protocol Support](/feature/proxy-protocol/) - support for HA Proxy PROXY protocol for inbound requests (use for Amazon ELB)
 * [Path Stripping](/feature/path-stripping/) - strip prefix paths from incoming requests
 * [Server-Sent Events/SSE](/feature/sse/) - support for Server-Sent Events/SSE
 * [TCP Proxy Support](/feature/tcp-proxy/) - raw TCP proxy support
 * [TCP-SNI Proxy Support](/feature/tcp-sni-proxy/) - forward TLS connections based on hostname without re-encryption
 * [Traffic Shaping](/feature/traffic-shaping/) - forward N% of traffic upstream without knowing the number of instances
 * [Web UI](/feature/web-ui/) - web ui to examine the current routing table
 * [Websocket Support](/feature/websockets/) - websocket support
