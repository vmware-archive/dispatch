#!/bin/sh
set -e -x

cd $(dirname $0)

docker build -t vmware/dispatch-openfaas-nodejs6-base:0.0.3-dev1 .
