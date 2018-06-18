#!/bin/bash

: ${DOCKER_REGISTRY:="vmware"}

PACKAGE=${1}

if [ -n "$CI" ]; then
    TAG=$IMAGE_TAG
fi

image=${DOCKER_REGISTRY}/dispatch-${PACKAGE}:${TAG}
echo $image

docker build -t $image -f images/${PACKAGE}/Dockerfile .
if [ -n "$PUSH_IMAGES" ]; then
    docker push $image
fi
