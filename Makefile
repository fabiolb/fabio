# CUR_TAG is the last git tag plus the delta from the current commit to the tag
# e.g. v1.5.5-<nr of commits since>-g<current git sha>
CUR_TAG = $(shell git describe)

# LAST_TAG is the last git tag
# e.g. v1.5.5
LAST_TAG = $(shell git describe --abbrev=0)

# VERSION is the last git tag without the 'v'
# e.g. 1.5.5
VERSION = $(shell git describe --abbrev=0 | cut -c 2-)

# GO runs the go binary with garbage collection disabled for faster builds.
# Do not specify a full path for go since travis will fail.
GO = GOGC=off go

# GOFLAGS is the flags for the go compiler. Currently, only the version number is
# passed to the linker via the -ldflags.
GOFLAGS = -ldflags "-X main.version=$(CUR_TAG)"

# GOVERSION is the current go version, e.g. go1.9.2
GOVERSION = $(shell go version | awk '{print $$3;}')

# GORELEASER is the path to the goreleaser binary.
GORELEASER = $(shell which goreleaser)

# GOVENDOR is the path to the govendor binary.
GOVENDOR = $(shell which govendor)

# VENDORFMT is the path to the vendorfmt binary.
VENDORFMT = $(shell which vendorfmt)

# all is the default target
all: build test

# help prints a help screen
help:
	@echo "build     - go build"
	@echo "install   - go install"
	@echo "test      - go test"
	@echo "gofmt     - go fmt"
	@echo "vet       - go vet"
	@echo "linux     - go build linux/amd64"
	@echo "release   - tag, build and publish release with goreleaser"
	@echo "pkg       - build, test and create pkg/fabio.tar.gz"
	@echo "clean     - remove temp files"

# build compiles fabio and the test dependencies
build: checkdeps vendorfmt gofmt
	$(GO) build -i $(GOFLAGS)
	$(GO) test -i ./...

# test runs the tests
test: checkdeps vendorfmt vet gofmt
	$(GO) test -v -test.timeout 15s `go list ./... | grep -v '/vendor/'`

# checkdeps ensures that all required dependencies are vendored in
checkdeps:
	[ -x "$(GOVENDOR)" ] || $(GO) get -u github.com/kardianos/govendor
	govendor list +e | grep '^ e ' && { echo "Found missing packages. Please run 'govendor add +e'"; exit 1; } || : echo

# vendorfmt ensures that the vendor/vendor.json file is formatted correctly
vendorfmt:
	[ -x "$(VENDORFMT)" ] || $(GO) get -u github.com/magiconair/vendorfmt/cmd/vendorfmt
	vendorfmt

# gofmt runs gofmt on the code
gofmt:
	gofmt -s -w `find . -type f -name '*.go' | grep -v vendor`

# linux builds a linux binary
linux:
	GOOS=linux GOARCH=amd64 $(GO) build -i -tags netgo $(GOFLAGS)

# install runs go install
install:
	$(GO) install $(GOFLAGS)

# vet runs go vet
vet:
	$(GO) vet ./...

# pkg builds a fabio.tar.gz package with only fabio in it
pkg: build test
	rm -rf pkg
	mkdir pkg
	tar czf pkg/fabio.tar.gz fabio

# release tags, builds and publishes a build with goreleaser
#
# Run this in sub-shells instead of dependencies so that
# later targets can pick up the new tag value.
release:
	$(MAKE) tag
	$(MAKE) preflight test gorelease homebrew docker-aliases

# preflight runs some checks before a release
preflight:
	[ "$(CUR_TAG)" == "$(LAST_TAG)" ] || ( echo "master not tagged. Last tag is $(LAST_TAG)" ; exit 1 )
	grep -q "$(LAST_TAG)" CHANGELOG.md main.go || ( echo "CHANGELOG.md or main.go not updated. $(LAST_TAG) not found"; exit 1 )

# tag tags the build
tag:
	build/tag.sh

# gorelease runs goreleaser to build and publish the artifacts
gorelease:
	[ -x "$(GORELEASER)" ] || ( echo "goreleaser not installed"; exit 1)
	GOVERSION=$(GOVERSION) goreleaser --rm-dist

# homebrew updates the brew recipe since goreleaser can only
# handle taps right now.
homebrew:
	build/homebrew.sh $(LAST_TAG)

# docker-aliases creates aliases for the docker containers
# since goreleaser doesn't handle that properly yet
docker-aliases:
	docker tag fabiolb/fabio:$(VERSION)-$(GOVERSION) magiconair/fabio:$(VERSION)-$(GOVERSION)
	docker tag fabiolb/fabio:$(VERSION)-$(GOVERSION) magiconair/fabio:latest
	docker push magiconair/fabio:$(VERSION)-$(GOVERSION)
	docker push magiconair/fabio:latest

# codeship runs the CI on codeship
codeship:
	go version
	go env
	wget -O ~/consul.zip https://releases.hashicorp.com/consul/1.0.5/consul_1.0.5_linux_amd64.zip
	wget -O ~/vault.zip https://releases.hashicorp.com/vault/0.9.3/vault_0.9.3_linux_amd64.zip
	unzip -o -d ~/bin ~/consul.zip
	unzip -o -d ~/bin ~/vault.zip
	vault --version
	consul --version
	cd ~/src/github.com/fabiolb/fabio && make test

# clean removes intermediate files
clean:
	$(GO) clean
	rm -rf pkg dist fabio
	find . -name '*.test' -delete

.PHONY: build clean docker gofmt homebrew install linux pkg release test vendorfmt vet
