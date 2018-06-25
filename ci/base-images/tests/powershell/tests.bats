#!/usr/bin/env bats

set -o pipefail

load ${DISPATCH_ROOT}/e2e/tests/helpers.bash

@test "Version" {
    run dispatch version
    echo_to_log
}