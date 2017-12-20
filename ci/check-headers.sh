#!/usr/bin/bash

set -x -e -u

pushd dispatch
./scripts/header-check.sh
popd
