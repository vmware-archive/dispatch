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

    run_with_retry "dispatch get function node-hello-no-schema --json | jq -r .status" "READY" 6 5
}

@test "Execute node function no schema" {
    run_with_retry "dispatch exec node-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .output.myField" "Hello, Jon from Winterfell" 0 0
}

@test "Create python function no schema" {
    run dispatch create function python3 python-hello-no-schema ${DISPATCH_ROOT}/examples/python3/hello.py
    echo_to_log
    assert_success

    run_with_retry "dispatch get function python-hello-no-schema --json | jq -r .status" "READY" 6 5
}

@test "Execute python function no schema" {
    run_with_retry "dispatch exec python-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .output.myField" "Hello, Jon from Winterfell" 0 0
}

@test "Create node function with schema" {
    run dispatch create function nodejs6 node-hello-with-schema ${DISPATCH_ROOT}/examples/nodejs6/hello.js --schema-in ${DISPATCH_ROOT}/examples/nodejs6/hello.schema.in.json --schema-out ${DISPATCH_ROOT}/examples/nodejs6/hello.schema.out.json
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-hello-with-schema --json | jq -r .status" "READY" 6 5
}

@test "Execute node function with schema" {
    run_with_retry "dispatch exec node-hello-with-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .output.myField" "Hello, Jon from Winterfell" 0 0
}

@test "Validate schema errors" {
    skip "schema validation only returns null"
}

@test "Delete functions" {
    delete_functions
}

@test "Cleanup" {
    cleanup
}