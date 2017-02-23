#!/bin/bash -e

readonly prgdir=$(cd $(dirname $0); pwd)
readonly basedir=$(cd $prgdir/..; pwd)
v=$1

[[ -n "$v" ]] || read -p "Enter version (e.g. 1.0.4): " v
if [[ -z "$v" ]] ; then
	echo "Usage: $0 [<version>] (e.g. 1.0.4)"
	exit 1
fi

go get -u github.com/mitchellh/gox
for go in go1.8; do
	echo "Building fabio with ${go}"
	gox -gocmd ~/${go}/bin/go -tags netgo -output "${basedir}/build/builds/fabio-${v}/fabio-${v}-${go}-{{.OS}}_{{.Arch}}"
done

( cd ${basedir}/build/builds/fabio-${v} && shasum -a 256 fabio-${v}-* > fabio-${v}.sha256 )
( cd ${basedir}/build/builds/fabio-${v} && gpg2 --output fabio-${v}.sha256.sig --detach-sig fabio-${v}.sha256 )
