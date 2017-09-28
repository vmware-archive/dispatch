#!/bin/bash

set -x

go get -u github.com/jteeuwen/go-bindata/...
go get -u github.com/alecthomas/gometalinter
go get -u github.com/vektra/mockery/.../
go get -u github.com/imikushin/prune

# install all the linters
gometalinter --install --update
