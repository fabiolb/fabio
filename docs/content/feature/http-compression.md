---
title: "HTTP Compression"
since: "1.3.4"
---

Enable dynamic compression of responses when the client sets the
`Accept-Encoding: gzip` header and the name of the requested file matches
a regular expression.

To configure which files should be compressed on the fly set configure
a regular expression in the `proxy.gzip.contenttype` property

```
# proxy.gzip.contenttype configures which responses should be compressed.
#
# By default, responses sent to the client are not compressed even if the
# client accepts compressed responses by setting the 'Accept-Encoding: gzip'
# header. By setting this value responses are compressed if the Content-Type
# header of the response matches and the response is not already compressed.
# The list of compressable content types is defined as a regular expression.
# The regular expression must follow the rules outlined in golang.org/pkg/regexp.
#
# A typical example is
#
# proxy.gzip.contenttype = ^(text/.*|application/(javascript|json|font-woff|xml)|.*\+(json|xml))(;.*)?$
#
# The default is
#
# proxy.gzip.contenttype =
```
