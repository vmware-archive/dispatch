#!/bin/bash

: ${DOCKER_REGISTRY:="vmware"}

PACKAGE=${1}
BUILD=${2}
TAG=dev-${BUILD}

if [ -n "$CI" ]; then
    TAG=$IMAGE_TAG
fi

image=${DOCKER_REGISTRY}/vs-${PACKAGE}:${TAG}
echo $image

docker build -t $image -f images/${PACKAGE}/Dockerfile .
docker push $image
