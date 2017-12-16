



# registry.backend configures which backend is used.
# Supported backends are: consul, static, file
#
# The default is
#
# registry.backend = consul


# registry.timeout configures how long fabio tries to connect to the registry
# backend during startup.
#
# The default is
#
# registry.timeout = 10s


# registry.retry configures the interval with which fabio tries to
# connect to the registry during startup.
#
# The default is
#
# registry.retry = 500ms


# registry.static.routes configures a static routing table.
#
# Example:
#
#     registry.static.routes = \
#       route add svc / http://1.2.3.4:5000/
#
# The default is
#
# registry.static.routes =


# registry.static.noroutehtmlpath configures the KV path for the HTML of the
# noroutes page.
#
# The default is
#
# registry.static.noroutehtmlpath =


# registry.file.path configures a file based routing table.
# The value configures the path to the file with the routing table.
#
# The default is
#
# registry.file.path =


# registry.file.noroutehtmlpath configures the KV path for the HTML of the
# noroutes page.
#
# The default is
#
# registry.file.noroutehtmlpath =


# registry.consul.addr configures the address of the consul agent to connect to.
#
# The default is
#
# registry.consul.addr = localhost:8500


# registry.consul.token configures the acl token for consul.
#
# The default is
#
# registry.consul.token =


# registry.consul.kvpath configures the KV path for manual routes.
#
# The consul KV path is watched for changes which get appended to
# the routing table. This allows for manual overrides and weighted
# round-robin routes.
#
# The default is
#
# registry.consul.kvpath = /fabio/config

# registry.consul.noroutehtmlpath configures the KV path for the HTML of the
# noroutes page.
#
# The consul KV path is watched for changes.
#
# The default is
#
# registry.consul.noroutehtmlpath = /fabio/noroutes.html

# registry.consul.service.status configures the valid service status
# values for services included in the routing table.
#
# The values are a comma separated list of
# "passing", "warning", "critical" and "unknown"
#
# The default is
#
# registry.consul.service.status = passing


# registry.consul.tagprefix configures the prefix for tags which define routes.
#
# Services which define routes publish one or more tags with host/path
# routes which they serve. These tags must have this prefix to be
# recognized as routes.
#
# The default is
#
# registry.consul.tagprefix = urlprefix-


# registry.consul.register.enabled configures whether fabio registers itself in consul.
#
# Fabio will register itself in consul only if this value is set to "true" which
# is the default. To disable registration set it to any other value, e.g. "false"
#
# The default is
#
# registry.consul.register.enabled = true


# registry.consul.register.addr configures the address for the service registration.
#
# Fabio registers itself in consul with this host:port address.
# It must point to the UI/API endpoint configured by ui.addr and defaults to its
# value.
#
# The default is
#
# registry.consul.register.addr = :9998


# registry.consul.register.name configures the name for the service registration.
#
# Fabio registers itself in consul under this service name.
#
# The default is
#
# registry.consul.register.name = fabio


# registry.consul.register.tags configures the tags for the service registration.
#
# Fabio registers itself with these tags. You can provide a comma separated list of tags.
#
# The default is
#
# registry.consul.register.tags =


# registry.consul.register.checkInterval configures the interval for the health check.
#
# Fabio registers an http health check on http(s)://${ui.addr}/health
# and this value tells consul how often to check it.
#
# The default is
#
# registry.consul.register.checkInterval = 1s


# registry.consul.register.checkTimeout configures the timeout for the health check.
#
# Fabio registers an http health check on http(s)://${ui.addr}/health
# and this value tells consul how long to wait for a response.
#
# The default is
#
# registry.consul.register.checkTimeout = 3s


# registry.consul.register.checkTLSSkipVerify configures TLS verification for the health check.
#
# Fabio registers an http health check on http(s)://${ui.addr}/health
# and this value tells consul to skip TLS certificate validation for
# https checks.
#
# The default is
#
# registry.consul.register.checkTLSSkipVerify = false


# metrics.target configures the backend the metrics values are
# sent to.
#
# Possible values are:
#  <empty>:  do not report metrics
#  stdout:   report metrics to stdout
#  graphite: report metrics to Graphite on ${metrics.graphite.addr}
#  statsd: report metrics to StatsD on ${metrics.statsd.addr}
#  circonus: report metrics to Circonus (http://circonus.com/)
#
# The default is
#
# metrics.target =


# metrics.prefix configures the template for the prefix of all reported metrics.
#
# Each metric has a unique name which is hard-coded to
#
#    prefix.service.host.path.target-addr
#
# The value is expanded by the text/template package and provides
# the following variables:
#
#  - Hostname:  the Hostname of the server
#  - Exec:      the executable name of application
#
# The following additional functions are defined:
#
#  - clean:     lowercase value and replace '.' and ':' with '_'
#
# Template may include regular string parts to customize final prefix
#
# Example:
#
#  Server hostname: test-001.something.com
#  Binary executable name: fabio
#
#  The template variables are:
#
#  .Hostname =  test-001.something.com
#  .Exec = fabio
#
# which results to the following prefix string when using the
# default template:
#
#  test-001_something_com.fabio
#
# The default is
#
# metrics.prefix = {{clean .Hostname}}.{{clean .Exec}}


