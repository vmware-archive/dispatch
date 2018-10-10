#!/bin/bash

set -x -e

: ${DISPATCH_SERVER_DOCKER_REPOSITORY:="vmware"}
: ${PREFIX:=.}

PACKAGE=${1}

if [ -n "$CI" ]; then
    TAG=$IMAGE_TAG
fi

if [[ ${PACKAGE} == dispatch-* ]]; then
    image=${DISPATCH_SERVER_DOCKER_REPOSITORY}/${PACKAGE}:${TAG}
else
    image=${DISPATCH_SERVER_DOCKER_REPOSITORY}/dispatch-${PACKAGE}:${TAG}
fi
echo $image

docker build -t $image -f ${PREFIX}/images/${PACKAGE}/Dockerfile .
if [ -n "$PUSH_IMAGES" ]; then
    docker push $image
fi
