#!/bin/bash -e

readonly prgdir=$(cd $(dirname $0); pwd)
readonly basedir=$(cd $prgdir/..; pwd)
v=$1

[[ -n "$v" ]] || read -p "Enter version (e.g. 1.0.4): " v
if [[ -z "$v" ]] ; then
	echo "Usage: $0 [<version>] (e.g. 1.0.4)"
	exit 1
fi

arch=amd64
for os in darwin linux ; do
	for go in go1.6.3 go1.7.1; do
		f=build/builds/fabio-${v}-${go}_${os}-${arch}
		echo "Building $f"
		( cd $basedir ; GOOS=${os} GOARCH=${arch} ~/$go/bin/go build -a -tags netgo -o $f )
	done
done

( cd build/builds && shasum -a 256 fabio-${v}-* > fabio-${v}.sha256 )
