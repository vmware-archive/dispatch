#!/bin/bash
set -e -o pipefail

: ${WORKDIR:="/root/go/src/github.com/vmware/dispatch"}
: ${ROOT_PACKAGE:="github.com/vmware/dispatch"}
: ${CUSTOM_RESOURCE_VERSION:="v1"}
: ${CI_IMAGE:="vmware/dispatch-golang-ci:1.10-20180930"}

CUSTOM_RESOURCE_NAME=${1}

GENERATE_COMMAND="src/k8s.io/code-generator/generate-groups.sh all \
    $ROOT_PACKAGE/pkg/resources/gen \
    $ROOT_PACKAGE/pkg/resources \
    $CUSTOM_RESOURCE_NAME:$CUSTOM_RESOURCE_VERSION"

if [[ -z ${CI} ]]; then
    docker run --rm -v `pwd`:${WORKDIR} ${CI_IMAGE} bash -c "/root/go/${GENERATE_COMMAND}"
else
    echo "CI is set to ${CI}"
    bash -c "/root/go/${GENERATE_COMMAND}"
fi