---
title: "Features"
weight: 200
---

Fabio supports the following:

 * [Access Control](/feature/access-control/) - route specific access control
 * [Access Logging](/feature/access-logging/) - customizable access logs
 * [Authorization](/feature/authorization/) - authorization mechanisms
 * [BGP Support](/feature/bgp) - bgp support
 * [Certificate Stores](/feature/certificate-stores/) - dynamic certificate stores like file system, HTTP server, [Consul](https://consul.io/) and [Vault](https://vaultproject.io/)
 * [Docker Support](/feature/docker/) - Official Docker image, Registrator and Docker Compose example
 * [Dynamic Reloading](/feature/dynamic-reloading/) - hot reloading of the routing table without downtime
 * [Graceful Shutdown](/feature/graceful-shutdown/) - wait until requests have completed before shutting down
 * [GRPC Proxy](/feature/grpc-proxy/) - Transparent GRPC proxy
 * [HTTP Compression](/feature/http-compression/) - GZIP compression for HTTP responses
 * [HTTP Header Support](/feature/http-headers/) - inject some HTTP headers into upstream requests
 * [HTTP Path Prepending](/feature/http-path-prepending/) - prepend a prefix path on to incoming requests
 * [HTTP Path Stripping](/feature/http-path-stripping/) - strip prefix paths from incoming requests
 * [HTTP redirects](/feature/http-redirects/) - redirect HTTP requests
 * [HTTPS TCP-SNI Proxy Support](/feature/https-tcp-sni-proxy/) - forward TLS connections based on hostname without re-encryption, or fallback to fabio terminating TLS and path routing as a fallback
 * [HTTPS Upstreams](/feature/https-upstream/) - forward requests to HTTPS upstream servers
 * [Metrics Support](/feature/metrics/) - support for Graphite, StatsD/DataDog and Circonus
 * [PROXY Protocol Support](/feature/proxy-protocol/) - support for HA Proxy PROXY protocol for inbound requests (use for Amazon ELB)
 * [Server-Sent Events/SSE](/feature/sse/) - support for Server-Sent Events/SSE
 * [TCP dynamic proxy](/feature/tcp-dynamic-proxy/) - TCP proxy based on urlprefix tag
 * [TCP Proxy Support](/feature/tcp-proxy/) - raw TCP proxy support
 * [TCP-SNI Proxy Support](/feature/tcp-sni-proxy/) - forward TLS connections based on hostname without re-encryption
 * [Traffic Shaping](/feature/traffic-shaping/) - forward N% of traffic upstream without knowing the number of instances
 * [Vault](/feature/vault/) - Store or generate certificates with HashiCorp Vault
 * [Web UI](/feature/web-ui/) - web ui to examine the current routing table
 * [Websocket Support](/feature/websockets/) - websocket support
