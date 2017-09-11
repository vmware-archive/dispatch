#!/bin/bash

set -x

go get -u github.com/go-swagger/go-swagger/cmd/swagger
go get -u github.com/jteeuwen/go-bindata/...
go get -u github.com/alecthomas/gometalinter

# install all the linters
gometalinter --install --update