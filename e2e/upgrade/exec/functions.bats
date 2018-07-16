#!/usr/bin/env bats

set -o pipefail

load ${E2E_TEST_ROOT}/helpers.bash
load ${E2E_TEST_ROOT}/variables.bash

@test "Execute node function no schema" {
    run_with_retry "dispatch exec node-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .output.myField" "Hello, Jon from Winterfell" 10 5
}

@test "Execute node function with schema" {
    run_with_retry "dispatch exec node-hello-with-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .output.myField" "Hello, Jon from Winterfell" 5 5
}

@test "Execute node function with input schema error" {
    run_with_retry "dispatch exec node-hello-with-schema --wait --json | jq -r .error.type" "InputError" 5 5
}

@test "Execute powershell with runtime deps" {
    run_with_retry "dispatch exec powershell-slack --wait --json | jq -r .output.result" "true" 5 5
}

@test "Execute python function with logging" {
    run_with_retry "dispatch exec logger --wait --json | jq -r \".logs.stderr | .[0]\"" "this goes to stderr" 5 5
    run_with_retry "dispatch exec logger --wait --json | jq -r \".logs.stdout | .[0]\"" "this goes to stdout" 5 5
}

@test "Update function python" {

    run dispatch update --work-dir ${E2E_TEST_ROOT} -f function_update.yaml
    assert_success
    run_with_retry "dispatch get function python-hello-no-schema --json | jq -r .status" "READY" 6 5

    run_with_retry "dispatch exec python-hello-no-schema --wait --json | jq -r .output.myField" "Goodbye, Noone from Nowhere" 6 5
}

@test "Delete python function no schema" {
    run dispatch delete function python-hello-no-schema
    echo_to_log
    assert_success

    run_with_retry "dispatch get runs python-hello-no-schema --json | jq '. | length'" 0 5 5
}
