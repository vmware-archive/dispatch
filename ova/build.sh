#!/bin/bash
# Copyright 2017-2018 VMware, Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# this file is responsible for triggering the build in a consisent way,
# regardless of whether it's being run in a development environment or CI
DEBUG=${DEBUG:-}
set -eu -o pipefail +h && [ -n "$DEBUG" ] && set -x
ROOT_DIR="$GOPATH/src/github.com/vmware/dispatch/"
ROOT_WORK_DIR="/go/src/github.com/vmware/dispatch/"

ROOT_INSTALLER_DIR="${ROOT_DIR}/ova"
ROOT_INSTALLER_WORK_DIR="${ROOT_WORK_DIR}/ova"

ROOT_UI_DIR="$GOPATH/src/github.com/vmware/dispatch-ui-binaries"
ROOT_WORK_UI_DIR="/go/src/github.com/vmware/dispatch-ui-binaries"

CI_IMAGE="dispatchframework/ova-builder:latest"

TAG=${TAG:-$(git describe --tags --dirty)}

function usage() {
    echo -e "Usage:
      <ova-dev|ova-ci>
      [passthrough args for ./bootable/build-main.sh]
    e.g.: $0 ova-dev
    e.g.: $0 ova-ci" >&2
    exit 1
}

[ $# -gt 0 ] || usage
step=$1; shift
[ ! "$step" == "ova-ci" ] || [ ! "$step" == "ova-dev" ] || usage

echo "--------------------------------------------------"
if [ "$step" == "ova-dev" ]; then
  echo "starting docker dev build container..."
  docker run -it --rm --privileged -v /dev:/dev \
    -v "${ROOT_DIR}/:/${ROOT_WORK_DIR}/":ro \
    -v "${ROOT_INSTALLER_DIR}/bin/:/${ROOT_INSTALLER_WORK_DIR}/bin/" \
    -v "${ROOT_UI_DIR}":"${ROOT_WORK_UI_DIR}" \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -e DEBUG="${DEBUG}" \
    -e TAG="${TAG}" \
    -e TERM -w "${ROOT_INSTALLER_WORK_DIR}" \
    "${CI_IMAGE}" ./bootable/build-main.sh -m ./ova-manifest.json -r "${ROOT_INSTALLER_WORK_DIR}/bin" "$@"
elif [ "$step" == "ova-ci" ]; then
  echo "starting ci build..."
  export DEBUG="${DEBUG}"
  export TAG="${TAG}"
  export CI_MODE="yes"
  bootable/build-main.sh -m ./ova-manifest.json -r "$(pwd)/bin" "$@"
else
  usage
fi
