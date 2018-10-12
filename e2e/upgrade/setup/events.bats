#!/usr/bin/env bats

set -o pipefail

load ${E2E_TEST_ROOT}/helpers.bash
load ${E2E_TEST_ROOT}/variables.bash

@test "Create subscription" {
    func_name=node-echo-back-sub
    sub_name=testsub
    event_name=test.event

    run dispatch create function --image=nodejs ${func_name} ${DISPATCH_ROOT}/examples/nodejs --handler=./debug.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function ${func_name} --json | jq -r .status" "READY" 8 5

    # https://github.com/vmware/dispatch/issues/364
    sleep 5

    run dispatch create subscription --name ${sub_name} --event-type ${event_name} ${func_name}
    echo_to_log
    assert_success

    run_with_retry "dispatch get subscription ${sub_name} --json | jq -r .status" "READY" 4 5
}

@test "Create event driver" {
    func_name=node-echo-back-driver
    driver_name=testdriver

    run dispatch create function --image=nodejs ${func_name} ${DISPATCH_ROOT}/examples/nodejs --handler=./debug.js
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
}