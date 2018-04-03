#!/bin/sh
set -e -x

cd $(dirname $0)

docker build -t vmware/dispatch-nodejs6-base:0.0.2-dev1 .
