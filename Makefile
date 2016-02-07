
GO = GODEBUG=sbrk=1 GO15VENDOREXPERIMENT=1 go
GOFLAGS = -tags netgo -ldflags "-X main.version=$(shell git describe --tags)"

build:
	$(GO) build -i $(GOFLAGS)

test:
	$(GO) test ./...

linux:
	GOOS=linux GOARCH=amd64 $(GO) build -i $(GOFLAGS)

install:
	$(GO) install $(GOFLAGS)

release: test
	build/release.sh $(filter-out $@,$(MAKECMDGOALS))

docker: build test
	build/docker.sh

.PHONY: build linux install release docker test
