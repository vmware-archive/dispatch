#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

@test "Batch load images" {
    batch_create_images
}

@test "Create subscription and emit an event" {
    func_name=node-echo-back-${RANDOM}
    sub_name=testsub-${RANDOM}
    event_name=test.event.${RANDOM}
    run dispatch create function --image=nodejs ${func_name} ${DISPATCH_ROOT}/examples/nodejs --handler=./debug.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function ${func_name} -o json | jq -r .status" "READY"

    run dispatch create subscription --name ${sub_name} --event-type ${event_name} ${func_name}
    echo_to_log
    assert_success

    run_with_retry "dispatch get subscription ${sub_name} -o json | jq -r .status" "READY"

    run dispatch emit ${event_name} --data='{"name": "Jon", "place": "Winterfell"}' --content-type="application/json"
    echo_to_log
    assert_success

    run_with_retry "dispatch get runs ${func_name} -o json | jq -r '.[0].functionName'" "${func_name}"
    run_with_retry "dispatch get runs ${func_name} -o json | jq -r '.[0].status'" "READY"
    result=$(dispatch get runs ${func_name} -o json | jq -r '.[0].output.context.event."eventType"')
    assert_equal "${event_name}" $result

    # Update subscription
    new_event_name=test.event.${RANDOM}
	update_tmp=$(mktemp)
	cat <<- EOF > ${update_tmp}
	eventType: ${new_event_name}
	function: ${func_name}
	kind: Subscription
	name: ${sub_name}
	secrets:
	tags:
	- label: update
	EOF
    assert_success

    run dispatch update -f ${update_tmp}
    assert_success

    run_with_retry "dispatch get subscription ${sub_name} -o json | jq -r .status" "READY"

    run dispatch emit ${new_event_name} --data='{"name": "Jon", "place": "Winterfell"}' --content-type="application/json"
    echo_to_log
    assert_success

    run_with_retry "dispatch get runs ${func_name} -o json | jq -r '.[0].functionName'" "${func_name}"
    run_with_retry "dispatch get runs ${func_name} -o json | jq -r '.[0].status'" "READY"
    run_with_retry "dispatch get runs ${func_name} -o json | jq 'sort_by(.finishedTime)' | jq -r '.[-1].output.context.event.eventType'" "${new_event_name}"
    rm ${update_tmp}

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

    run dispatch create function --image=nodejs ${func_name} ${DISPATCH_ROOT}/examples/nodejs --handler=./debug.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function ${func_name} -o json | jq -r .status" "READY"

    run dispatch create eventdrivertype ticker dispatchframework/ticker:0.0.2
    echo_to_log
    assert_success

    run dispatch get eventdrivertype ticker
    echo_to_log
    assert_success

    run dispatch create eventdriver ticker --name ${driver_name} --set seconds=2
    run_with_retry "dispatch get eventdriver ${driver_name} -o json | jq -r '.status'" "READY"

    run dispatch create subscription --event-type ticker.tick --name ${sub_name} ${func_name}
    echo_to_log
    assert_success

    run_with_retry "dispatch get subscription ${sub_name} -o json | jq -r .status" "READY"

    run_with_retry "dispatch get runs ${func_name} -o json | jq -r '.[0].status'" "READY"

    # Update event driver
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

    run_with_retry "dispatch get eventdriver ${driver_name} -o json | jq -r .status" "READY"
    result=$(dispatch get eventdriver ${driver_name} -o json | jq -r .config[0].value)
    assert_equal 5 $result

    rm ${update_tmp}

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

@test "Create event driver without available image" {
    driver_name=testdriver-${RANDOM}

    run dispatch create eventdrivertype baddrivertype unavailable-image:latest
    echo_to_log
    assert_success

    run dispatch get eventdrivertype baddrivertype
    echo_to_log
    assert_success

    run dispatch create eventdriver baddrivertype --name ${driver_name} --set seconds=2
    run_with_retry "dispatch get eventdriver ${driver_name} -o json | jq -r '.status'" "CREATING"

    run_with_retry "dispatch get eventdriver ${driver_name} -o json | jq -r '.status'" "ERROR"

    run dispatch delete eventdriver ${driver_name}
    echo_to_log
    assert_success

    run dispatch delete eventdrivertype baddrivertype
    echo_to_log
    assert_success
}

@test "Create eventdriver with invalid name (Upper case, underscores)" {

    bad_driver_name1=testDriver-${RANDOM}
    bad_driver_name2=test_driver-${RANDOM}

    run dispatch create eventdrivertype ticker kars7e/timer:latest
    echo_to_log
    assert_success

    run dispatch get eventdrivertype ticker
    echo_to_log
    assert_success

    run dispatch create eventdriver ticker --name ${bad_driver_name1} --set seconds=2
    echo_to_log
    assert_failure

    run dispatch create eventdriver ticker --name ${bad_driver_name2} --set seconds=2
    echo_to_log
    assert_failure

}

@test "Cleanup" {
    cleanup
}


