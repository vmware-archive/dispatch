#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

@test "Batch load images" {
    batch_create_images images.yaml
}

@test "Create node function with schema" {
    run dispatch create function nodejs6 node-echo-back ${DISPATCH_ROOT}/examples/nodejs6/debug.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-echo-back --json | jq -r .status" "READY" 8 5
}

@test "Create subscription" {
    run dispatch create subscription --name testsubscription --event-type test.event node-echo-back
    echo_to_log
    assert_success

    run_with_retry "dispatch get subscription testsubscription --json | jq -r .status" "READY" 4 5
}

@test "Emit event" {
    sleep 5 # Still seeing the issue https://github.com/vmware/dispatch/issues/67
    run dispatch emit test.event --data='{"name": "Jon", "place": "Winterfell"}' --content-type="application/json"
    echo_to_log
    assert_success

    run_with_retry "dispatch get runs node-echo-back --json | jq -r '.[0].functionName'" "node-echo-back" 4 5
    run_with_retry "dispatch get runs node-echo-back --json | jq -r '.[0].status'" "READY" 6 5
    result=$(dispatch get runs node-echo-back --json | jq -r '.[0].output.context.event."event-type"')
    assert_equal "test.event" $result
}

@test "Delete subscription" {
    run dispatch delete subscription testsubscription
    echo_to_log
    assert_success
}

@test "Register new event driver type" {
    run dispatch create eventdrivertype ticker kars7e/timer:latest
    echo_to_log
    assert_success

    run dispatch get eventdrivertype ticker
    echo_to_log
    assert_success
}

@test "Create event driver" {
    run dispatch create eventdriver ticker --name my-ticker --set seconds=2
    run_with_retry "dispatch get eventdriver my-ticker --json | jq -r '.status'" "READY" 4 5
}

@test "Create event driver subscription" {
    initial_runs=$(dispatch get runs node-echo-back --json | jq -r '. | length')

    run dispatch create subscription --source-type ticker --event-type ticker.tick --name tickersub node-echo-back
    echo_to_log
    assert_success

    run_with_retry "dispatch get subscription tickersub --json | jq -r .status" "READY" 4 5

    expected_runs=$((initial_runs+1))
    run_with_retry_check_ret "min_value_met ${expected_runs} $(dispatch get runs node-echo-back --json | jq '. | length')" 4 3
}

@test "Cleanup" {
    cleanup
}


