---
title: "proxy.gzip.contenttype"
---

`proxy.gzip.contenttype` configures which responses should be compressed.

By default, responses sent to the client are not compressed even if the
client accepts compressed responses by setting the 'Accept-Encoding: gzip'
header. By setting this value responses are compressed if the `Content-Type`
header of the response matches and the response is not already compressed.
The list of compressable content types is defined as a regular expression.
The regular expression must follow the rules outlined in https://golang.org/pkg/regexp.

A typical example is

    proxy.gzip.contenttype = ^(text/.*|application/(javascript|json|font-woff|xml)|.*\+(json|xml))(;.*)?$

The default is

    proxy.gzip.contenttype =
