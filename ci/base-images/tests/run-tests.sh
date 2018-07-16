#!/usr/bin/env bash
# Usage: ./run-tests.sh [tests bat file]

set -e

mkdir $(pwd)/test-logs
export BATS_LOG=$(pwd)/test-logs/bats.log

export DISPATCH_ROOT=$(pwd)/dispatch

function run_bats() {
    echo "==> running tests:"
    bats $1
    if [[ $? -ne 0 ]]; then
        EXIT_STATUS=1
    fi
    echo "tests finished <=="
}


EXIT_STATUS=0
run_bats "$1"

# TODO: delete
cat ${BATS_LOG}

exit ${EXIT_STATUS}
