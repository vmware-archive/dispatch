#!/usr/bin/env bats

set -o pipefail

load ${E2E_TEST_ROOT}/helpers.bash
load ${E2E_TEST_ROOT}/variables.bash

@test "Execute function with a default secret" {
    run_with_retry "dispatch exec i-have-a-default-secret --wait --json | jq -r .output.message" "The password is OpenSesame" 5 5
}

@test "Execute function without a default secret" {
    run_with_retry "dispatch exec i-have-a-secret --wait --json | jq -r .output.message" "I know nothing" 5 5
    run_with_retry "dispatch exec i-have-a-secret --secret open-sesame --wait --json | jq -r .output.message" "The password is OpenSesame" 5 5
}

@test "Update secret" {
    run dispatch update --work-dir ${E2E_TEST_ROOT} -f secret_update.yaml
    assert_success

    run_with_retry "dispatch get secret open-sesame --json | jq -r .secrets.password" "OpenSesameStreet" 5 5
}