#!/bin/sh
set -e -x

docker build -t vmware/vs-openfaas-nodejs6-base:0.0.2-dev1 .
