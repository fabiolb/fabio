FROM golang:1.11.5-alpine AS build

ARG consul_version=1.4.2
ADD https://releases.hashicorp.com/consul/${consul_version}/consul_${consul_version}_linux_amd64.zip /usr/local/bin
RUN cd /usr/local/bin && unzip consul_${consul_version}_linux_amd64.zip

ARG vault_version=1.0.3
ADD https://releases.hashicorp.com/vault/${vault_version}/vault_${vault_version}_linux_amd64.zip /usr/local/bin
RUN cd /usr/local/bin && unzip vault_${vault_version}_linux_amd64.zip

WORKDIR /go/src/github.com/fabiolb/fabio
COPY . .
ENV GO111MODULE=on
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go test -mod=vendor -ldflags "-s -w" ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags "-s -w"

FROM alpine:3.8
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=build /go/src/github.com/fabiolb/fabio/fabio /usr/bin
ADD fabio.properties /etc/fabio/fabio.properties
EXPOSE 9998 9999
ENTRYPOINT ["/usr/bin/fabio"]
CMD ["-cfg", "/etc/fabio/fabio.properties"]
