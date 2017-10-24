#!/bin/sh
set -e -x

docker build -t serverless-docker-local.artifactory.eng.vmware.com/openfaas-nodejs-base:0.0.1-dev1 .
