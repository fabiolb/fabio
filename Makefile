
# do not specify a full path for go since travis will fail
GO = GOGC=off go
GOFLAGS = -ldflags "-X main.version=$(shell git describe --tags)"

all: build test

help:
	@echo "build     - go build"
	@echo "install   - go install"
	@echo "test      - go test"
	@echo "gofmt     - go fmt"
	@echo "linux     - go build linux/amd64"
	@echo "release   - build/release.sh"
	@echo "homebrew  - build/homebrew.sh"
	@echo "buildpkg  - build/build.sh"

build:
	$(GO) build -i $(GOFLAGS)

test:
	$(GO) test -i ./...
	$(GO) test -test.timeout 15s -v `go list ./... | grep -v '/vendor/'`

gofmt:
	gofmt -w `find . -type f -name '*.go' | grep -v vendor`

linux:
	GOOS=linux GOARCH=amd64 $(GO) build -i -tags netgo $(GOFLAGS)

install:
	$(GO) install $(GOFLAGS)

release: test
	build/release.sh

homebrew:
	build/homebrew.sh

buildpkg: test
	build/build.sh

.PHONY: build linux gofmt install release docker test homebrew buildpkg
