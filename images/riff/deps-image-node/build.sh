#!/bin/sh
set -e -x

cd $(dirname $0)

docker build -t vmware/dispatch-riff-nodejs6-base:0.0.3-dev1 .
