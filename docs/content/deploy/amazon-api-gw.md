---
title: "Amazon API Gateway"
weight: 500
---

You can deploy fabio as the target of an [Amazon API Gateway](https://aws.amazon.com/api-gateway/).

<pre>
internet -- HTTP/HTTPS --> API GW -+- HTTP -> fabio -+-> service-b (host-b)
</pre>

or behind an ELB with PROXY protocol support:

<pre>
                                           +- HTTP w/PROXY -> fabio -+-> service-a (host-a)
                                           |                         |
internet -- HTTP/HTTPS --> API GW --> ELB -+- HTTP w/PROXY -> fabio -+-> service-b (host-b)
                                           |                         |
                                           +- HTTP w/PROXY -> fabio -+-> service-c (host-c)
</pre>

You can authenticate calls from the API Gateway with a client certificate. This requires that you
configure an HTTPS listener on fabio with a valid certificate.

<pre>
internet -- HTTPS --> API GW -+- HTTPS w/client cert -> fabio -+-> service
</pre>

To enable fabio to validate the Amazon
generated certificate you need to configure the `aws.apigw.cert.cn` as follows:

    proxy.addr = 1.2.3.4:9999;your/cert.pem;your/key.pem;api-gw-cert.pem
    aws.apigw.cert.cn = ApiGateway

`api-gw-cert.pem` is the certificate generated in the AWS Management Console. `your/cert.pem` and `your/key.pem`
is the certificate/key pair for the HTTPS certificate. Since the Amazon API Gateway certificates don't have the `CA` flag set fabio needs to trust them for the client certificate authentication to work. Otherwise, you will get an `TLS handshake error: failed to verify client's certificate`. See [Issue 108](/eBay/fabio/issues/108) for details.

**Note:** The `aws.apigw.cert.cn` parameter will not be supported in version 1.2 and later which support dynamic certificate stores. You will have to add the `caupgcn=ApiGateway` parameter to the certificate source configuration instead. See [Certificate Stores](/feature/certificate-stores/) for more detail.
