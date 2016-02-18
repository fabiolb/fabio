#!/bin/bash -e

prgdir=$(cd $(dirname $0); pwd)
basedir=$(cd $prgdir/..; pwd)
v=$1

arch=amd64
for os in darwin linux ; do
	for go in go1.5.3 go1.6; do
		f=fabio-${v}-${go}_${os}-${arch}
		echo "Building $f"
		( cd $basedir ; GO15VENDOREXPERIMENT=1 GOOS=${os} GOARCH=${arch} ~/$go/bin/go build -a -tags netgo -o $f )
	done
done

