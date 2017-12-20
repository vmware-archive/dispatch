#!/bin/bash

set -e -u -x

export SRC_PATH=$GOPATH/src/github.com/vmware/
export REPO_PATH=$PWD/dispatch


pushd dispatch
    git checkout -b pr
popd

mkdir -p ${SRC_PATH}

git clone ${REPO_PATH} ${SRC_PATH}/dispatch

pushd ${SRC_PATH}/dispatch
    git checkout pr
    mkdir -p .cover
    ./scripts/coverage.sh
popd

mkdir -p coverage
mv ${SRC_PATH}/dispatch/.cover/cover.html coverage/

