#!/bin/sh
set -e -x

cd $(dirname $0)

docker build -t vmware/dispatch-powershell-base:0.0.1-dev1 .
