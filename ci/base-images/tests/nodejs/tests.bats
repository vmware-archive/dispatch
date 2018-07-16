#!/usr/bin/env bats

set -o pipefail

load ${DISPATCH_ROOT}/e2e/tests/helpers.bash

@test "Create nodejs base image" {

    run dispatch create base-image nodejs-base-image ${image_url} --language nodejs
    echo_to_log
    assert_success

    run_with_retry "dispatch get base-image nodejs-base-image --json | jq -r .status" "READY" 8 5
}

@test "Create nodejs image" {
    run dispatch create image nodejs-image nodejs-base-image
    echo_to_log
    assert_success

    run_with_retry "dispatch get image nodejs-image --json | jq -r .status" "READY" 8 5
}

@test "Create nodejs function no schema" {
    run dispatch create function --image=nodejs-image nodejs-hello-no-schema ${DISPATCH_ROOT}/examples/nodejs --handler=./hello.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function nodejs-hello-no-schema --json | jq -r .status" "READY" 10 5
}

@test "Execute node function no schema" {
    run_with_retry "dispatch exec node-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .output.myField" "Hello, Jon from Winterfell" 10 5
}

@test "Delete node function no schema" {
    run dispatch delete function node-hello-no-schema
    echo_to_log
    assert_success

    run_with_retry "dispatch get runs node-hello-no-schema --json | jq '. | length'" 0 5 5
}

@test "Create node function with schema" {
    run dispatch create function --image=nodejs node-hello-with-schema ${DISPATCH_ROOT}/examples/nodejs --handler=./hello.js --schema-in ${DISPATCH_ROOT}/examples/nodejs/hello.schema.in.json --schema-out ${DISPATCH_ROOT}/examples/nodejs/hello.schema.out.json
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-hello-with-schema --json | jq -r .status" "READY" 6 5
}

@test "Execute node function with schema" {
    run_with_retry "dispatch exec node-hello-with-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .output.myField" "Hello, Jon from Winterfell" 5 5
}

@test "Execute node function with input schema error" {
    run_with_retry "dispatch exec node-hello-with-schema --wait --json | jq -r .error.type" "InputError" 5 5
}

@test "Cleanup" {
    delete_entities function
    cleanup
}