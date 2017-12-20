#!/usr/bin/bash

set -x -e -u

pushd dispatch
./scripts/test-check.sh
popd
