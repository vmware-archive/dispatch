#!/usr/bin/bash

set -x -e -u

pushd dispatch
./scripts/gofmtcheck.sh
popd
