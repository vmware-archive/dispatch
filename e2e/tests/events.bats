#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

@test "Batch load images" {
    batch_create_images images.yaml
}

@test "Create subscription and emit an event" {
    func_name=node-echo-back-${RANDOM}
    sub_name=testsub-${RANDOM}
    event_name=test.event.${RANDOM}
    run dispatch create function nodejs6 ${func_name} ${DISPATCH_ROOT}/examples/nodejs6/debug.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function ${func_name} --json | jq -r .status" "READY" 8 5

    # https://github.com/vmware/dispatch/issues/364
    sleep 5

    run dispatch create subscription --name ${sub_name} --event-type ${event_name} ${func_name}
    echo_to_log
    assert_success

    run_with_retry "dispatch get subscription ${sub_name} --json | jq -r .status" "READY" 4 5

    run dispatch emit ${event_name} --data='{"name": "Jon", "place": "Winterfell"}' --content-type="application/json"
    echo_to_log
    assert_success

    run_with_retry "dispatch get runs ${func_name} --json | jq -r '.[0].functionName'" "${func_name}" 4 5
    run_with_retry "dispatch get runs ${func_name} --json | jq -r '.[0].status'" "READY" 6 5
    result=$(dispatch get runs ${func_name} --json | jq -r '.[0].output.context.event."event-type"')
    assert_equal "${event_name}" $result

    run dispatch delete subscription ${sub_name}
    echo_to_log
    assert_success

    run dispatch delete function ${func_name}
    echo_to_log
    assert_success
}


@test "Create event driver and matching subscription" {
    func_name=node-echo-back-${RANDOM}
    sub_name=testsub-${RANDOM}
    driver_name=testdriver-${RANDOM}

    run dispatch create function nodejs6 ${func_name} ${DISPATCH_ROOT}/examples/nodejs6/debug.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function ${func_name} --json | jq -r .status" "READY" 8 5

    run dispatch create eventdrivertype ticker kars7e/timer:latest
    echo_to_log
    assert_success

    run dispatch get eventdrivertype ticker
    echo_to_log
    assert_success

    run dispatch create eventdriver ticker --name ${driver_name} --set seconds=2
    run_with_retry "dispatch get eventdriver ${driver_name} --json | jq -r '.status'" "READY" 4 5

    run dispatch create subscription --source-type ticker --event-type ticker.tick --name ${sub_name} ${func_name}
    echo_to_log
    assert_success

    run_with_retry "dispatch get subscription ${sub_name} --json | jq -r .status" "READY" 4 5

    run_with_retry "dispatch get runs ${func_name} --json | jq -r '.[0].status'" "READY" 4 5

    run dispatch delete eventdriver ${driver_name}
    echo_to_log
    assert_success

    run dispatch delete eventdrivertype ticker
    echo_to_log
    assert_success

    run dispatch delete subscription ${sub_name}
    echo_to_log
    assert_success

    run dispatch delete function ${func_name}
    echo_to_log
    assert_success
}


@test "Cleanup" {
    cleanup
}


