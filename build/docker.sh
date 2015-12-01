#!/bin/bash -e
#
# Docker image build script

# set path
export PATH=/usr/local/go/bin:$PATH

# set gopath
export GOPATH=~/go

v=`git describe --tags`
v=${v/v/}
tag=magiconair/fabio

# check go version
if [[ `go version` != 'go version go1.5.1 linux/amd64' ]]; then
	echo "Invalid go version. Want go 1.5.1"
	exit 1
fi

echo "Building fabio $v with `go version`"
go clean
go build -tags netgo -ldflags "-X main.version=$v"

echo "Building docker image $tag:$v"
docker build -q -t $tag:$v .

echo "Building docker image $tag"
docker build -q -t $tag .

docker images
