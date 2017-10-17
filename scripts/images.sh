#!/bin/bash

: ${REPO:="serverless-docker-local.artifactory.eng.vmware.com"}
: ${VALUES_PATH:="values.yaml"}

PACKAGE=${1}
BUILD=${2}
TAG=dev-${BUILD}
if [ ! -z $CI ]; then
    TAG=ci-${CI_JOB_ID}
fi

image=${REPO}/${PACKAGE}:${TAG}
echo $image

docker build -t $image -f images/${PACKAGE}/Dockerfile .
docker push $image