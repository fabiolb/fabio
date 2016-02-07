#!/bin/bash -e
#
# Script for replacing the version number
# in main.go, committing and tagging the code

# use vendor path
export GO15VENDOREXPERIMENT=1

prgdir=$(cd $(dirname $0); pwd)
basedir=$(cd $prgdir/..; pwd)
v=$1

if [[ "$v" == "" ]]; then
	echo "Usage: $0 <version> (e.g. 1.0.4)"
	exit 1
fi

grep -q "$v" README.md || ( echo "README.md not updated"; exit 1 )
grep -q "$v" CHANGELOG.md || ( echo "CHANGELOG.md not updated"; exit 1 )

read -p "Release fabio version $v? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
	exit 1
fi

sed -i -e "s|^var version .*$|var version = \"$v\"|" $basedir/main.go
git add $basedir/main.go
git commit -m "Release v$v"
git commit --amend
git tag v$v

arch=amd64
for os in darwin linux ; do
	echo "Building release packages for $v $os"
	( cd $prgdir/.. ; GOOS=${os} GOARCH=${arch} go build -a -tags netgo -o /vagrant/fabio-${v}_${os}-${arch} )
done
