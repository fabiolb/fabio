---
title: "GRPC Proxy"
since: "1.5.11"
---

fabio can run a transparent GRPC proxy which dynamically forwards an incoming
RPC on a given port to services which advertise rpc service or method. To use GRPC
proxy support the service needs to advertise `urlprefix-/my.service/Method proto=grpc` in
Consul. In addition, fabio needs to be configured with a grpc listener:

```
fabio -proxy.addr ':1234;proto=grpc'
```

As per the HTTP/2 spec, the host header is not required, so host matching is not supported for GRPC proxying.

GRPC proxy support can be combined with [Certificate Stores](/feature/certificate-stores/) to provide TLS termination on fabio. Configure `proxy.addr` with `proto=grpcs`.

```
fabio -proxy.cs 'cs=ssl;type=path;path=/etc/ssl' -proxy.addr ':1234;proto=grpcs;cs=ssl'
```

To support TLS upstream servers add the `proto=grpcs` option to the
`urlprefix-` tag. The current implementation uses the clientca specified in the [Certificate Store](/feature/certificate-stores/) for the listener. To disable certificate
validation for a target set the `tlsskipverify=true` option.

```
urlprefix-/foo proto=grpcs
urlprefix-/foo proto=grpcs tlsskipverify=true
```

For TLS upstream servers (when using the consul registry) fabio will direct your traffic to an advertised service IP. If your service certificate does not contain an IP SAN, the certificate verification will fail. You can set the override the server name in the tls config by setting `grpcservername=<servername>` in the `urlprefix-` tag.

```
urlprefix-/ proto=grpcs grpcservername=my.service.hostname
```
