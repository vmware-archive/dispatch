#!/bin/sh
set -e -x

cd $(dirname $0)

docker build -t vmware/dispatch-openfaas-python-base:0.0.5-dev1 .
