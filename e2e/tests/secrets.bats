#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

@test "Batch load images" {
    batch_create_images
}

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

@test "Update secret" {
    run dispatch update --work-dir ${BATS_TEST_DIRNAME} -f secret_update.yaml
    assert_success

    run_with_retry "dispatch get secret open-sesame --json | jq -r .secrets.password" "OpenSesameStreet" 5 5
}

@test "Delete secrets" {
    delete_entities secret
}

@test "Cleanup" {
    cleanup
}
