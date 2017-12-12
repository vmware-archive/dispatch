#!/bin/sh
set -e -x

docker build -t vmware/vs-openfaas-python3-base:0.0.4-dev1 .
