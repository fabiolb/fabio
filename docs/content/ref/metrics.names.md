---
title: "metrics.names"
---

`metrics.names` configures the template for the route metric names.
The value is expanded by the [text/template](https://golang.org/pkg/text/template) package and provides
the following variables:

* `Service`:   the service name
* `Host`:      the host part of the URL prefix
* `Path`:      the path part of the URL prefix
* `TargetURL`: the URL of the target

The following additional functions are defined:

* `clean`:     lowercase value and replace `.` and `:` with `_`

Given a route rule of

	route add testservice www.example.com/ http://10.1.2.3:12345/

the template variables are:

	.Service = testservice
	.Host = www.example.com
	.Path  = /
	.TargetURL.Host = 10.1.2.3:12345

which results to the following metric name when using the default template:

	testservice.www_example_com./.10_1_2_3_12345

The default is

	metrics.names = {{clean .Service}}.{{clean .Host}}.{{clean .Path}}.{{clean .TargetURL.Host}}
