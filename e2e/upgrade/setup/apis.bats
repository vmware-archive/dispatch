#!/usr/bin/env bats

set -o pipefail

load ${E2E_TEST_ROOT}/helpers.bash
load ${E2E_TEST_ROOT}/variables.bash

@test "Create functions for APIs" {
    run dispatch create function --image=nodejs func-nodejs ${DISPATCH_ROOT}/examples/nodejs --handler=./hello.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function func-nodejs --json | jq -r .status" "READY" 10 5

    run dispatch create function --image=nodejs node-echo-back ${DISPATCH_ROOT}/examples/nodejs --handler=./debug.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-echo-back --json | jq -r .status" "READY" 10 5
}

@test "Create APIs with HTTP(S)" {
    run dispatch create api api-test-http func-nodejs -m POST -p /http --auth public
    echo_to_log
    assert_success

    run_with_retry "dispatch get api api-test-http --json | jq -r .status" "READY" 10 5
}

@test "Create APIs with HTTPS ONLY" {
    run dispatch create api api-test-https-only func-nodejs -m POST --https-only -p /https-only --auth public
    echo_to_log
    assert_success
    run_with_retry "dispatch get api api-test-https-only --json | jq -r .status" "READY" 6 5
}

@test "Create APIs with Kong Plugins" {
    run dispatch create api api-test func-nodejs -m GET -m DELETE -m POST -m PUT -p /hello --auth public
    echo_to_log
    assert_success
    run_with_retry "dispatch get api api-test --json | jq -r .status" "READY" 6 5

    run dispatch create api api-echo node-echo-back -m GET -m DELETE -m POST -m PUT -p /echo --auth public
    echo_to_log
    assert_success
    run_with_retry "dispatch get api api-echo --json | jq -r .status" "READY" 6 5
}

@test "Create APIs with CORS" {
    run dispatch create api api-test-cors func-nodejs -m POST -m PUT -p /cors --auth public --cors
    echo_to_log
    assert_success
    run_with_retry "dispatch get api api-test-cors --json | jq -r .status" "READY" 10 5
}

@test "Create API Updates" {
    run dispatch create api api-test-update func-nodejs -m GET -p /hello --auth public
    assert_success
    run_with_retry "dispatch get api api-test-update --json | jq -r .status" "READY" 6 5

    run_with_retry "curl -s -X GET ${API_GATEWAY_HTTP_HOST}/${DISPATCH_ORGANIZATION}/hello -k | jq -r .myField" "Hello, Noone from Nowhere" 6 5
}

