#!/bin/bash
#
# homebrew.sh creates an updated homebrew on fabio

set -o nounset
set -o errexit
set -o pipefail

readonly prgdir=$(cd $(dirname $0); pwd)
readonly brewdir=$(brew --prefix)/Homebrew/Library/Taps/homebrew/homebrew-core

v=${1:-}
[[ -n "$v" ]] || read -p "Enter version (e.g. 1.0.4): " v
if [[ -z "$v" ]] ; then
	echo "Usage: $0 <version> (e.g. 1.0.4)"
	exit 1
fi
v=${v/v/}

srcurl=https://github.com/fabiolb/fabio/archive/v${v}.tar.gz
shasum=$(wget -O- -q "$srcurl" | shasum -a 256 | awk '{ print $1; }')
echo -e "/urlDAurl \"$srcurl\"/sha256DAsha256 \"$shasum\":wq" > $prgdir/homebrew.vim

brew update
brew update
(
	cd $brewdir
	git checkout -b fabio-$v origin/master
	vim -u NONE -s $prgdir/homebrew.vim $brewdir/Formula/fabio.rb
	brew install --build-from-source fabio
	brew test fabio
	brew install fabio
	brew audit --strict fabio
	git add Formula/fabio.rb
	git commit -m "fabio $v"
	git push --set-upstream magiconair fabio-$v
)

echo "Goto https://github.com/magiconair/homebrew-core to create pull request"
open https://github.com/magiconair/homebrew-core

exit 0
