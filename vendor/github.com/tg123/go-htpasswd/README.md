# htpasswd for Go

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
