#!/bin/sh
set -e -x

cd $(dirname $0)

docker build -t vmware/dispatch-powercli-base:0.0.1 .
