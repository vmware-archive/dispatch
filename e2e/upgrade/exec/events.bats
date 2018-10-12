#!/usr/bin/env bats

set -o pipefail

load ${E2E_TEST_ROOT}/helpers.bash
load ${E2E_TEST_ROOT}/variables.bash

@test "Emit an event for subscription" {
    func_name=node-echo-back-sub
    sub_name=testsub
    event_name=test.event
    
    run dispatch emit ${event_name} --data='{"name": "Jon", "place": "Winterfell"}' --content-type="application/json"
    echo_to_log
    assert_success

    run_with_retry "dispatch get runs ${func_name} --json | jq -r '.[0].functionName'" "${func_name}" 4 5
    run_with_retry "dispatch get runs ${func_name} --json | jq -r '.[0].status'" "READY" 12 5
    result=$(dispatch get runs ${func_name} --json | jq -r '.[0].output.context.event."eventType"')
    assert_equal "${event_name}" $result
}

@test "Create subscription with event driver" {
    func_name=node-echo-back-driver
    sub_name=driversub
    driver_name=testdriver

    run dispatch create subscription --event-type ticker.tick --name ${sub_name} ${func_name}
    echo_to_log
    assert_success

    run_with_retry "dispatch get subscription ${sub_name} --json | jq -r .status" "READY" 4 5

    run_with_retry "dispatch get runs ${func_name} --json | jq -r '.[0].status'" "READY" 4 5
}

@test "Update event driver" {
    driver_name=testdriver

	update_tmp=$(mktemp)
	cat <<- EOF > ${update_tmp}
	config:
	- key: seconds
	  value: '5'
	kind: Driver
	name: ${driver_name}
	secrets:
	tags:
	type: ticker
	EOF
    assert_success

    run dispatch update -f ${update_tmp}
    echo_to_log
    assert_success

    run_with_retry "dispatch get eventdriver ${driver_name} --json | jq -r .status" "READY" 4 5
    result=$(dispatch get eventdriver ${driver_name} --json | jq -r .config[0].value)
    assert_equal 5 $result

    rm ${update_tmp}
}