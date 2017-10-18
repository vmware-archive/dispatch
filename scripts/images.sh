#!/bin/bash

: ${DOCKER_REGISTRY:="serverless-docker-local.artifactory.eng.vmware.com"}

PACKAGE=${1}
BUILD=${2}
TAG=dev-${BUILD}

if [ -n "$CI" ]; then
    TAG=$IMAGE_TAG
fi

image=${DOCKER_REGISTRY}/${PACKAGE}:${TAG}
echo $image

docker build -t $image -f images/${PACKAGE}/Dockerfile .
docker push $image
