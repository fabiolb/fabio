
# do not specify a full path for go since travis will fail
GO = GOGC=off go
GOFLAGS = -ldflags "-X main.version=$(shell git describe --tags)"
GOVENDOR = $(shell which govendor)

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
	@echo "pkg       - build, test and create pkg/fabio.tar.gz"
	@echo "clean     - remove temp files"

build: checkdeps
	$(GO) build -i $(GOFLAGS)
	$(GO) test -i ./...

test: checkdeps
	$(GO) test -v -test.timeout 15s `go list ./... | grep -v '/vendor/'`

checkdeps:
	[ -x "$(GOVENDOR)" ] || $(GO) get -u github.com/kardianos/govendor
	govendor list +e | grep '^ e ' && { echo "Found missing packages. Please run 'govendor add +e'"; exit 1; } || : echo

gofmt:
	gofmt -w `find . -type f -name '*.go' | grep -v vendor`

linux:
	GOOS=linux GOARCH=amd64 $(GO) build -i -tags netgo $(GOFLAGS)

install:
	$(GO) install $(GOFLAGS)

pkg: build test
	rm -rf pkg
	mkdir pkg
	tar czf pkg/fabio.tar.gz fabio

release: test
	build/release.sh

homebrew:
	build/homebrew.sh

codeship: ci-setup test

ci-setup:
	go version
	go env
	wget --version
	wget https://releases.hashicorp.com/consul/0.8.3/consul_0.8.3_linux_amd64.zip
	wget https://releases.hashicorp.com/vault/0.6.4/vault_0.6.4_linux_amd64.zip
	mkdir $(GOPATH)/bin/
	unzip -d $(GOPATH)/bin consul_0.8.3_linux_amd64.zip
	unzip -d $(GOPATH)/bin vault_0.6.4_linux_amd64.zip
	vault --version
	consul --version

buildpkg: test
	build/build.sh

clean:
	$(GO) clean
	rm -rf pkg

.PHONY: build linux gofmt install release docker test homebrew buildpkg pkg clean
