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
    run dispatch create function --image=python3 i-have-a-default-secret ${DISPATCH_ROOT}/examples/python3 --handler=i-have-a-secret.handle --secret open-sesame
    echo_to_log
    assert_success

    run_with_retry "dispatch get function i-have-a-default-secret -o json | jq -r .status" "READY" 15 5
}

@test "Create function without a default secret" {
    run dispatch create function --image=python3 i-have-a-secret ${DISPATCH_ROOT}/examples/python3 --handler=i-have-a-secret.handle
    echo_to_log
    assert_success

    run_with_retry "dispatch get function i-have-a-secret -o json | jq -r .status" "READY" 8 5
}

@test "Execute function with a default secret" {
    run_with_retry "dispatch exec i-have-a-default-secret --wait -o json | jq -r .output.message" "The password is OpenSesame" 5 5
}

@test "Execute function without a default secret" {
    run_with_retry "dispatch exec i-have-a-secret --wait -o json | jq -r .output.message" "I know nothing" 5 5
}

@test "Validate invalid secret errors" {
    skip "secret validation only returns null"
}

@test "Update secret" {
    run dispatch update --work-dir ${BATS_TEST_DIRNAME} -f secret_update.yaml
    assert_success

    run_with_retry "dispatch get secret open-sesame -o json | jq -r .secrets.password" "OpenSesameStreet" 5 5
}

@test "Delete secrets" {
    delete_entities secret
}

@test "Cleanup" {
    cleanup
}
