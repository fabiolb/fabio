---
title: "HTTP Redirects"
since: "1.5.4"
---

To redirect an HTTP request to another URL you can use the `redirect=<code>` option. The `code` is the
HTTP status code used for the redirect response and must be between 300-399 for the route to be valid.

	# redirect /path to https://www.google.com/
	route add svc /path https://www.google.com/ opts "redirect=301"

To use the redirect with the `urlprefix-` tags you need to specify the target URL in after the code since
the target of the request is usually the address of the service that registers the tag.

	urlprefix-/path redirect=301,https://www.google.com/

If you want to include the original request URI in the redirect target append the `$path` pseudo-variable
to the target URL.

	urlprefix-/path redirect=303,https://www.foo.com$path

To redirect from HTTP to HTTPS you must include the `host:port` of the HTTP endpoint:

	route add example.com:80/ https://example.com/ opts "redirect=301"
