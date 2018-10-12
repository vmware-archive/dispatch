#!/usr/bin/env bats

set -o pipefail

load ${E2E_TEST_ROOT}/helpers.bash
load ${E2E_TEST_ROOT}/variables.bash

@test "Create node function no schema" {

    run dispatch create function --image=nodejs node-hello-no-schema ${DISPATCH_ROOT}/examples/nodejs --handler=./hello.js
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-hello-no-schema --json | jq -r .status" "READY" 8 5
}

@test "Create node function with schema" {
    run dispatch create function --image=nodejs node-hello-with-schema ${DISPATCH_ROOT}/examples/nodejs --handler=./hello.js --schema-in ${DISPATCH_ROOT}/examples/nodejs/hello.schema.in.json --schema-out ${DISPATCH_ROOT}/examples/nodejs/hello.schema.out.json
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-hello-with-schema --json | jq -r .status" "READY" 6 5
}

@test "Create powershell function with runtime deps" {
    run dispatch create image powershell-with-slack powershell-base --runtime-deps ${DISPATCH_ROOT}/examples/powershell/requirements.psd1
    assert_success
    run_with_retry "dispatch get image powershell-with-slack --json | jq -r .status" "READY" 10 5

    run dispatch create function --image=powershell-with-slack powershell-slack ${DISPATCH_ROOT}/examples/powershell --handler=test-slack.ps1::handle
    echo_to_log
    assert_success

    run_with_retry "dispatch get function powershell-slack --json | jq -r .status" "READY" 20 5
}

@test "Create python function with logging" {
    src_dir=$(mktemp -d)
    cat << EOF > ${src_dir}/logging_test.py
import sys

def handle(ctx, payload):
    print("this goes to stdout")
    print("this goes to stderr", file=sys.stderr)
EOF

    run dispatch create function --image=python3 logger ${src_dir} --handler=logging_test.handle
    echo_to_log
    assert_success

    run_with_retry "dispatch get function logger --json | jq -r .status" "READY" 10 5
}

@test "Create python function no schema" {
    run dispatch create function --image=python3 python-hello-no-schema ${DISPATCH_ROOT}/examples/python3 --handler=hello.handle
    echo_to_log
    assert_success

    run_with_retry "dispatch get function python-hello-no-schema --json | jq -r .status" "READY" 10 5
}