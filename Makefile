# CUR_TAG is the last git tag plus the delta from the current commit to the tag
# e.g. v1.5.5-<nr of commits since>-g<current git sha>
CUR_TAG ?= $(shell git describe --tags --first-parent)

# LAST_TAG is the last git tag
# e.g. v1.5.5
LAST_TAG ?= $(shell git describe --tags --first-parent --abbrev=0)

# VERSION is the last git tag without the 'v'
# e.g. 1.5.5
VERSION ?= $(shell git describe --tags --first-parent --abbrev=0 | cut -c 2-)

# GOFLAGS is the flags for the go compiler.
GOFLAGS ?= -mod=vendor -ldflags "-X main.version=$(CUR_TAG)"

# GOVERSION is the current go version, e.g. go1.9.2
GOVERSION ?= $(shell go version | awk '{print $$3;}')

# GORELEASER is the path to the goreleaser binary.
GORELEASER ?= $(shell which goreleaser)

# pin versions for CI builds
CI_CONSUL_VERSION ?= 1.8.4
CI_VAULT_VERSION ?= 1.5.2
CI_HUGO_VERSION ?= 0.63.2

# all is the default target
all: test

# help prints a help screen
help:
	@echo "build     - go build"
	@echo "install   - go install"
	@echo "test      - go test"
	@echo "gofmt     - go fmt"
	@echo "linux     - go build linux/amd64"
	@echo "release   - tag, build and publish release with goreleaser"
	@echo "pkg       - build, test and create pkg/fabio.tar.gz"
	@echo "clean     - remove temp files"

# build compiles fabio and the test dependencies
build: gofmt
	go build $(GOFLAGS)

# test builds and runs the tests
test: build
	go test $(GOFLAGS) -v -test.timeout 15s ./...

# mod performs go module maintenance
mod:
	go mod tidy
	go mod vendor

# gofmt runs gofmt on the code
gofmt:
	gofmt -s -w `find . -type f -name '*.go' | grep -v vendor`

# linux builds a linux binary
linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -tags netgo $(GOFLAGS)

# install runs go install
install:
	CGO_ENABLED=0 go install -trimpath $(GOFLAGS)

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
	$(MAKE) preflight docker-test gorelease homebrew

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

# docker-test runs make test in a Docker container with
# pinned versions of the external dependencies
#
# We download the binaries outside the Docker build to
# cache the binaries and prevent repeated downloads since
# ADD <url> downloads the file every time.
docker-test:
	docker build \
		--build-arg consul_version=$(CI_CONSUL_VERSION) \
		--build-arg vault_version=$(CI_VAULT_VERSION) \
		-t test-fabio \
		-f Dockerfile \
		.

# travis runs tests on Travis CI
travis:
	wget -q -O ~/consul.zip https://releases.hashicorp.com/consul/$(CI_CONSUL_VERSION)/consul_$(CI_CONSUL_VERSION)_linux_amd64.zip
	wget -q -O ~/vault.zip https://releases.hashicorp.com/vault/$(CI_VAULT_VERSION)/vault_$(CI_VAULT_VERSION)_linux_amd64.zip
	unzip -o -d ~/bin ~/consul.zip
	unzip -o -d ~/bin ~/vault.zip
	vault --version
	consul --version
	make test

# travis-pages runs the GitHub pages (https://fabiolb.net/) deploy on Travis CI
travis-pages:
	wget -q -O ~/hugo.tgz https://github.com/gohugoio/hugo/releases/download/v$(CI_HUGO_VERSION)/hugo_$(CI_HUGO_VERSION)_Linux-64bit.tar.gz
	tar -C ~/bin -zxf ~/hugo.tgz hugo
	hugo version
	(cd docs && hugo --verbose)

# github runs tests on github actions
github:
	wget -q -O ~/consul.zip https://releases.hashicorp.com/consul/$(CI_CONSUL_VERSION)/consul_$(CI_CONSUL_VERSION)_linux_amd64.zip
	wget -q -O ~/vault.zip https://releases.hashicorp.com/vault/$(CI_VAULT_VERSION)/vault_$(CI_VAULT_VERSION)_linux_amd64.zip
	unzip -o -d ~/bin ~/consul.zip
	unzip -o -d ~/bin ~/vault.zip
	vault --version
	consul --version
	make test

# clean removes intermediate files
clean:
	go clean
	rm -rf pkg dist fabio
	find . -name '*.test' -delete

.PHONY: all help build test mod gofmt linux install pkg release preflight tag gorelease homebrew docker-test travis travis-pages clean
