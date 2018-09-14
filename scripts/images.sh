#!/bin/bash

: ${DOCKER_REPOSITORY:="vmware"}

PACKAGE=${1}

if [ -n "$CI" ]; then
    TAG=$IMAGE_TAG
fi

if [[ ${PACKAGE} == dispatch-* ]]; then
    image=${DOCKER_REPOSITORY}/${PACKAGE}:${TAG}
else
    image=${DOCKER_REPOSITORY}/dispatch-${PACKAGE}:${TAG}
fi
echo $image

docker build -t $image -f images/${PACKAGE}/Dockerfile .
if [ -n "$PUSH_IMAGES" ]; then
    docker push $image
fi
