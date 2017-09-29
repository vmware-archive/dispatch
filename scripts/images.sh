#!/bin/bash

: ${REPO:="serverless-docker-local.artifactory.eng.vmware.com"}

PACKAGE=${1}
BUILD=${2}
TAG=dev-${BUILD}
if [ ! -z $CI ]; then
    TAG=ci-${CI_JOB_ID}
fi

echo ${REPO}/${PACKAGE}:${TAG}

docker build -t ${REPO}/${PACKAGE}:${TAG} -f images/${PACKAGE}/Dockerfile .
docker push ${REPO}/${PACKAGE}:${TAG}