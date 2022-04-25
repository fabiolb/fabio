.. image:: https://travis-ci.org/GehirnInc/crypt.svg?branch=master
    :target: https://travis-ci.org/GehirnInc/crypt

crypt - A password hashing library for Go
=========================================
crypt provides pure golang implementations of UNIX's crypt(3).

The goal of crypt is to bring a library of many common and popular password
hashing algorithms to Go and to provide a simple and consistent interface to
each of them. As every hashing method is implemented in pure Go, this library
should be as portable as Go itself.

All hashing methods come with a test suite which verifies their operation
against itself as well as the output of other password hashing implementations
to ensure compatibility with them.

I hope you find this library to be useful and easy to use!

Install
-------

To install crypt, use the *go get* command.

.. code-block:: sh

   go get github.com/GehirnInc/crypt


Usage
-----

.. code-block:: go

    package main

    import (
    	"fmt"

    	"github.com/GehirnInc/crypt"
    	_ "github.com/GehirnInc/crypt/sha256_crypt"
    )

    func main() {
    	crypt := crypt.SHA256.New()
    	ret, _ := crypt.Generate([]byte("secret"), []byte("$5$salt"))
    	fmt.Println(ret)

    	err := crypt.Verify(ret, []byte("secret"))
    	fmt.Println(err)

    	// Output:
    	// $5$salt$kpa26zwgX83BPSR8d7w93OIXbFt/d3UOTZaAu5vsTM6
    	// <nil>
    }

Documentation
-------------

The documentation is available on GoDoc_.

.. _GoDoc: https://godoc.org/github.com/GehirnInc/crypt
