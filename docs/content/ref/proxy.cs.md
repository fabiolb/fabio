---
title: "proxy.cs"
---

`proxy.cs` configures one or more certificate sources.

Each certificate source is configured with a list of
key/value options. Each source must have a unique
name which can then be referred to in a listener
configuration.

    cs=<name>;type=<type>;opt=arg;opt[=arg];...

All certificates need to be provided in PEM format.

The following types of certificate sources are available:

#### File

The `file` certificate source supports one certificate which is loaded at
startup and is cached until the service exits.

The `cert` option contains the path to the certificate file. The `key`
option contains the path to the private key file. If the certificate file
contains both the certificate and the private key the `key` option can be
omitted. The `clientca` option contains the path to one or more client
authentication certificates.

    cs=<name>;type=file;cert=p/a-cert.pem;key=p/a-key.pem;clientca=p/clientAuth.pem

#### Path

The `path` certificate source loads certificates from a directory in
alphabetical order and refreshes them periodically.

The `cert` option provides the path to the TLS certificates and the
`clientca` option provides the path to the certificates for client
authentication.

TLS certificates are stored either in one or two files:

    www.example.com.pem or www.example.com-{cert,key}.pem

TLS certificates are loaded in alphabetical order and the first certificate
is the default for clients which do not support SNI.

The `refresh` option can be set to specify the refresh interval for the TLS
certificates. Client authentication certificates cannot be refreshed since
Go does not provide a mechanism for that yet.

The default refresh interval is 3 seconds and cannot be lower than 1 second
to prevent busy loops. To load the certificates only once and disable
automatic refreshing set `refresh` to zero.

    cs=<name>;type=path;cert=path/to/certs;clientca=path/to/clientcas;refresh=3s

#### HTTP

The `http` certificate source loads certificates from an HTTP/HTTPS server.

The `cert` option provides a URL to a text file which contains all files
that should be loaded from this directory. The filenames follow the same
rules as for the path source. The text file can be generated with:

    ls -1 *.pem > list

The `clientca` option provides a URL for the client authentication
certificates analogous to the `cert` option.

Authentication credentials can be provided in the URL as request parameter,
as basic authentication parameters or through a header.

The `refresh` option can be set to specify the refresh interval for the TLS
certificates. Client authentication certificates cannot be refreshed since
Go does not provide a mechanism for that yet.

The default refresh interval is 3 seconds and cannot be lower than 1 second
to prevent busy loops. To load the certificates only once and disable
automatic refreshing set `refresh` to zero.

    cs=<name>;type=http;cert=https://host.com/path/to/cert/list&token=123
    cs=<name>;type=http;cert=https://user:pass@host.com/path/to/cert/list
    cs=<name>;type=http;cert=https://host.com/path/to/cert/list;hdr=Authorization: Bearer 1234

#### Consul

The `consul` certificate source loads certificates from Consul.

The `cert` option provides a KV store URL where the the TLS certificates are
stored.

The `clientca` option provides a URL to a path in the KV store where the the
client authentication certificates are stored.

The filenames follow the same rules as for the path source.

The TLS certificates are updated automatically whenever the KV store
changes. The client authentication certificates cannot be updated
automatically since Go does not provide a mechanism for that yet.

    cs=<name>;type=consul;cert=http://localhost:8500/v1/kv/path/to/cert&token=123

#### Vault

The `vault` certificate store uses HashiCorp Vault as the certificate
store.

The `cert` option provides the path to the TLS certificates and the
`clientca` option provides the path to the certificates for client
authentication.

The `refresh` option can be set to specify the refresh interval for the TLS
certificates. Client authentication certificates cannot be refreshed since
Go does not provide a mechanism for that yet.

The default refresh interval is 3 seconds and cannot be lower than 1 second
to prevent busy loops. To load the certificates only once and disable
automatic refreshing set `refresh` to zero.

The path to vault must be provided in the VAULT_ADDR environment
variable. The token must be provided in the VAULT_TOKEN environment
variable.

    cs=<name>;type=vault;cert=secret/fabio/certs

#### Vault PKI

The `vault-pki` certificate store uses HashiCorp Vault's PKI backend to issue
certificates on-demand.

The `cert` option provides a PKI backend path for issuing certificates. The
`clientca` option works in the same way as for the generic Vault source.

The `refresh` option determines how long before the expiration date
certificates are re-issued. Values smaller than one hour are silently changed
to one hour, which is also the default.

    cs=<name>;type=vault-pki;cert=pki/issue/example-dot-com;refresh=24h;clientca=secret/fabio/client-certs

This source will issue server certificates on-demand using the PKI backend
and re-issue them 24 hours before they expire. The CA for client
authentication is expected to be stored at secret/fabio/client-certs.

#### Common options

All certificate stores support the following options:

    caupgcn: Upgrade a self-signed client auth certificate with this common-name
             to a CA certificate. Typically used for self-singed certificates
             for the Amazon AWS Api Gateway certificates which do not have the
             CA flag set which makes them unsuitable for client certificate
             authentication in Go. For the AWS Api Gateway set this value
             to `ApiGateway` to allow client certificate authentication.
             This replaces the deprecated parameter `aws.apigw.cert.cn`
             which was introduced in version 1.1.5.

#### Examples

    # file based certificate source
    proxy.cs = cs=some-name;type=file;cert=p/a-cert.pem;key=p/a-key.pem

    # path based certificate source
    proxy.cs = cs=some-name;type=path;path=path/to/certs

    # HTTP certificate source
    proxy.cs = cs=some-name;type=http;cert=https://user:pass@host:port/path/to/certs

    # Consul certificate source
    proxy.cs = cs=some-name;type=consul;cert=https://host:port/v1/kv/path/to/certs?token=abc123

    # Vault certificate source
    proxy.cs = cs=some-name;type=vault;cert=secret/fabio/certs

    # Vault PKI certificate source
    proxy.cs = cs=some-name;type=vault-pki;cert=pki/issue/example-dot-com

    # Multiple certificate sources
    proxy.cs = cs=srcA;type=path;path=path/to/certs,\
               cs=srcB;type=http;cert=https://user:pass@host:port/path/to/certs

    # path based certificate source for AWS Api Gateway
    proxy.cs = cs=some-name;type=path;path=path/to/certs;clientca=path/to/clientcas;caupgcn=ApiGateway

The default is

	proxy.cs =

