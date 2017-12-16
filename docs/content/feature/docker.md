---
title: "Docker Support"
since: "1.0"
---

To run fabio within Docker use the official Docker image `fabiolb/fabio` and
mount your own config file to `/etc/fabio/fabio.properties`

    docker run -d -p 9999:9999 -p 9998:9998 -v $PWD/fabio/fabio.properties:/etc/fabio/fabio.properties fabiolb/fabio

If you want to run the Docker image with one or more SSL certificates then
you can store your configuration and certificates in `/etc/fabio` and mount
the entire directory, e.g.

    $ cat ~/fabio/fabio.properties
    proxy.addr=:443;/etc/fabio/ssl/mycert.pem;/etc/fabio/ssl/mykey.pem

    docker run -d -p 443:443 -p 9998:9998 -v $PWD/fabio:/etc/fabio fabiolb/fabio

The official Docker image contains the root CA certificates from a recent and updated
Ubuntu 12.04.5 LTS installation.

### Registrator

If you use Gliderlabs [Registrator](https://github.com/gliderlabs/registrator) to register your services
you can pass the `urlprefix-` tags via the `SERVICE_TAGS` environment variable as follows:

```
$ docker run -d \
    --name=registrator \
    --net=host \        
    --volume=/var/run/docker.sock:/tmp/docker.sock \
    gliderlabs/registrator:latest \
    consul://localhost:8500

$ docker run -d -p 80:8000 \
    -e SERVICE_8000_CHECK_HTTP=/foo/healthcheck  \
    -e SERVICE_8000_NAME=foo \
    -e SERVICE_CHECK_INTERVAL=10s \
    -e SERVICE_CHECK_TIMEOUT=5s  \
    -e SERVICE_TAGS=urlprefix-/foo \
    test/foo
```

### Docker Compose

If you are using [Docker compose](https://docs.docker.com/compose/) you can add the `SERVICE_TAGS`
to the `environment` section as follows:

    bar:
      environment:
        - SERVICE_TAGS=urlprefix-/bar

