FROM golang:1.24.0-alpine3.21 AS build

ARG consul_version=1.20.2
ADD https://releases.hashicorp.com/consul/${consul_version}/consul_${consul_version}_linux_amd64.zip /usr/local/bin
RUN cd /usr/local/bin && unzip consul_${consul_version}_linux_amd64.zip consul

ARG vault_version=1.18.4
ADD https://releases.hashicorp.com/vault/${vault_version}/vault_${vault_version}_linux_amd64.zip /usr/local/bin
RUN cd /usr/local/bin && unzip vault_${vault_version}_linux_amd64.zip

RUN apk update && apk add --no-cache git libcap
WORKDIR /src
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go test -trimpath -ldflags "-s -w" ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w"
RUN setcap cap_net_bind_service=+ep /src/fabio

FROM alpine:3.21
RUN apk update && apk add --no-cache ca-certificates
COPY --from=build /src/fabio /usr/bin
COPY --chown=nobody:nogroup fabio.properties /etc/fabio/fabio.properties
USER nobody:nogroup
EXPOSE 9998 9999
ENTRYPOINT ["/usr/bin/fabio"]
CMD ["-cfg", "/etc/fabio/fabio.properties"]
