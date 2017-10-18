#!/bin/sh
set -e -x

docker build -t serverless-docker-local.artifactory.eng.vmware.com/photon-func-deps-node:7.7.4 .