# metrics.names configures the template for the route metric names.
# The value is expanded by the text/template package and provides
# the following variables:
#
#  - Service:   the service name
#  - Host:      the host part of the URL prefix
#  - Path:      the path part of the URL prefix
#  - TargetURL: the URL of the target
#
# The following additional functions are defined:
#
#  - clean:     lowercase value and replace '.' and ':' with '_'
#
# Given a route rule of
#
#  route add testservice www.example.com/ http://10.1.2.3:12345/
#
# the template variables are:
#
#  .Service = testservice
#  .Host = www.example.com
#  .Path  = /
#  .TargetURL.Host = 10.1.2.3:12345
#
# which results to the following metric name when using the
# default template:
#
#  testservice.www_example_com./.10_1_2_3_12345
#
# The default is
#
# metrics.names = {{clean .Service}}.{{clean .Host}}.{{clean .Path}}.{{clean .TargetURL.Host}}


# metrics.interval configures the interval in which metrics are
# reported.
#
# The default is
#
# metrics.interval = 30s


# metrics.timeout configures how long fabio tries to connect to the metrics
# backend during startup.
#
# The default is
#
# metrics.timeout = 10s


# metrics.retry configures the interval with which fabio tries to
# connect to the metrics backend during startup.
#
# The default is
#
# metrics.retry = 500ms


# metrics.graphite.addr configures the host:port of the Graphite
# server. This is required when ${metrics.target} is set to "graphite".
#
# The default is
#
# metrics.graphite.addr =


# metrics.statsd.addr configures the host:port of the StatsD
# server. This is required when ${metrics.target} is set to "statsd".
#
# The default is
#
# metrics.statsd.addr =


# metrics.circonus.apikey configures the API token key to use when
# submitting metrics to Circonus. See: https://login.circonus.com/user/tokens
# This is required when ${metrics.target} is set to "circonus".
#
# The default is
#
# metrics.circonus.apikey =


# metrics.circonus.apiapp configures the API token app to use when
# submitting metrics to Circonus. See: https://login.circonus.com/user/tokens
# This is optional when ${metrics.target} is set to "circonus".
#
# The default is
#
# metrics.circonus.apiapp = fabio


# metrics.circonus.apiurl configures the API URL to use when
# submitting metrics to Circonus. https://api.circonus.com/v2/
# will be used if no specific URL is provided.
# This is optional when ${metrics.target} is set to "circonus".
#
# The default is
#
# metrics.circonus.apiurl =


# metrics.circonus.brokerid configures a specific broker to use when
# creating a check for submitting metrics to Circonus.
# This is optional when ${metrics.target} is set to "circonus".
# Optional for public brokers, required for Inside brokers.
# Only applicable if a check is being created.
#
# The default is
#
# metrics.circonus.brokerid =


# metrics.circonus.checkid configures a specific check to use when
# submitting metrics to Circonus.
# This is optional when ${metrics.target} is set to "circonus".
# An attempt will be made to search for a previously created check,
# if no applicable check is found, one will be created.
#
# The default is
#
# metrics.circonus.checkid =


# runtime.gogc configures GOGC (the GC target percentage).
#
# Setting runtime.gogc is equivalent to setting the GOGC
# environment variable which also takes precedence over
# the value from the config file.
#
# Increasing this value means fewer but longer GC cycles
# since there is more garbage to collect.
#
# The default of GOGC=100 works for Go 1.4 but shows
# a significant performance drop for Go 1.5 since the
# concurrent GC kicks in more often.
#
# During benchmarking I have found the following values
# to work for my setup and for now I consider them sane
# defaults for both Go 1.4 and Go 1.5.
#
# GOGC=100: Go 1.5 40% slower than Go 1.4
# GOGC=200: Go 1.5 == Go 1.4 with GOGC=100 (default)
# GOGC=800: both Go 1.4 and 1.5 significantly faster (40%/go1.4, 100%/go1.5)
#
# The default is
#
# runtime.gogc = 800


# runtime.gomaxprocs configures GOMAXPROCS.
#
# Setting runtime.gomaxprocs is equivalent to setting the GOMAXPROCS
# environment variable which also takes precedence over
# the value from the config file.
#
# If runtime.gomaxprocs < 0 then all CPU cores are used.
#
# The default is
#
# runtime.gomaxprocs = -1


# ui.access configures the access mode for the UI.
#
#  ro:  read-only access
#  rw:  read-write access
#
# The default is
#
# ui.access = rw


# ui.addr configures the address the UI is listening on.
# The listener uses the same syntax as proxy.addr but
# supports only a single listener. To enable HTTPS
# configure a certificate source. You should use
# a different certificate source than the one you
# use for the external connections, e.g. 'cs=ui'.
#
# The default is
#
# ui.addr = :9998


# ui.color configures the background color of the UI.
# Color names are from http://materializecss.com/color.html
#
# The default is
#
# ui.color = light-green


# ui.title configures an optional title for the UI.
#
# The default is
#
# ui.title =
