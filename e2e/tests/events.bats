#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

@test "Batch load images" {
    batch_create_images images.yaml
}

@test "Create node function with schema" {
    run dispatch create function nodejs6 node-hello-subscribe-${RUN_ID} ${DISPATCH_ROOT}/examples/nodejs6/hello.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-hello-subscribe-${RUN_ID} --json | jq -r .status" "READY" 6 5
    sleep 5 # https://github.com/vmware/dispatch/issues/67
}

@test "Create subscription" {
    run dispatch create subscription test.event node-hello-subscribe-${RUN_ID}
    echo_to_log
    assert_success

    run_with_retry "dispatch get subscription test_event_node-hello-subscribe-${RUN_ID} --json | jq -r .status" "READY" 4 5
}

@test "Emit event" {
    run dispatch emit test.event --payload='{"name": "Jon", "place": "Winterfell"}'
    echo_to_log
    assert_success

    run_with_retry "dispatch get runs node-hello-subscribe-${RUN_ID} --json | jq -r .[0].output.myField" "Hello, Jon from Winterfell" 4 5
}

@test "Delete subscription" {
    run dispatch delete subscription test_event_node-hello-subscribe-${RUN_ID}
    echo_to_log
    assert_success
}

@test "Cleanup" {
    cleanup
}


