#!/bin/bash -e
#
# Script for replacing the version number
# in main.go, committing and tagging the code

readonly prgdir=$(cd $(dirname $0); pwd)
readonly basedir=$(cd $prgdir/..; pwd)
v=$1

[[ -n "$v" ]] || read -p "Enter version (e.g. 1.0.4): " v
if [[ -z "$v" ]]; then
	echo "Usage: $0 <version> (e.g. 1.0.4)"
	exit 1
fi

grep -q "$v" CHANGELOG.md || echo "CHANGELOG.md not updated"

read -p "Tag fabio version $v? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
	exit 1
fi

sed -i '' -e "s|^var version .*$|var version = \"$v\"|" $basedir/main.go
git add $basedir/main.go
git commit -S -m "Release v$v"
git tag -s v$v -m "Tag v${v}"
