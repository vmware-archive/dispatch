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

if [[ ${EXIT_STATUS} -ne 0 ]]; then
    cat ${BATS_LOG}
fi

exit ${EXIT_STATUS}
