#!/bin/bash

set -e

gofmt -s -l .

go vet ./...
go fix ./...

golint .

ineffassign .
