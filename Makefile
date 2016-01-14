
GO=~/go1.5.3/bin/go
GOFLAGS = -tags netgo -ldflags "-X main.version=$(shell git describe --tags)"

.PHONY: build
build:
	$(GO) build $(GOFLAGS)

.PHONY: linux
linux:
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS)

.PHONY: install
install:
	$(GO) install $(GOFLAGS)

.PHONY: release
release:
	build/release.sh $(filter-out $@,$(MAKECMDGOALS))

.PHONY: docker
docker: build
	build/docker.sh
