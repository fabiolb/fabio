FROM alpine:3.2
ENTRYPOINT ["/bin/fabio"]

COPY . /go/src/github.com/eBay/fabio

RUN apk update
RUN apk add -t build-deps go git \
	&& cd /go/src/github.com/eBay/fabio \
	&& export GOPATH=/go \
	&& go get \
	&& go build -o	/bin/fabio  \
	&& rm -rf /go \
	&& apk del --purge build-deps
