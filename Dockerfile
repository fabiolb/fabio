FROM golang:alpine AS build

ARG TARGETARCH
ARG consul_version=1.21.5
ADD https://releases.hashicorp.com/consul/${consul_version}/consul_${consul_version}_linux_${TARGETARCH}.zip /usr/local/bin
RUN cd /usr/local/bin && unzip consul_${consul_version}_linux_${TARGETARCH}.zip consul

ARG vault_version=1.20.4
ADD https://releases.hashicorp.com/vault/${vault_version}/vault_${vault_version}_linux_${TARGETARCH}.zip /usr/local/bin
RUN cd /usr/local/bin && unzip vault_${vault_version}_linux_${TARGETARCH}.zip vault

RUN apk update && apk add --no-cache git ca-certificates libcap
WORKDIR /src
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go test -trimpath -ldflags "-s -w" ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -trimpath -ldflags "-s -w"
RUN setcap cap_net_bind_service=+ep /src/fabio
RUN echo "nobody:x:65534:65534:nobody:/:/sbin/nologin" > /passwd
RUN echo "nogroup:x:65533:" > /group

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /src/fabio /usr/bin/
COPY --from=build /passwd /etc/
COPY --from=build /group /etc/
ADD --chown=nobody:nogroup fabio.properties /etc/fabio/fabio.properties
USER nobody:nogroup
EXPOSE 9998 9999
ENTRYPOINT ["/usr/bin/fabio"]
CMD ["-cfg", "/etc/fabio/fabio.properties"]
