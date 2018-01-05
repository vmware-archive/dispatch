#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

@test "Batch load images" {
    batch_create_images images.yaml
}

@test "Create secret" {
    run dispatch create secret open-sesame ${DISPATCH_ROOT}/examples/nodejs6/secret.json
    echo_to_log
    assert_success
}

@test "Create function with a default secret" {
    run dispatch create function nodejs6 i-have-a-default-secret ${DISPATCH_ROOT}/examples/nodejs6/i-have-a-secret.js --secret open-sesame
    echo_to_log
    assert_success

    run_with_retry "dispatch get function i-have-a-default-secret --json | jq -r .status" "READY" 8 5
}

@test "Create function without a default secret" {
    run dispatch create function nodejs6 i-have-a-secret ${DISPATCH_ROOT}/examples/nodejs6/i-have-a-secret.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function i-have-a-secret --json | jq -r .status" "READY" 8 5
}

@test "Execute function with a default secret" {
    run_with_retry "dispatch exec i-have-a-default-secret --wait --json | jq -r .output.message" "The password is OpenSesame" 5 5
}

@test "Execute function without a default secret" {
    run_with_retry "dispatch exec i-have-a-secret --wait --json | jq -r .output.message" "I know nothing" 5 5
    run_with_retry "dispatch exec i-have-a-secret --secret open-sesame --wait --json | jq -r .output.message" "The password is OpenSesame" 5 5
}

@test "Validate invalid secret errors" {
    skip "secret validation only returns null"
}

@test "Delete secrets" {
    delete_secrets
}

@test "Cleanup" {
    cleanup
}