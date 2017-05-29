FROM golang:1.8-alpine
RUN apk add --update --no-cache git make openssl unzip
RUN go get github.com/kardianos/govendor
RUN go get github.com/fabiolb/fabio
RUN wget https://releases.hashicorp.com/consul/0.8.3/consul_0.8.3_linux_amd64.zip
RUN wget https://releases.hashicorp.com/vault/0.6.5/vault_0.6.5_linux_amd64.zip
RUN unzip -d /usr/bin consul_0.8.3_linux_amd64.zip
RUN unzip -d /usr/bin vault_0.6.5_linux_amd64.zip
RUN rm *.zip
RUN apk del unzip openssl
WORKDIR /go/src/github.com/fabiolb/fabio
ENTRYPOINT git fetch origin && git checkout ${BRANCH:=master} && make test
