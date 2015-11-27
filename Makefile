
GOFLAGS = -tags netgo -ldflags "-X main.version=$(shell git describe --tags)"

.PHONY: build
build:
	go build $(GOFLAGS)

.PHONY: install
install:
	go install $(GOFLAGS)

.PHONY: release
release:
	build/release.sh $(filter-out $@,$(MAKECMDGOALS))

.PHONY: docker
docker: build
	build/docker.sh
