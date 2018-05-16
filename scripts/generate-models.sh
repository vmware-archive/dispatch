#!/bin/bash
set -e -o pipefail

: ${WORKDIR:="/root/go/src/github.com/vmware/dispatch"}
: ${CI_IMAGE:="vmware/dispatch-golang-ci:1.10-20180512"}
: ${QUIET:="-q"}
: ${PACKAGE:="github.com/vmware/dispatch/pkg/api/v1"}

echo Using image ${CI_IMAGE}

TARGET=${1}

GENERATE_COMMAND="pushd ${WORKDIR} && \
         CGO_ENABLED=0 swagger generate spec ${QUIET} -o ${TARGET} -b ${PACKAGE} -m && \
         popd"

if [[ -z ${CI} ]]; then
docker run --rm -v `pwd`:${WORKDIR} ${CI_IMAGE} bash -c "${GENERATE_COMMAND}"
else
    echo "CI is set to ${CI}"
    bash -c "${GENERATE_COMMAND}"
fi