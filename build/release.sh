#!/bin/bash -e
#
# Script for replacing the version number
# in main.go, committing and tagging the code

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

$prgdir/build.sh $v
