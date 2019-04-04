# htpasswd for Go

[![Build Status](https://farmer1992.visualstudio.com/opensources/_apis/build/status/go-htpasswd?branchName=master)](https://farmer1992.visualstudio.com/opensources/_build/latest?definitionId=17&branchName=master)
[![Go Doc](https://godoc.org/github.com/tg123/go-htpasswd?status.svg)](https://godoc.org/github.com/tg123/go-htpasswd)
[![Go Report Card](https://goreportcard.com/badge/github.com/tg123/go-htpasswd)](https://goreportcard.com/report/github.com/tg123/go-htpasswd)


This is a libary to validate user credentials against an HTTPasswd file. 

This was forked from <https://github.com/jimstudt/http-authentication/tree/master/basic> 
with modifications by @brian-avery to support SSHA, Md5Crypt, and Bcrypt.

## Currently, this supports:
* SSHA
* MD5Crypt
* APR1Crypt
* SHA
* Bcrypt
* Plain text

## Not supported:
* Crypt
