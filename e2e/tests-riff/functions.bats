#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

@test "Batch load images" {
    batch_create_images images.yaml
}

@test "Create node function no schema" {

    run dispatch create function nodejs6 node-hello-no-schema ${DISPATCH_ROOT}/examples/nodejs6/hello.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-hello-no-schema --json | jq -r .status" "READY" 8 5
}

@test "Create a function with duplicate name" {
    run dispatch create function nodejs6 node-hello-dup ${DISPATCH_ROOT}/examples/nodejs6/hello.js
    echo_to_log
    assert_success

    run dispatch create function nodejs6 node-hello-dup ${DISPATCH_ROOT}/examples/nodejs6/hello.js
    assert_failure
}

@test "Execute node function no schema" {
    run_with_retry "dispatch exec node-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .output.myField" "Hello, Jon from Winterfell" 5 5
}

@test "Create python function no schema" {
    skip

    run dispatch create function python3 python-hello-no-schema ${DISPATCH_ROOT}/examples/python3/hello.py
    echo_to_log
    assert_success

    run_with_retry "dispatch get function python-hello-no-schema --json | jq -r .status" "READY" 6 5
}

@test "Execute python function no schema" {
    skip

    run_with_retry "dispatch exec python-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .output.myField" "Hello, Jon from Winterfell" 5 5
}

@test "Create powershell function no schema" {
    skip

    run dispatch create function powershell powershell-hello-no-schema ${DISPATCH_ROOT}/examples/powershell/hello.ps1
    echo_to_log
    assert_success

    run_with_retry "dispatch get function powershell-hello-no-schema --json | jq -r .status" "READY" 8 5
}

@test "Execute powershell function no schema" {
    skip

    run_with_retry "dispatch exec powershell-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .output.myField" "Hello, Jon from Winterfell" 5 5
}

@test "Create powershell function with runtime deps" {
    skip

    run dispatch create image powershell-with-slack powershell-base --runtime-deps ${DISPATCH_ROOT}/examples/powershell/requirements.psd1
    assert_success
    run_with_retry "dispatch get image powershell-with-slack --json | jq -r .status" "READY" 10 5

    run dispatch create function powershell-with-slack powershell-slack ${DISPATCH_ROOT}/examples/powershell/test-slack.ps1
    echo_to_log
    assert_success

    run_with_retry "dispatch get function powershell-slack --json | jq -r .status" "READY" 10 5
}

@test "Execute powershell with runtime deps" {
    skip

    run_with_retry "dispatch exec powershell-slack --wait --json | jq -r .output.result" "true" 5 5
}

@test "Create node function with schema" {
    run dispatch create function nodejs6 node-hello-with-schema ${DISPATCH_ROOT}/examples/nodejs6/hello.js --schema-in ${DISPATCH_ROOT}/examples/nodejs6/hello.schema.in.json --schema-out ${DISPATCH_ROOT}/examples/nodejs6/hello.schema.out.json
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-hello-with-schema --json | jq -r .status" "READY" 6 5
}

@test "Execute node function with schema" {
    run_with_retry "dispatch exec node-hello-with-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .output.myField" "Hello, Jon from Winterfell" 5 5
}

@test "Validate schema errors" {
    skip "schema validation only returns null"
}

@test "Create python function with runtime deps" {
    skip

    run dispatch create function python3 http ${DISPATCH_ROOT}/examples/python3/http.py
    echo_to_log
    assert_success

    run_with_retry "dispatch get function http --json | jq -r .status" "READY" 6 5
}

@test "Execute python function with runtime deps" {
    skip

    run_with_retry "dispatch exec http --wait --json | jq -r .output.status" "200" 5 5
}

@test "Delete functions" {
    delete_entities function
}

@test "Cleanup" {
    cleanup
}