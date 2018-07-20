#!/usr/bin/env bats

set -o pipefail

load ${DISPATCH_ROOT}/e2e/tests/helpers.bash

@test "Version" {
    run dispatch version
    echo_to_log
}

@test "Create java base image" {

    run dispatch create base-image java-base-image ${image_url} --language java
    echo_to_log
    assert_success

    run_with_retry "dispatch get base-image java-base-image -o json | jq -r .status" "READY" 8 5
}

@test "Create java image" {
    run dispatch create image java-image java-base-image
    echo_to_log
    assert_success

    run_with_retry "dispatch get image java-image -o json | jq -r .status" "READY" 8 5
}

@test "Create java function no schema" {
    run dispatch create function --image=java-image java-hello-no-schema ${DISPATCH_ROOT}/examples/java/hello-with-deps --handler=io.dispatchframework.examples.Hello
    echo_to_log
    assert_success

    run_with_retry "dispatch get function java-hello-no-schema -o json | jq -r .status" "READY" 20 5
}

@test "Execute java function no schema" {
    run_with_retry "dispatch exec java-hello-no-schema --input='{\"name\": \"Jon\", \"place\": \"Winterfell\"}' --wait -o json | jq -r .output" "Hello, Jon from Winterfell" 5 5
}

@test "Cleanup" {
    delete_entities function
    cleanup
}