#!/usr/bin/env bats

set -o pipefail

load ${E2E_TEST_ROOT}/helpers.bash
load ${E2E_TEST_ROOT}/variables.bash

@test "Create secret" {
    run dispatch create secret open-sesame ${DISPATCH_ROOT}/examples/nodejs/secret.json
    echo_to_log
    assert_success
}

@test "Create function with a default secret" {
    run dispatch create function --image=nodejs i-have-a-default-secret ${DISPATCH_ROOT}/examples/nodejs --handler=./i-have-a-secret.js --secret open-sesame
    echo_to_log
    assert_success

    run_with_retry "dispatch get function i-have-a-default-secret --json | jq -r .status" "READY" 15 5
}

@test "Create function without a default secret" {
    run dispatch create function --image=nodejs i-have-a-secret ${DISPATCH_ROOT}/examples/nodejs --handler=./i-have-a-secret.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function i-have-a-secret --json | jq -r .status" "READY" 8 5
}