---
title: "Development"
weight: 200
---

For newcomers to Go, you can't just `git clone` your forked repo and work from
there, due to how Go's `GOPATH` works. You can follow the steps below to get started:

1. Fork this repository to your own account (named `myfork` below)
1. Make sure you have [Consul](https://www.consul.io/downloads.html) and [Vault](https://www.vaultproject.io/downloads.html) installed in your `$PATH`
1. `go get github.com/fabiolb/fabio`, change to the directory where the code was cloned
   (`$GOPATH/src/github.com/fabiolb/fabio`) and add your fork as remote: `git remote add myfork git@github.com:myfork/fabio.git`
1. Hack away!
1. `go fmt` and `make test` your code
1. Commit your changes and *push to your own fork*: `git push myfork`
1. Create a pull-request
