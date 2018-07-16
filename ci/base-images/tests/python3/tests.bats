#!/usr/bin/env bats

set -o pipefail

load ${DISPATCH_ROOT}/e2e/tests/helpers.bash

@test "Version" {
    run dispatch version
    echo_to_log
}

@test "Create python3 base image" {

    run dispatch create base-image python3-base-image ${image_url} --language python3
    echo_to_log
    assert_success

    run_with_retry "dispatch get base-image python3-base-image --json | jq -r .status" "READY" 8 5
}

@test "Create python3 image" {
    run dispatch create image python3-image python3-base-image
    echo_to_log
    assert_success

    run_with_retry "dispatch get image python3-image --json | jq -r .status" "READY" 8 5
}

@test "Create python function no schema" {
    run dispatch create function --image=python3-image python-hello-no-schema ${DISPATCH_ROOT}/examples/python3 --handler=hello.handle
    echo_to_log
    assert_success

    run_with_retry "dispatch get function python-hello-no-schema --json | jq -r .status" "READY" 10 5
}

@test "Execute python function no schema" {
    run_with_retry "dispatch exec python-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait --json | jq -r .output.myField" "Hello, Jon from Winterfell" 5 5
}

@test "Cleanup" {
    delete_entities function
    cleanup
}