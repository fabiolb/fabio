---
title: "Quickstart"
weight: 100
---


1. Install from source, binary, Docker or Homebrew.

	```
	go get github.com/fabiolb/fabio                     (>= go1.8)

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

4. Routes are stored in Consul [Service Tags](https://www.consul.io/docs/agent/services.html)
and you need to add a separate `urlprefix-` tag for every `host/path` prefix the service serves.
	
	For example, if your service handles `/user` and `/product` then add two tags `urlprefix-/user` and `urlprefix-/product`. 
	You can register as many prefixes as you want.

	fabio can forward HTTP, HTTPS and TCP traffic. Below are some configuration examples:

	```
	# HTTP/S examples
	# Make sure the prefix for HTTP routes contains at least one slash (/).
	urlprefix-/css                                     # path route
	urlprefix-i.com/static                             # host specific path route
	urlprefix-mysite.com/                              # host specific catch all route
	urlprefix-/foo/bar strip=/foo                      # path stripping (forward '/bar' to upstream)
	urlprefix-/foo/bar proto=https                     # HTTPS upstream
	urlprefix-/foo/bar proto=https tlsskipverify=true  # HTTPS upstream and self-signed cert

	# TCP examples
	urlprefix-:3306 proto=tcp                          # route external port 3306
	
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
