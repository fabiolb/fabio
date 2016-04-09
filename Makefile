
# do not specify a full path for go since travis will fail
GO = GOGC=off go
GOFLAGS = -tags netgo -ldflags "-X main.version=$(shell git describe --tags)"

all: build test

build:
	$(GO) build -i $(GOFLAGS)

test:
	$(GO) test -i ./...
	$(GO) test -test.timeout 5s `go list ./... | grep -v '/vendor/'`

gofmt:
	gofmt -w `find . -type f -name '*.go' | grep -v vendor`

linux:
	GOOS=linux GOARCH=amd64 $(GO) build -i $(GOFLAGS)

install:
	$(GO) install $(GOFLAGS)

release: test
	build/release.sh $(filter-out $@,$(MAKECMDGOALS))

docker: build test
	build/docker.sh

.PHONY: build linux gofmt install release docker test
