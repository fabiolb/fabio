
# do not specify a full path for go since travis will fail
GO = GOGC=off go
GOFLAGS = -ldflags "-X main.version=$(shell git describe --tags)"
GOVENDOR = $(shell which govendor)
VENDORFMT = $(shell which vendorfmt)

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

build: checkdeps vendorfmt
	$(GO) build -i $(GOFLAGS)
	$(GO) test -i ./...

test: checkdeps vendorfmt
	$(GO) test -v -test.timeout 15s `go list ./... | grep -v '/vendor/'`

checkdeps:
	[ -x "$(GOVENDOR)" ] || $(GO) get -u github.com/kardianos/govendor
	govendor list +e | grep '^ e ' && { echo "Found missing packages. Please run 'govendor add +e'"; exit 1; } || : echo

vendorfmt:
	[ -x "$(VENDORFMT)" ] || $(GO) get -u github.com/magiconair/vendorfmt/cmd/vendorfmt
	vendorfmt

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

codeship:
	go version
	go env
	wget -O ~/consul.zip https://releases.hashicorp.com/consul/1.0.0/consul_1.0.0_linux_amd64.zip
	wget -O ~/vault.zip https://releases.hashicorp.com/vault/0.8.3/vault_0.8.3_linux_amd64.zip
	unzip -o -d ~/bin ~/consul.zip
	unzip -o -d ~/bin ~/vault.zip
	vault --version
	consul --version
	cd ~/src/github.com/fabiolb/fabio && make test

fabio-builder: make-fabio-builder push-fabio-builder

make-fabio-builder:
	docker build -t fabiolb/fabio-builder --squash build/fabio-builder

push-fabio-builder:
	docker push fabiolb/fabio-builder

buildpkg: test
	build/build.sh

clean:
	$(GO) clean
	rm -rf pkg

.PHONY: build buildpkg clean docker gofmt homebrew install linux pkg release test vendorfmt
