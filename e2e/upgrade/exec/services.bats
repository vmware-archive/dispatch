#!/usr/bin/env bats

set -o pipefail

load ${E2E_TEST_ROOT}/helpers.bash
load ${E2E_TEST_ROOT}/variables.bash

@test "Execute a function which echos the service context" {

    run dispatch exec node-echo-service --wait
    echo_to_log
    assert_success

    run_with_retry "dispatch exec node-echo-service --wait --input '{\"in-1\": \"baz\"}' | jq -r '.output.context.serviceBindings.\"ups-with-schema\".\"special-key-1\"'" "special-value-1" 12 10
    echo_to_log
    assert_success
}

@test "Delete service instance" {

    run dispatch delete serviceinstance ups-with-schema
    echo_to_log
    assert_success
    # See issue https://github.com/vmware/dispatch/issues/542
    sleep 60
}

@test "[Re]Create service instance" {

    run dispatch create serviceinstance ups-with-schema user-provided-service-with-schemas default --params '{"param-1": "foo", "param-2": "bar"}'
    echo_to_log
    assert_success

    run_with_retry "dispatch get serviceinstances ups-with-schema --json | jq -r .status" "READY" 6 10
    run_with_retry "dispatch get serviceinstances ups-with-schema --json | jq -r .binding.status" "READY" 6 10
}

@test "[Re]Delete service instance" {

    run dispatch delete serviceinstance ups-with-schema
    echo_to_log
    assert_success
}

@test "Tear down catalog" {

    run helm delete --purge ups-broker
    echo_to_log
    assert_success

    run helm delete --purge catalog
    echo_to_log
    assert_success

    run kubectl delete namespace ${DISPATCH_NAMESPACE}-services
    echo_to_log
    assert_success
}