#!/bin/bash -e
#
# docker.sh will build docker images from
# the versions provided on the command line.
# The binaries must already exist in the build/builds
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
# * fabiolb/fabio:1.1-go1.5.4
# * fabiolb/fabio:1.1-go1.6.2
# * fabiolb/fabio (which contains 1.1-go1.6.2)
#
if [[ $# = 0 ]]; then
	echo "Usage: docker.sh <1.x-go1.x.x> <1.x-go1.x.y>"
	exit 1
fi

for v in "$@" ; do
	echo "Building docker image fabiolb/fabio:$v"
	(
		cp dist/linuxamd64/fabio fabio
		docker build -q -t fabiolb/fabio:${v} .
	)
	docker tag fabiolb/fabio:$v magiconair/fabio:$v
	docker tag fabiolb/fabio:$v magiconair/fabio:latest
	docker tag fabiolb/fabio:$v fabiolb/fabio:latest
done

docker images | grep '/fabio' | egrep "($v|latest)"

read -p "Push docker images? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
	echo "Not pushing images. Exiting"
	exit 0
fi

echo "Pushing images..."
docker push fabiolb/fabio:$v
docker push fabiolb/fabio:latest
docker push magiconair/fabio:$v
docker push magiconair/fabio:latest
