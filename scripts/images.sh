#!/bin/bash

: ${DOCKER_REGISTRY:="vmware"}

PACKAGE=${1}

if [ -n "$CI" ]; then
    TAG=$IMAGE_TAG
fi

if [[ ${PACKAGE} == dispatch-* ]]; then
    image=${DOCKER_REGISTRY}/${PACKAGE}:${TAG}
else
    image=${DOCKER_REGISTRY}/dispatch-${PACKAGE}:${TAG}
fi
echo $image

docker build -t $image -f images/${PACKAGE}/Dockerfile .
if [ -n "$PUSH_IMAGES" ]; then
    docker push $image
fi
