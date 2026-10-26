---
title: "Quickstart"
weight: 100
---


1. Install from source, binary, Docker or Homebrew.

	```
	go get github.com/fabiolb/fabio                     (>= go1.15)

	brew install fabio                                  (OSX/macOS stable)

	docker pull fabiolb/fabio                           (Docker)

	https://github.com/fabiolb/fabio/releases           (pre-built binaries)
	```

2. Register your service in Consul.

	Make sure that each instance registers with a unique `ServiceID` and a service name **without spaces**.

3. Register a health check in Consul as described [here](https://www.consul.io/docs/agent/checks.html).

	Make sure the health check is <button type="button" class="btn btn-xs
	btn-success">PASSING</button> since fabio will only watch services which
	have a passing health check.

4. Prefer Consul service metadata for route declarations. Give each route a
		unique metadata key starting with `urlprefix-` and place its
		`host/path` and options in the value. Add the exact `fabio` discovery
		tag to services using only metadata routes. For example, this Consul
		service registration includes a health check:

		```json
		{
		  "service": {
		    "id": "users-1",
		    "name": "users",
		    "address": "127.0.0.1",
		    "port": 8080,
		    "tags": ["fabio", "prod"],
		    "meta": {
		      "urlprefix-0": "/users",
		      "urlprefix-1": "example.com/ proto=https"
		    },
		    "check": {
		      "http": "http://127.0.0.1:8080/health",
		      "interval": "10s"
		    }
		  }
		}
		```

		Legacy `urlprefix-` service tags still work and may be combined with
		metadata routes. For example, a service handling `/user` and `/product`
		can use the two tags `urlprefix-/user` and `urlprefix-/product`
		without the `fabio` discovery tag.

		For new routes, use distinct metadata keys (such as `urlprefix-0` and
		`urlprefix-1`) with the configured `registry.consul.tagprefix`.
		Services without either a prefixed route tag or the `fabio` discovery
		tag cannot be found by the health-state watcher. To refresh metadata-only
		edits periodically, set `registry.consul.pollinterval` to a nonzero
		duration (for example, `5s`).

	fabio can forward HTTP, HTTPS and TCP traffic. Below are some configuration examples:

	```
	# HTTP/S examples
	# Make sure the prefix for HTTP routes contains at least one slash (/).
	urlprefix-/css                                     # path route
	urlprefix-i.com/static                             # host specific path route
	urlprefix-mysite.com/                              # host specific catch all route
	urlprefix-/foo/bar strip=/foo                      # path stripping (forward '/bar' to upstream)
	urlprefix-/bar prepend=/foo                        # path prepending (forward '/foo/bar' to upstream)
	urlprefix-/foo/bar proto=https                     # HTTPS upstream
	urlprefix-/foo/bar proto=https tlsskipverify=true  # HTTPS upstream and self-signed cert

	# TCP examples
	urlprefix-:3306 proto=tcp                          # route external port 3306
	urlprefix-:3306 proto=tcp pxyproto=true            # enables outbound PROXY protocol v1 headers
	
	# GRPC/S examples
	urlprefix-/my.service/Method proto=grpc                      # method specific route
	urlprefix-/my.service proto=grpc                             # service specific route
	urlprefix-/my.service proto=grpcs                            # TLS upstream
	urlprefix-/my.service proto=grpcs grpcservername=my.service  # TLS upstream with servername override
	urlprefix-/my.service proto=grpcs tlsskipverify=true         # TLS upstream and self-signed cert
	```

5. Start fabio without a config file

	```
	$ fabio
	```

	This assumes that a Consul agent is running on `localhost:8500`.

	Watch the log output how fabio picks up the route to your service.

	**Note:** For running fabio in Docker [look here](/feature/docker/).

6. Try starting/stopping your service to see how the routing table changes instantly.

7. Test that you can access the upstream service via fabio
	
	```
	# for urlprefix-/foo
	curl -i http://localhost:9999/foo

	# for urlprefix-mysite.com/foo
	curl -i -H 'Host: mysite.com' http://localhost:9999/foo

	```

8. Send all your HTTP traffic to fabio on port `9999`
