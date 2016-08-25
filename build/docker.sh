#!/bin/bash -e
#
# docker.sh will build docker images from
# the versions provided on the command line.
# The binaries must already exist in the /vagrant
# directory and are usually built with the build.sh
# or the release.sh script. The last specified
# version will be used as the 'latest' image.
#
# Example:
#
#   build/docker.sh 1.1-go1.5.4 1.1-go1.6
#
# will build three containers
#
# * magiconair/fabio:1.1-go1.5.4
# * magiconair/fabio:1.1-go1.6.2
# * magiconair/fabio (which contains 1.1-go1.6.2)
#
tag=magiconair/fabio

if [[ $# = 0 ]]; then
	echo "Usage: docker.sh <version> <version>"
	exit 1
fi

for v in "$@" ; do
	echo "Building docker image $tag:$v"
	( cp build/builds/fabio-${v}_linux-amd64 fabio ; docker build -q -t ${tag}:${v} . )
done

echo "Building docker image $tag"
docker build -q -t $tag .

docker images
