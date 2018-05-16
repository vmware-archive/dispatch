#!/bin/bash
set -e -o pipefail

# Generates server & client code based on swagger spec. Uses existing model's structures.
# the spec references the model.json which is generated from code using generate-models.sh
# For server, spec must be flattened first, otherwise the embedded spec still contains
# references to external file and does not validate properly when services are started.

: ${WORKDIR:="/root/go/src/github.com/vmware/dispatch"}
: ${CI_IMAGE:="vmware/dispatch-golang-ci:1.10-20180512"}
: ${QUIET:="-q"}
: ${MODELS_PACKAGE:="github.com/vmware/dispatch/pkg/api/v1"}

echo Using image ${CI_IMAGE}

PACKAGE=${1}
APP=${2}
SWAGGER=${3}

SERVER_COMMAND="pushd ${WORKDIR} && \
        swagger generate server ${QUIET} -A ${APP} -t ./pkg/${PACKAGE}/gen -f ./swagger/${SWAGGER} --existing-models=${MODELS_PACKAGE} --model-package=v1 --skip-models --exclude-main && \
        popd"

CLIENT_COMMAND="pushd ${WORKDIR} && \
        swagger generate client ${QUIET} -A ${APP} -t ./pkg/${PACKAGE}/gen -f ./swagger/${SWAGGER} --existing-models=${MODELS_PACKAGE} --model-package=v1 --skip-models && \
        popd"

mkdir -p ./pkg/${PACKAGE}/gen
if [[ -z ${CI} ]]; then
    docker run --rm -v `pwd`:${WORKDIR} ${CI_IMAGE} bash -c "${SERVER_COMMAND}"
    docker run --rm -v `pwd`:${WORKDIR} ${CI_IMAGE} bash -c "${CLIENT_COMMAND}"
else
    echo "CI is set to ${CI}"
    bash -c "${SERVER_COMMAND}"
    bash -c "${CLIENT_COMMAND}"
fi

