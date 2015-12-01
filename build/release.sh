#!/bin/bash -e
#
# Script for replacing the version number
# in main.go, committing and tagging the code

prgdir=$(cd $(dirname $0); pwd)
basedir=$(cd $prgdir/..; pwd)
v=$1

if [[ "$v" == "" ]]; then
	echo "Usage: $0 <version>"
	exit 1
fi

read -p "Release fabio version $v? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
	exit 1
fi

sed -i '' -e "s|^var version .*$|var version = \"$v\"|" $basedir/main.go
git add $basedir/main.go
git commit -m "Release v$v"
git commit --amend
git tag v$v

arch=amd64
for os in darwin linux ; do
	echo "Building release packages for $v $os"
	( cd $prgdir/.. ; GOOS=${os} GOARCH=${arch} go build -a -tags netgo -o build/fabio-${v}_${os}-${arch} )
done
