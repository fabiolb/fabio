---
title: "metrics.prefix"
---

`metrics.prefix` configures the template for the prefix of all reported metrics.

Each metric has a unique name which is hard-coded to

	prefix.service.host.path.target-addr

The value is expanded by the text/template package and provides
the following variables:

* `Hostname`:  the Hostname of the server
* `Exec`:      the executable name of application

The following additional functions are defined:

* `clean`:     lowercase value and replace `.` and `:` with `_`

Template may include regular string parts to customize final prefix

#### Example

Server hostname: `test-001.something.com`
Binary executable name: `fabio`

The template variables are:

	.Hostname =  test-001.something.com
	.Exec = fabio

which results to the following prefix string when using the default template:

	test-001_something_com.fabio

The default is

	metrics.prefix = {{clean .Hostname}}.{{clean .Exec}}
